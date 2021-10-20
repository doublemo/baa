package im

import (
	"context"
	"errors"
	"sync"
	"time"

	grpcpool "github.com/doublemo/baa/cores/pool/grpc"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/kits/im/cache"
	"github.com/doublemo/baa/kits/snid"
	snproto "github.com/doublemo/baa/kits/snid/proto"
	snpb "github.com/doublemo/baa/kits/snid/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

type snidRouter struct {
	c     conf.RPCClient
	pool  *grpcpool.Pool
	mutex sync.RWMutex
}

func (r *snidRouter) Serve(req *corespb.Request) (*corespb.Response, error) {
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
	defer cancel()
	conn, err := p.Get(ctx)
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

func (r *snidRouter) ServeHTTP(req *corespb.Request) (*corespb.Response, error) {
	return nil, errors.New("notsupported")
}

func (r *snidRouter) createPool() (*grpcpool.Pool, error) {
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

func newSnidRouter(c conf.RPCClient) *snidRouter {
	return &snidRouter{c: c}
}

func getSNID(num int32) ([]uint64, error) {
	if num > 100 {
		return nil, errors.New("the number cannot be greater then 100")
	}

	frame := snpb.SNID_Request{N: num}
	b, _ := grpcproto.Marshal(&frame)
	resp, err := muxRouter.Handler(snid.ServiceName, &corespb.Request{Command: snproto.SnowflakeCommand.Int32(), Payload: b})
	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		resp := snpb.SNID_Reply{}
		if err := grpcproto.Unmarshal(payload.Content, &resp); err != nil {
			return nil, err
		}

		if len(resp.Values) != int(num) {
			return nil, errors.New("errorSNIDLen")
		}

		return resp.Values, nil

	case *corespb.Response_Error:
		return nil, errors.New(payload.Error.Message)
	}

	return nil, errors.New("snid failed")
}

func getSnowflakeID() (uint64, error) {
	if cache.GetSnowflakeLen() > 0 {
		return cache.GetSnowflakeID(), nil
	}

	values, err := getSNID(100)
	if err != nil {
		return 0, err
	}

	cache.ResetSnowflakeID(values...)
	return cache.GetSnowflakeID(), nil
}
