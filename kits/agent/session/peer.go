package session

import (
	"errors"
)

var (
	// ErrPeerWriteMessageFailed 输入信息错误
	ErrPeerWriteMessageFailed = errors.New("ErrPeerWriteMessageFailed")

	// ErrPeerChannelIsFulled 写入通道已经满
	ErrPeerChannelIsFulled = errors.New("ErrPeerWriteMessageFailed")
)

type (
	// PeerOnReceiveCallback Peer接收信息回调
	PeerOnReceiveCallback func(Peer, PeerMessagePayload) error

	// PeerOnWriteCallback Peer发送数据时
	PeerOnWriteCallback func(PeerMessagePayload) PeerMessagePayload

	// PeerOnCloseCallback Peer关闭时回调
	PeerOnCloseCallback func(Peer)

	// Peer 连接
	Peer interface {
		ID() string
		OnReceive(PeerOnReceiveCallback)
		OnWrite(PeerOnWriteCallback)
		OnClose(PeerOnCloseCallback)
		Use(...PeerMessageMiddleware)
		Send(PeerMessagePayload) error
		LoadOrResetSeqNo(...uint32) uint32
		Params(string) (interface{}, bool)
		SetParams(string, interface{})
		Close() error
	}

	// PeerMessagePayload 信息结构
	PeerMessagePayload struct {
		Type int
		Data []byte
	}

	// PeerMessageMiddleware 信息处理中间件
	PeerMessageMiddleware interface {
		Receive() func(PeerMessageProcessor) PeerMessageProcessor
		Write() func(PeerMessageProcessor) PeerMessageProcessor
	}

	// PeerMessageProcessArgs 信息处理中间参数
	PeerMessageProcessArgs struct {
		Peer    Peer
		Payload PeerMessagePayload
	}

	// PeerMessageMiddlewares 中间
	PeerMessageMiddlewares []func(PeerMessageProcessor) PeerMessageProcessor

	// PeerMessageProcessor 中间件类型接口
	PeerMessageProcessor interface {
		Process(args PeerMessageProcessArgs)
	}

	// PeerMessageProcessFunc 中间件函数类型
	PeerMessageProcessFunc func(args PeerMessageProcessArgs)

	chainHandler struct {
		middlewares PeerMessageMiddlewares
		Last        PeerMessageProcessor
		current     PeerMessageProcessor
	}
)

func (p PeerMessageProcessFunc) Process(args PeerMessageProcessArgs) {
	p(args)
}

func (mws PeerMessageMiddlewares) Process(h PeerMessageProcessor) PeerMessageProcessor {
	return &chainHandler{mws, h, chain(mws, h)}
}

func (mws PeerMessageMiddlewares) ProcessFunc(h PeerMessageProcessor) PeerMessageProcessor {
	return &chainHandler{mws, h, chain(mws, h)}
}

func newDCChain(m []func(p PeerMessageProcessor) PeerMessageProcessor) PeerMessageMiddlewares {
	return PeerMessageMiddlewares(m)
}

func (c *chainHandler) Process(args PeerMessageProcessArgs) {
	c.current.Process(args)
}

func chain(mws []func(processor PeerMessageProcessor) PeerMessageProcessor, last PeerMessageProcessor) PeerMessageProcessor {
	if len(mws) == 0 {
		return last
	}
	h := mws[len(mws)-1](last)
	for i := len(mws) - 2; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// func xN(next PeerMessageProcessor) PeerMessageProcessor {
// 	return PeerMessageProcessFunc(func(m PeerMessageProcessArgs) {
// 		fmt.Println("okkkkk---1", m)
// 		next.Process(m)
// 	})
// }

// func xN2(next PeerMessageProcessor) PeerMessageProcessor {
// 	return PeerMessageProcessFunc(func(m PeerMessageProcessArgs) {
// 		fmt.Println("okkkkk---2", m)
// 		next.Process(m)
// 	})
// }
