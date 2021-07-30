package sfu

import (
	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/cores/sd/etcdv3"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/kits/sfu/sd"
)

// NewServiceDiscoveryProcessActor 创建服务发现
func NewServiceDiscoveryProcessActor(machineID string, conf conf.RPC, etcd conf.Etcd) (*os.ProcessActor, error) {
	if err := sd.Init(machineID, conf, etcd); err != nil {
		return nil, err
	}

	registrar := etcdv3.NewRegistrar(sd.Client(), etcdv3.Service{Prefix: etcd.BasePath, Endpoint: sd.Endpoint()})
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
