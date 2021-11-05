package sd

import (
	"context"
	"errors"
	"time"

	"github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/cores/sd/etcdv3"
	"github.com/doublemo/baa/internal/conf"
	"google.golang.org/grpc"
)

var (
	endpoint         sd.Endpoint
	endpointer       sd.Endpointer
	instancer        sd.Instancer
	client           etcdv3.Client
	prefix           string
	ErrEndpointerNil = errors.New("ErrEndpointerNil")
)

// Init 初始化节点信息
func Init(etcd conf.Etcd, e sd.Endpoint) error {
	config := etcdv3.Config{
		Addrs:         etcd.Addr,
		DialTimeout:   3 * time.Second,
		DialKeepAlive: 3 * time.Second,
		DialOptions:   []grpc.DialOption{grpc.WithBlock()},
	}

	c, err := etcdv3.NewClient(context.Background(), config)
	if err != nil {
		return err
	}

	instance, err := etcdv3.NewInstancer(c, etcd.BasePath)
	if err != nil {
		return err
	}

	client = c
	prefix = etcd.BasePath
	endpointer = sd.NewEndpointer(instance, sd.InvalidateOnError(time.Second))
	endpoint = e
	instancer = instance
	return nil
}

//Endpoint  获取节点信息
func Endpoint() sd.Endpoint {
	return endpoint
}

// Endpointer 获取节点发现
func Endpointer() sd.Endpointer {
	return endpointer
}

func Instancer() sd.Instancer {
	return instancer
}

// Client 获取etcdv3 客户端
func Client() etcdv3.Client {
	return client
}

// Prefix 获取连接前缀
func Prefix() string {
	return prefix
}

// Endpoints 获取所有节点
func Endpoints() ([]sd.Endpoint, error) {
	if endpointer == nil {
		return nil, ErrEndpointerNil
	}

	return endpointer.Endpoints()
}
