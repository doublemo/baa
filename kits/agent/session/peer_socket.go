package session

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	kitlog "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/types"
	uuid "github.com/satori/go.uuid"
)

// PeerSocket 处理来至scoket的连接
type PeerSocket struct {
	id                 string
	conn               net.Conn
	seqNo              uint32
	readDeadline       time.Duration
	writeDeadline      time.Duration
	cacheBytes         []byte
	onReceive          PeerOnReceiveCallback
	onWrite            PeerOnWriteCallback
	onClose            PeerOnCloseCallback
	receiveMiddlewares PeerMessageMiddlewares
	writeMiddlewares   PeerMessageMiddlewares
	writeChan          chan PeerMessagePayload
	readyedChan        chan struct{}
	stopChan           chan struct{}
	notifyDieChan      chan struct{}
	stoped             types.AtomicBool
	params             atomic.Value
	mutex              sync.Mutex
	mutexRW            sync.RWMutex
	closeOnce          sync.Once
}

// ID 返回Peer ID
func (p *PeerSocket) ID() string {
	return p.id
}

// Use 处理中间件
func (p *PeerSocket) Use(middlewares ...PeerMessageMiddleware) {
	p.mutexRW.Lock()
	for _, m := range middlewares {
		r := m.Receive()
		w := m.Write()
		if r != nil {
			p.receiveMiddlewares = append(p.receiveMiddlewares, r)
		}

		if w != nil {
			p.writeMiddlewares = append(p.writeMiddlewares, w)
		}
	}

	p.mutexRW.Unlock()
}

// OnReceive 处理当收到数据时
func (p *PeerSocket) OnReceive(fn PeerOnReceiveCallback) {
	p.mutexRW.Lock()
	p.onReceive = fn
	p.mutexRW.Unlock()
}

// OnWrite 处理当发到数据时
func (p *PeerSocket) OnWrite(fn PeerOnWriteCallback) {
	p.mutexRW.Lock()
	p.onWrite = fn
	p.mutexRW.Unlock()
}

// OnClose 处理关闭
func (p *PeerSocket) OnClose(fn PeerOnCloseCallback) {
	p.mutexRW.Lock()
	p.onClose = fn
	p.mutexRW.Unlock()
}

// Send 发送数据
func (p *PeerSocket) Send(frame PeerMessagePayload) error {
	if stoped := p.stoped.Get(); stoped {
		return ErrPeerWriteMessageFailed
	}

	select {
	case p.writeChan <- frame:
	default:
		return ErrPeerChannelIsFulled
	}

	return nil
}

// Close 关闭
func (p *PeerSocket) Close() error {
	p.closeOnce.Do(func() {
		p.stoped.Set(true)
		p.conn.Close()
		close(p.stopChan)
	})
	return nil
}

// LoadOrResetSeqNo 信息ID
func (p *PeerSocket) LoadOrResetSeqNo(v ...uint32) uint32 {
	if len(v) > 0 {
		return atomic.AddUint32(&p.seqNo, v[0])
	}

	return atomic.LoadUint32(&p.seqNo)
}

// Params 信息
func (p *PeerSocket) Params(key string) (interface{}, bool) {
	m := p.params.Load().(map[string]interface{})
	v, ok := m[key]
	return v, ok
}

// SetParams 设置
func (p *PeerSocket) SetParams(key string, value interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	m1 := p.params.Load().(map[string]interface{})
	m2 := make(map[string]interface{})
	for k, v := range m1 {
		m2[k] = v
	}

	m2[key] = value
	p.params.Store(m2)
}

