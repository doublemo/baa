package imf

import (
	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/kits/imf/nats"
	grpcproto "github.com/golang/protobuf/proto"
	natsgo "github.com/nats-io/nats.go"
)

// NewNatsProcessActor nats
func NewNatsProcessActor(config conf.Nats) (*os.ProcessActor, error) {
	if err := nats.Connect(config, Logger()); err != nil {
		return nil, err
	}

	nc := nats.Conn()
	msgChan := make(chan *natsgo.Msg, config.ChanSubscribeBuffer)
	if _, err := nc.ChanSubscribe(config.Name, msgChan); err != nil {
		return nil, err
	}

	if _, err := nc.ChanQueueSubscribe(NatsGroupSubject(), ServiceName, msgChan); err != nil {
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
			for {
				select {
				case msg, ok := <-msgChan:
					if !ok {
						return nil
					}
					go onFromNatsMessage(msg)

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
	var frame corespb.Request
	{
		if err := grpcproto.Unmarshal(msg.Data, &frame); err != nil {
			log.Error(Logger()).Log("action", "onFromNatsMessage", "error", err, "frame", msg.Data)
			return
		}
	}

	resp, err := nr.Handler(&frame)
	if err != nil {
		log.Error(Logger()).Log("action", "Handler", "error", err, "frame", frame.Command)
		return
	}

	if resp == nil {
		return
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Error:
		log.Error(Logger()).Log("action", "onFromNatsMessage", "code", payload.Error.Code, "error", string(payload.Error.Message))
		return

	case *corespb.Response_Content:
		reply, _ := grpcproto.Marshal(&corespb.Request{
			Command: resp.Command,
			Header:  resp.Header,
			Payload: payload.Content,
		})

		if err := msg.Respond(reply); err != nil {
			log.Error(Logger()).Log("action", "msg.Respond", "error", err)
		}
	}
}

// NatsGroupSubject nats group queue
func NatsGroupSubject() string {
	return ServiceName + "_filter"
}
