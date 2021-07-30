package sd

import (
	"context"
	"strconv"
	"time"

	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/cores/sd/etcdv3"
	"github.com/doublemo/baa/internal/conf"
	"google.golang.org/grpc"
)

var (
	endpoint   coressd.Endpoint
	endpointer coressd.Endpointer
	client     etcdv3.Client
)

func Endpoint() coressd.Endpoint {
	return endpoint
}

func Endpointer() coressd.Endpointer {
	return endpointer
}

func Client() etcdv3.Client {
	return client
}

func Init(machineID string, conf conf.RPC, etcd conf.Etcd) error {
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
	endpoint = coressd.NewEndpoint(machineID, conf.Name)
	endpoint.Set(coressd.FEndpointAddr, conf.Addr)
	endpoint.Set(coressd.FEndpointGroup, conf.Group)
	endpoint.Set(coressd.FEndpointWeight, strconv.Itoa(conf.Weight))
	endpointer = coressd.NewEndpointer(instancer)
	return nil
}
