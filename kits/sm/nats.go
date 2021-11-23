package sm

import (
	"errors"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/cores/pool/worker"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
	natsgo "github.com/nats-io/nats.go"
)

const (
	// ChannelStateEventBroadcast 状态变更时通知通道
	ChannelStateEventBroadcast = "sm.broadcast"

	// ChannelStateEventReceiver 接收状态事件通道
	ChannelStateEventReceiver = "sm.receiver"

	// ChannelStateEventInternalBroadcast 内部广播数据同步
	ChannelStateEventInternalBroadcast = "sm.internal.broadcast"
)

// NewNatsProcessActor nats
func NewNatsProcessActor(config conf.Nats) (*os.ProcessActor, error) {
	if err := nats.Connect(config, Logger()); err != nil {
		return nil, err
	}

	nc := nats.Conn()
	msgChan := make(chan *natsgo.Msg, 1)
	if _, err := nc.ChanQueueSubscribe(ChannelStateEventReceiver, ChannelStateEventReceiver, msgChan); err != nil {
		return nil, err
	}

	msgInternalChan := make(chan *natsgo.Msg, 1)
	if _, err := nc.ChanSubscribe(ChannelStateEventInternalBroadcast, msgInternalChan); err != nil {
		return nil, err
	}

	exitChan := make(chan struct{})
	return &os.ProcessActor{
		Exec: func() error {
			defer func() {
				nc := nats.Conn()
				nc.Drain()
				close(msgChan)
				Logger().Log("transport", "nats", "on", "shutdown")
			}()

			Logger().Log("transport", "nats", "on", config.Name)
			workers := worker.New(config.MaxWorkers)
			fn := func(m *natsgo.Msg) func() {
				return func() {
					onFromNatsMessage(m)
				}
			}

			internalfn := func(m *natsgo.Msg) func() {
				return func() {
					onFromNatsInternalMessage(m)
				}
			}

			for {
				select {
				case msg, ok := <-msgChan:
					if !ok {
						return nil
					}
					workers.Submit(fn(msg))

				case internalMsg, ok := <-msgInternalChan:
					if !ok {
						return nil
					}
					workers.Submit(internalfn(internalMsg))

				case <-exitChan:
					return nil
				}
			}
		},

		Interrupt: func(err error) {
		},

		Close: func() {
			close(exitChan)
		},
	}, nil
}

// onFromNatsMessage 处理来至nats订阅的消息
func onFromNatsMessage(msg *natsgo.Msg) {
	req := coresproto.RequestBytes{}
	if err := req.Unmarshal(msg.Data); err != nil {
		log.Error(Logger()).Log("action", "onFromNatsMessage", "error", err, "frame", msg.Data)
		return
	}

	resp, err := nrRouter.Handler(req.Cmd.Int32(), &corespb.Request{Header: make(map[string]string), Command: req.SubCmd.Int32(), Payload: req.Content})
	if err != nil {
		log.Error(Logger()).Log("action", "Handler", "error", err)
		return
	}

	if resp == nil || req.Ver == 0 {
		return
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		req.Content = payload.Content
	case *corespb.Response_Error:
		log.Error(Logger()).Log("action", "Handler", "error", string(payload.Error.Message), "code", payload.Error.Code)
		return
	}

	// 防止出现消息死循环
	reply, _ := req.Marshal()
	if err := msg.Respond(reply); err != nil {
		log.Error(Logger()).Log("action", "msg.Respond", "error", err)
	}
}

func onFromNatsInternalMessage(msg *natsgo.Msg) {
	req := coresproto.RequestBytes{}
	if err := req.Unmarshal(msg.Data); err != nil {
		log.Error(Logger()).Log("action", "onFromNatsInternalMessage", "error", err, "frame", msg.Data)
		return
	}

	_, err := nInternalRouter.Handler(&corespb.Request{Header: make(map[string]string), Command: req.SubCmd.Int32(), Payload: req.Content})
	if err != nil {
		log.Error(Logger()).Log("action", "onFromNatsInternalMessage", "error", err)
		return
	}
}

func broadcastEvent(data []byte) error {
	nc := nats.Conn()
	if nc == nil {
		return errors.New("nats is nil")
	}

	if err := nc.Publish(ChannelStateEventBroadcast, data); err != nil {
		return err
	}

	return nc.FlushTimeout(time.Second)
}

func internalBroadcastEvent(cmd coresproto.Command, evt *pb.SM_Event) error {
	req := coresproto.RequestBytes{
		Cmd:    kit.SM,
		SubCmd: cmd,
		SeqID:  1,
	}

	evtBytes, err := grpcproto.Marshal(evt)
	if err != nil {
		return err
	}

	req.Content = evtBytes
	data, err := req.Marshal()
	if err != nil {
		return err
	}

	nc := nats.Conn()
	if nc == nil {
		return errors.New("nats is nil")
	}

	if err := nc.Publish(ChannelStateEventInternalBroadcast, data); err != nil {
		return err
	}

	return nc.FlushTimeout(time.Second)
}
