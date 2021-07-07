// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package conf

import (
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
)

// Etcd etcd 配置信息
type Etcd struct {
	// Addr etcd集群地址
	Addr []string `alias:"addr"`

	// BasePath 服务前缀
	BasePath string `alias:"basepath" default:"/services/baa"`
}

// Clone etcd
func (etcd Etcd) Clone() *Etcd {
	return &Etcd{
		Addr:     etcd.Addr,
		BasePath: etcd.BasePath,
	}
}

// RPCXRegisterPlugin rpcx
func (etcd *Etcd) RPCXRegisterPlugin(servcieAddr string) (*serverplugin.EtcdV3RegisterPlugin, error) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: servcieAddr,
		EtcdServers:    etcd.Addr,
		BasePath:       etcd.BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}

	if err := r.Start(); err != nil {
		return nil, err
	}

	return r, nil
}
