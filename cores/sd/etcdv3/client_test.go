package etcdv3

import (
	"context"
	"testing"
	"time"

	"github.com/doublemo/baa/cores/sd"
	"google.golang.org/grpc"
)

const addr string = "127.0.0.1:2379"

func TestNewClientLocal(t *testing.T) {
	c := Config{
		Addrs:         []string{addr},
		DialTimeout:   3 * time.Second,
		DialKeepAlive: 3 * time.Second,
		DialOptions:   []grpc.DialOption{grpc.WithBlock()},
	}
	client, err := NewClient(context.Background(), c)
	if err != nil {
		t.Fatal(err)
		return
	}

	client.Register(Service{Prefix: "/services/baa/xx", Endpoint: sd.NewEndpoint("test01", "test", "")})
	t.Log(client.GetEntries("/services/baa"))

	ch := make(chan struct{})
	go client.WatchPrefix("/services/baa", ch)

	<-ch
	t.Log(client.GetEntries("/services/baa"))
	t.Log("ddd----")
	//select {}
}
