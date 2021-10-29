package agent

import (
	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/cores/pool/worker"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/nats"
	natsgo "github.com/nats-io/nats.go"
)

// NewNatsProcessActor nats
func NewNatsProcessActor(config conf.Nats) (*os.ProcessActor, error) {
	if err := nats.Connect(config, Logger()); err != nil {
		return nil, err
	}

	nc := nats.Conn()
	msgChan := make(chan *natsgo.Msg, 1)
	nc.ChanSubscribe(config.Name, msgChan)
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

			for {
				select {
				case msg, ok := <-msgChan:
					if !ok {
						return nil
					}

					workers.Submit(fn(msg))

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
