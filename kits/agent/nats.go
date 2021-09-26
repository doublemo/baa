package agent

import (
	"fmt"

	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/kits/agent/nats"
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

					onFromNatsMessage(msg)

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
	fmt.Println("from nats message:", string(msg.Data))
}
