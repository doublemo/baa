package session

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	kitlog "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/types"
	awebrtc "github.com/doublemo/baa/kits/agent/webrtc"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	uuid "github.com/satori/go.uuid"
)

// PeerWebsocket 处理来至webscoket的连接
type PeerWebsocket struct {
	id                 string
	conn               *websocket.Conn
	seqNo              uint32
	readDeadline       time.Duration
	writeDeadline      time.Duration
	maxMessageSize     int64
	messageType        int
	onReceive          atomic.Value
	onWrite            atomic.Value
	onClose            atomic.Value
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
func (p *PeerWebsocket) ID() string {
	return p.id
}

// Use 处理中间件
func (p *PeerWebsocket) Use(middlewares ...PeerMessageMiddleware) {
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
func (p *PeerWebsocket) OnReceive(fn PeerOnReceiveCallback) {
	p.onReceive.Store(fn)
}

// OnWrite 处理当发到数据时
func (p *PeerWebsocket) OnWrite(fn PeerOnWriteCallback) {
	p.onWrite.Store(fn)
}

// OnClose 处理关闭
func (p *PeerWebsocket) OnClose(fn PeerOnCloseCallback) {
	p.onClose.Store(fn)
}

// Send 发送数据
func (p *PeerWebsocket) Send(frame PeerMessagePayload) error {
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
func (p *PeerWebsocket) Close() error {
	p.closeOnce.Do(func() {
		p.stoped.Set(true)
		p.conn.Close()
		close(p.stopChan)
	})
	return nil
}

// LoadOrResetSeqNo 信息ID
func (p *PeerWebsocket) LoadOrResetSeqNo(v ...uint32) uint32 {
	if len(v) > 0 {
		return atomic.AddUint32(&p.seqNo, v[0])
	}

	return atomic.LoadUint32(&p.seqNo)
}

// Params 信息
func (p *PeerWebsocket) Params(key string) (interface{}, bool) {
	m := p.params.Load().(map[string]interface{})
	v, ok := m[key]
	return v, ok
}

// SetParams 设置
func (p *PeerWebsocket) SetParams(key string, value interface{}) {
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

// MessageType 设置和获取数类型
func (p *PeerWebsocket) MessageType(t ...int) int {
	if len(t) > 0 {
		if t[0] != websocket.BinaryMessage && t[0] != websocket.TextMessage {
			return p.messageType
		}

		p.mutexRW.Lock()
		p.messageType = t[0]
		p.mutexRW.Unlock()
		return t[0]
	}

	p.mutexRW.RLock()
	defer p.mutexRW.RUnlock()
	return p.messageType
}

func (p *PeerWebsocket) receiver() {
	defer func() {
		if stoped := p.stoped.Get(); !stoped {
			close(p.stopChan)
		}
	}()

	p.readyedChan <- struct{}{}
	for {
		p.conn.SetReadLimit(p.maxMessageSize)
		p.conn.SetReadDeadline(time.Now().Add(p.readDeadline))
		p.conn.SetPongHandler(func(string) error {
			p.conn.SetReadDeadline(time.Now().Add(p.readDeadline))
			return nil
		})

		_, payload, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseProtocolError) {
				kitlog.Error(Logger()).Log("error", err)
			}
			return
		}

		p.LoadOrResetSeqNo(1)
		p.mutexRW.RLock()
		mws := newDCChain(p.receiveMiddlewares)
		p.mutexRW.RUnlock()

		m := mws.Process(PeerMessageProcessFunc(func(args PeerMessageProcessArgs) {
			if handler, ok := p.onReceive.Load().(PeerOnReceiveCallback); ok && handler != nil {
				if err := handler(args.Peer, args.Payload); err != nil {
					kitlog.Error(Logger()).Log("error", "onReceiveCallback failed", "reason", err.Error(), "peer_id", args.Peer.ID())
					args.Peer.Close()
					return
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

func (p *PeerWebsocket) writer() {
	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
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
					if err := p.write(p.MessageType(), payload.Data); err != nil {
						kitlog.Error(Logger()).Log("action", "weite", "error", err)
						return
					}
				}

			}))

			m.Process(PeerMessageProcessArgs{Peer: p, Payload: frame})

		case <-ticker.C:
			if p.writeDeadline.Nanoseconds() > 0 {
				p.conn.SetWriteDeadline(time.Now().Add(p.writeDeadline))
			}

			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseProtocolError) {
					kitlog.Error(Logger()).Log("error", err)
				}
				return
			}

		case <-p.stopChan:
			return

		case <-p.notifyDieChan:
			return
		}
	}
}

func (p *PeerWebsocket) write(frametype int, frame []byte) error {
	w, err := p.conn.NextWriter(frametype)
	if err != nil {
		return err
	}

	w.Write(frame)
	return w.Close()
}

func (p *PeerWebsocket) writeToDataChannel(frame []byte) error {
	if p.dc == nil {
		return errors.New("datachnnel is nil")
	}
	return p.dc.Send(frame)
}

// DataChannel 获取数据通道
func (p *PeerWebsocket) DataChannel() *DataChannel {
	p.mutexRW.RLock()
	defer p.mutexRW.RUnlock()
	return p.dc
}

// CreateDataChannel 数据通道
func (p *PeerWebsocket) CreateDataChannel(w awebrtc.WebRTCTransportConfig) (*DataChannel, error) {
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&webrtc.MediaEngine{}), webrtc.WithSettingEngine(w.Setting))
	pc, err := api.NewPeerConnection(w.Configuration)
	if err != nil {
		return nil, err
	}

	mdc := &DataChannel{
		pc:         pc,
		dcs:        make([]*webrtc.DataChannel, 0),
		candidates: make([]webrtc.ICECandidateInit, 0),
	}

	pc.OnDataChannel(func(peer *PeerWebsocket, m *DataChannel) func(*webrtc.DataChannel) {
		return func(dc *webrtc.DataChannel) {
			m.AddDataChannel(dc)
			dc.OnClose(func() { m.RemoveDataChannel(dc.Label()) })
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				peer.mutexRW.RLock()
				mws := newDCChain(peer.receiveMiddlewares)
				peer.mutexRW.RUnlock()
				m := mws.Process(PeerMessageProcessFunc(func(args PeerMessageProcessArgs) {
					if handler, ok := p.onReceive.Load().(PeerOnReceiveCallback); ok && handler != nil {
						if err := handler(args.Peer, args.Payload); err != nil {
							kitlog.Error(Logger()).Log("error", "onReceiveCallback failed", "reason", err.Error())
						}
					}
				}))
				m.Process(PeerMessageProcessArgs{Peer: peer, Payload: PeerMessagePayload{Channel: PeerMessageChannelWebrtc, Data: msg.Data}})
			})
		}
	}(p, mdc))

	p.mutexRW.Lock()
	p.dc = mdc
	p.mutexRW.Unlock()
	return mdc, nil
}

// Go start
func (p *PeerWebsocket) Go() {
	p.readyedChan = make(chan struct{})
	go p.receiver()
	go p.writer()

	for i := 0; i < 2; i++ {
		<-p.readyedChan
	}

	close(p.readyedChan)
}

// NewPeerWebsocket 创建
func NewPeerWebsocket(conn *websocket.Conn, readDeadline, writeDeadline time.Duration, maxMessageSize int64, notifyChan chan struct{}) *PeerWebsocket {
	if maxMessageSize < 1 {
		maxMessageSize = 10240
	}

	peer := &PeerWebsocket{
		id:                 uuid.NewV4().String(),
		conn:               conn,
		readDeadline:       readDeadline,
		writeDeadline:      writeDeadline,
		maxMessageSize:     maxMessageSize,
		messageType:        websocket.BinaryMessage,
		receiveMiddlewares: make(PeerMessageMiddlewares, 0),
		writeMiddlewares:   make(PeerMessageMiddlewares, 0),
		writeChan:          make(chan PeerMessagePayload, 1024),
		readyedChan:        make(chan struct{}),
		stopChan:           make(chan struct{}),
		notifyDieChan:      notifyChan,
	}

	// init params
	peer.params.Store(make(map[string]interface{}))
	return peer
}