func (p *PeerSocket) receiver() {
	defer func() {
		if stoped := p.stoped.Get(); !stoped {
			close(p.stopChan)
		}
	}()

	p.readyedChan <- struct{}{}
	header := make([]byte, 2)

	for {
		p.conn.SetReadDeadline(time.Now().Add(p.readDeadline))
		n, err := io.ReadFull(p.conn, header)
		if err != nil {
			return
		}

		size := binary.BigEndian.Uint16(header)
		payload := make([]byte, size)
		n, err = io.ReadFull(p.conn, payload)
		if err != nil {
			kitlog.Error(Logger()).Log("error", "read payload failed", "reason", err.Error(), "size", n)
			return
		}

		p.LoadOrResetSeqNo(1)
		p.mutexRW.RLock()
		onReceiveCallback := p.onReceive
		mws := newDCChain(p.receiveMiddlewares)
		p.mutexRW.RUnlock()

		m := mws.Process(PeerMessageProcessFunc(func(args PeerMessageProcessArgs) {
			if onReceiveCallback != nil {
				if err := onReceiveCallback(args.Peer, args.Payload); err != nil {
					kitlog.Error(Logger()).Log("error", "onReceiveCallback failed", "reason", err.Error())
					args.Peer.Close()
				}
			}
		}))

		m.Process(PeerMessageProcessArgs{Peer: p, Payload: PeerMessagePayload{Type: 0, Data: payload}})

		select {
		case <-p.stopChan:
			return

		case <-p.notifyDieChan:
			return

		default:
		}
	}
}

func (p *PeerSocket) writer() {
	defer func() {
		p.stoped.Set(true)

		p.mutexRW.RLock()
		onCloseCallback := p.onClose
		p.mutexRW.RUnlock()

		if onCloseCallback != nil {
			onCloseCallback(p)
		}
	}()

	p.readyedChan <- struct{}{}
	for {
		select {
		case frame, ok := <-p.writeChan:
			if !ok {
				return
			}

			if p.writeDeadline.Nanoseconds() > 0 {
				p.conn.SetWriteDeadline(time.Now().Add(p.writeDeadline))
			}

			p.mutexRW.RLock()
			onWriteCallback := p.onWrite
			mws := newDCChain(p.writeMiddlewares)
			p.mutexRW.RUnlock()

			m := mws.Process(PeerMessageProcessFunc(func(args PeerMessageProcessArgs) {
				payload := args.Payload
				if onWriteCallback != nil {
					payload = onWriteCallback(args.Payload)
				}

				if n, err := p.write(payload.Type, payload.Data); err != nil {
					kitlog.Error(Logger()).Log("action", "write", "error", err, "size", n)
					return
				}
			}))

			m.Process(PeerMessageProcessArgs{Peer: p, Payload: frame})

		case <-p.stopChan:
			return

		case <-p.notifyDieChan:
			return
		}
	}
}

func (p *PeerSocket) write(frametype int, frame []byte) (int, error) {
	size := len(frame)
	binary.BigEndian.PutUint16(p.cacheBytes, uint16(size))
	copy(p.cacheBytes[2:], frame)
	return p.conn.Write(p.cacheBytes[:size+2])
}

// NewPeerSocket 创建
func NewPeerSocket(conn net.Conn, readDeadline, writeDeadline time.Duration, notifyChan chan struct{}) *PeerSocket {
	peer := &PeerSocket{
		id:                 uuid.NewV4().String(),
		conn:               conn,
		readDeadline:       readDeadline,
		writeDeadline:      writeDeadline,
		cacheBytes:         make([]byte, 65535),
		receiveMiddlewares: make(PeerMessageMiddlewares, 0),
		writeMiddlewares:   make(PeerMessageMiddlewares, 0),
		writeChan:          make(chan PeerMessagePayload, 1024),
		readyedChan:        make(chan struct{}),
		stopChan:           make(chan struct{}),
		notifyDieChan:      notifyChan,
	}

	// init params
	peer.params.Store(make(map[string]interface{}))

	go peer.receiver()
	go peer.writer()

	for i := 0; i < 2; i++ {
		<-peer.readyedChan
	}

	close(peer.readyedChan)
	return peer
}