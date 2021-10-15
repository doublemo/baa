package nats

import (
	"fmt"
	"strings"
	"time"

	coreslog "github.com/doublemo/baa/cores/log"
	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/internal/conf"
	"github.com/nats-io/nats.go"
)

var (
	nc *nats.Conn
	js nats.JetStreamContext
)

// Connect 连接到nats
func Connect(config conf.Nats, logger coreslog.Logger) error {
	var err error
	opts := make([]nats.Option, 0)
	opts = append(opts, nats.Name(config.Name))
	opts = append(opts, nats.RetryOnFailedConnect(true))
	opts = append(opts, nats.MaxReconnects(config.MaxReconnects))
	opts = append(opts, nats.ReconnectWait(time.Duration(config.ReconnectWait)*time.Second))
	opts = append(opts, nats.PingInterval(time.Second*time.Duration(config.PingInterval)))
	if len(config.ReconnectJitter) == 2 {
		opts = append(opts, nats.ReconnectJitter(time.Duration(config.ReconnectJitter[0])*time.Millisecond, time.Duration(config.ReconnectJitter[1])*time.Second))
	}

	opts = append(opts, nats.ReconnectHandler(func(c *nats.Conn) {
		log.Info(logger).Log("nats", "reconnect", "addr", c.ConnectedAddr())
	}))

	opts = append(opts, nats.ClosedHandler(func(c *nats.Conn) {
		if err := c.LastError(); err != nil {
			log.Error(logger).Log("nats", "reconnect", "error", err)
		}
	}))

	opts = append(opts, nats.DisconnectErrHandler(func(c *nats.Conn, e error) {
		if e != nil {
			log.Error(logger).Log("nats", "disconnected", "error", fmt.Sprintf("Disconnected due to:%v, will attempt reconnects for %ds", err, config.ReconnectWait))
		}
	}))

	if config.Authentication.UserCreds != "" {
		opts = append(opts, nats.UserCredentials(config.Authentication.UserCreds))
	}

	if config.Authentication.TLSClientCert != "" && config.Authentication.TLSClientKey != "" {
		opts = append(opts, nats.ClientCert(config.Authentication.TLSClientCert, config.Authentication.TLSClientKey))
	}

	if config.Authentication.TLSCACert != "" {
		opts = append(opts, nats.RootCAs(config.Authentication.TLSCACert))
	}

	if config.Authentication.NkeyFile != "" {
		o, err := nats.NkeyOptionFromSeed(config.Authentication.NkeyFile)
		if err != nil {
			return err
		}

		opts = append(opts, o)
	}

	nc, err = nats.Connect(strings.Join(config.Urls, ","), opts...)
	if err != nil {
		return err
	}

	js, err = nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return err
	}

	if err := nc.FlushTimeout(time.Second * 1); err != nil {
		log.Warn(logger).Log("nats", "FlushTimeout", "error", err)
	}

	if err := nc.LastError(); err != nil {
		return err
	}

	return nil
}

// Conn 获取nats连接
func Conn() *nats.Conn {
	return nc
}

// JetStream 获取jetstream
func JetStream() nats.JetStreamContext {
	return js
}

// eg:
// _, err = js.Subscribe("ORDERS.*", func(msg *nats.Msg) {
// 	msg.Ack()
// 	fmt.Println("-x-x-x-x-x-x-x")
// 	if handler, ok := onReceiveJetStream.Load().(func(*nats.Msg)); ok && handler != nil {
// 		handler(msg)
// 	}
// },
// 	nats.Durable("MONITOR"),
// 	nats.ManualAck(),
// 	nats.DeliverLast(),
// )

//nc := nats.Conn()

// if nc.IsConnected() {
// 	nc.Publish(config.Name, []byte("哪哪里的会"))
// }

//js := nats.JetStreamContext()
// i, err := js.AddStream(&natsgo.StreamConfig{
// 	Name:              "ORDERS",
// 	Subjects:          []string{"ORDERS.*"},
// 	Storage:           natsgo.FileStorage,
// 	MaxMsgs:           -1,
// 	MaxBytes:          -1,
// 	MaxAge:            time.Hour * 24,
// 	Retention:         natsgo.LimitsPolicy,
// 	MaxMsgSize:        -1,
// 	Discard:           natsgo.DiscardOld,
// 	MaxConsumers:      10,
// 	MaxMsgsPerSubject: -1,
// 	Replicas:          1,
// 	NoAck:             false,
// })

// fmt.Println("stream info:", i, err)

// s, err := js.AddConsumer("ORDERS", &natsgo.ConsumerConfig{
// 	Durable: "New",
// 	//DeliverSubject: "ORDERS.received",
// 	FilterSubject: "",
// 	AckPolicy:     natsgo.AckExplicitPolicy,
// 	AckWait:       30 * time.Second,
// 	MaxDeliver:    20,
// 	DeliverPolicy: natsgo.DeliverAllPolicy,
// 	ReplayPolicy:  natsgo.ReplayInstantPolicy,
// })

// fmt.Println("consumer info:", s, err)

// s1, err := js.AddConsumer("ORDERS", &natsgo.ConsumerConfig{
// 	Durable:        "MONITOR",
// 	DeliverSubject: "monitor.ORDERS",
// 	FilterSubject:  "",
// 	AckPolicy:      natsgo.AckNonePolicy,
// 	MaxDeliver:     -1,
// 	DeliverPolicy:  natsgo.DeliverLastPolicy,
// 	ReplayPolicy:   natsgo.ReplayInstantPolicy,
// })

// fmt.Println("consumer info2:", s1, err)

// _, err := js.Publish("ORDERS", []byte("咋回事呢"))
// fmt.Println("Publish:", err)
// nc.Flush()
