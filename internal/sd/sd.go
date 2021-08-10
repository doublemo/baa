package sd

import (
	"context"
	"strconv"
	"time"

	"github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/cores/sd/etcdv3"
	"github.com/doublemo/baa/internal/conf"
	"google.golang.org/grpc"
)

var (
	endpoint   sd.Endpoint
	endpointer sd.Endpointer
	client     etcdv3.Client
	prefix     string
)

//Endpoint  获取节点信息
func Endpoint() sd.Endpoint {
	return endpoint
}

// Endpointer 获取节点发现
func Endpointer() sd.Endpointer {
	return endpointer
}

// Client 获取etcdv3 客户端
func Client() etcdv3.Client {
	return client
}

// Prefix 获取连接前缀
func Prefix() string {
	return prefix
}

// Init 初始化节点信息
func Init(machineID string, etcd *conf.Etcd, rpc *conf.RPC) error {
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

	instancer, err := etcdv3.NewInstancer(c, etcd.BasePath)
	if err != nil {
		return err
	}

	client = c
	prefix = etcd.BasePath
	endpointer = sd.NewEndpointer(instancer)

	if rpc != nil {
		endpoint = sd.NewEndpoint(machineID, rpc.Name, rpc.Addr)
		endpoint.Set("group", rpc.Group)
		endpoint.Set("weight", strconv.Itoa(rpc.Weight))
	}
	return nil
}
