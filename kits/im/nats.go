package im

import (
	"fmt"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/kits/im/nats"
	grpcproto "github.com/golang/protobuf/proto"
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

	sName := ServiceName
	if frame.Header != nil {
		if m, ok := frame.Header["service"]; ok {
			sName = m
		}
	}
	fmt.Println("ftm->", msg)
	fmt.Println("ftm->", frame)
	resp, err := nrRouter.Handler(sName, &frame)
	if err != nil {
		log.Error(Logger()).Log("action", "Handler", "error", err, "frame", frame.Command)
		return
	}

	if resp == nil {
		return
	}

	// 不需要回复消息,防止出现消息死循环
	// reply, _ := grpcproto.Marshal(resp)
	// if err := msg.Respond(reply); err != nil {
	// 	log.Error(Logger()).Log("action", "msg.Respond", "error", err)
	// }
}
