package router

import (
	"context"
	"errors"
	"sync"
	"time"

	grpcpool "github.com/doublemo/baa/cores/pool/grpc"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"google.golang.org/grpc"
)

type Call struct {
	c     conf.RPCClient
	pool  *grpcpool.Pool
	mutex sync.RWMutex
}

func (r *Call) Serve(req *corespb.Request) (*corespb.Response, error) {
	var err error
	r.mutex.RLock()
	p := r.pool
	r.mutex.RUnlock()

	if p == nil {
		p, err = r.createPool()
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	conn, err := p.Get(ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
	defer func() {
		conn.Close()
		cancel2()
	}()

	client := corespb.NewServiceClient(conn.ClientConn)
	resp, err := client.Call(ctx2, req)
	return resp, err
}

func (r *Call) ServeHTTP(req *corespb.Request) (*corespb.Response, error) {
	return nil, errors.New("notsupported")
}

func (r *Call) createPool() (*grpcpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	fn := func(ctx context.Context) (*grpc.ClientConn, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		conn, err := rpc.NewConnect(r.c)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	p, err := grpcpool.NewWithContext(ctx, fn, r.c.Pool.Init, r.c.Pool.Capacity, time.Duration(r.c.Pool.IdleTimeout)*time.Minute, time.Duration(r.c.Pool.MaxLife)*time.Minute)
	if err != nil {
		return nil, err
	}

	r.mutex.Lock()
	if r.pool == nil {
		r.pool = p
	} else {
		p = r.pool
	}
	r.mutex.Unlock()
	return p, nil
}

// NewCall 创建Call
func NewCall(c conf.RPCClient) *Call {
	return &Call{c: c}
}
