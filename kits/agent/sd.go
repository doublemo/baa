package agent

import (
	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/cores/sd/etcdv3"
	"github.com/doublemo/baa/kits/agent/sd"
)

// NewServiceDiscoveryProcessActor 创建服务发现
func NewServiceDiscoveryProcessActor() (*os.ProcessActor, error) {
	registrar := etcdv3.NewRegistrar(sd.Client(), etcdv3.Service{Prefix: sd.Prefix(), Endpoint: sd.Endpoint()})
	ch := make(chan struct{})
	return &os.ProcessActor{
		Exec: func() error {
			Logger().Log("transport", "registrar", "on", sd.Endpoint().Marshal())
			registrar.Register()
			<-ch
			return nil
		},
		Interrupt: func(err error) {
			if err == nil {
				return
			}
			log.Error(Logger()).Log("transport", "registrar", "error", err)
		},

		Close: func() {
			close(ch)
			registrar.Deregister()
			Logger().Log("transport", "registrar", "on", "shutdown")
		},
	}, nil
}
