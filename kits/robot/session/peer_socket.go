package session

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	kitlog "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/types"
	"github.com/pion/webrtc/v3"
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
	onReceive          atomic.Value
	onWrite            atomic.Value
	onClose            atomic.Value
	onTimeout          atomic.Value
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
	dc                 *DataChannel
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
	p.onReceive.Store(fn)
}

// OnWrite 处理当发到数据时
func (p *PeerSocket) OnWrite(fn PeerOnWriteCallback) {
	p.onWrite.Store(fn)
}

// OnClose 处理关闭
func (p *PeerSocket) OnClose(fn PeerOnCloseCallback) {
	p.onClose.Store(fn)
}

// OnTimeout 处理关闭
func (p *PeerSocket) OnTimeout(fn func(Peer) error) {
	p.onTimeout.Store(fn)
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
		if stoped := p.stoped.Get(); stoped {
			return
		}

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
			if m, ok := err.(*net.OpError); ok && m.Timeout() {
				if handler, ok := p.onTimeout.Load().(func(Peer) error); ok && handler != nil {
					if err := handler(p); err == nil {
						continue
					}
				}
			}
			return
		}

		size := binary.BigEndian.Uint16(header)
		payload := make([]byte, size)
		n, err = io.ReadFull(p.conn, payload)
		if err != nil {
			kitlog.Error(Logger()).Log("error", "read payload failed", "reason", err.Error(), "size", n)
			return
		}

		if n < 1 {
			continue
		}

		p.mutexRW.RLock()
		mws := newDCChain(p.receiveMiddlewares)
		p.mutexRW.RUnlock()
		m := mws.Process(PeerMessageProcessFunc(func(args PeerMessageProcessArgs) {
			if handler, ok := p.onReceive.Load().(PeerOnReceiveCallback); ok && handler != nil {
				if err := handler(args.Peer, args.Payload); err != nil {
					kitlog.Error(Logger()).Log("error", "onReceiveCallback failed", "reason", err.Error())
					args.Peer.Close()
				}
			}
		}))

		m.Process(PeerMessageProcessArgs{Peer: p, Payload: PeerMessagePayload{Data: payload}})

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
		dc := p.dc
		p.mutexRW.RUnlock()

		if dc != nil {
			dc.Close()
		}

		if handler, ok := p.onClose.Load().(PeerOnCloseCallback); ok && handler != nil {
			handler(p)
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
			mws := newDCChain(p.writeMiddlewares)
			p.mutexRW.RUnlock()

			m := mws.Process(PeerMessageProcessFunc(func(args PeerMessageProcessArgs) {
				payload := args.Payload
				if handler, ok := p.onWrite.Load().(PeerOnWriteCallback); ok && handler != nil {
					payload = handler(args.Payload)
				}

				if payload.Channel == PeerMessageChannelWebrtc {
					if err := p.writeToDataChannel(payload.Data); err != nil {
						kitlog.Error(Logger()).Log("action", "write", "error", err)
						return
					}
				} else {
					if n, err := p.write(payload.Data); err != nil {
						kitlog.Error(Logger()).Log("action", "write", "error", err, "size", n)
						return
					}
					p.LoadOrResetSeqNo(1)
				}
			}))

			data := make([]byte, len(frame.Data))
			copy(data[0:1], frame.Data[0:1])
			binary.BigEndian.PutUint32(data[1:], p.LoadOrResetSeqNo()+1)
			copy(data[5:], frame.Data[5:])
			frame.Data = data
			m.Process(PeerMessageProcessArgs{Peer: p, Payload: frame})

		case <-p.stopChan:
			return

		case <-p.notifyDieChan:
			return
		}
	}
}

func (p *PeerSocket) write(frame []byte) (int, error) {
	size := len(frame)
	binary.BigEndian.PutUint16(p.cacheBytes, uint16(size))
	copy(p.cacheBytes[2:], frame)
	return p.conn.Write(p.cacheBytes[:size+2])
}

func (p *PeerSocket) writeToDataChannel(frame []byte) error {
	size := len(frame)
	binary.BigEndian.PutUint16(p.cacheBytes, uint16(size))
	copy(p.cacheBytes[2:], frame)
	if p.dc == nil {
		return errors.New("datachnnel is nil")
	}
	return p.dc.Send(p.cacheBytes[:size+2])
}

// DataChannel 获取数据通道
func (p *PeerSocket) DataChannel() *DataChannel {
	p.mutexRW.RLock()
	defer p.mutexRW.RUnlock()
	return p.dc
}

func (p *PeerSocket) CreateDataChannel(configuration webrtc.Configuration) (*DataChannel, error) {
	pc, err := webrtc.NewPeerConnection(configuration)
	if err != nil {
		return nil, err
	}

	mdc := &DataChannel{
		pc:         pc,
		dcs:        make(map[string]*webrtc.DataChannel),
		candidates: make([]webrtc.ICECandidateInit, 0),
	}

	dc, err := mdc.CreateDataChannel(defaultDataChannelLabel, &webrtc.DataChannelInit{})
	if err != nil {
		return nil, err
	}

	dc.OnClose(func() {
		mdc.RemoveDataChannel(dc.Label())
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		p.mutexRW.RLock()
		mws := newDCChain(p.receiveMiddlewares)
		p.mutexRW.RUnlock()
		m := mws.Process(PeerMessageProcessFunc(func(args PeerMessageProcessArgs) {
			if handler, ok := p.onReceive.Load().(PeerOnReceiveCallback); ok && handler != nil {
				if err := handler(args.Peer, args.Payload); err != nil {
					kitlog.Error(Logger()).Log("error", "onReceiveCallback failed", "reason", err.Error())
				}
			}
		}))
		m.Process(PeerMessageProcessArgs{Peer: p, Payload: PeerMessagePayload{Channel: PeerMessageChannelWebrtc, Data: msg.Data}})
	})

	p.mutexRW.Lock()
	p.dc = mdc
	p.mutexRW.Unlock()
	return mdc, nil
}

// Go start
func (p *PeerSocket) Go() {
	p.readyedChan = make(chan struct{})
	go p.receiver()
	go p.writer()

	for i := 0; i < 2; i++ {
		<-p.readyedChan
	}

	close(p.readyedChan)
}

// NewPeerSocket 创建
func NewPeerSocket(id string, conn net.Conn, readDeadline, writeDeadline time.Duration, notifyChan chan struct{}) *PeerSocket {
	if id == "" {
		id = uuid.NewV4().String()
	}

	peer := &PeerSocket{
		id:                 id,
		conn:               conn,
		readDeadline:       readDeadline,
		writeDeadline:      writeDeadline,
		cacheBytes:         make([]byte, 65535),
		receiveMiddlewares: make(PeerMessageMiddlewares, 0),
		writeMiddlewares:   make(PeerMessageMiddlewares, 0),
		writeChan:          make(chan PeerMessagePayload, 1024),
		stopChan:           make(chan struct{}),
		notifyDieChan:      notifyChan,
	}

	// init params
	peer.params.Store(make(map[string]interface{}))
	return peer
}
