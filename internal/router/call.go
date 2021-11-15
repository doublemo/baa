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
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

type Call struct {
	c     conf.RPCClient
	pool  map[string]*grpcpool.Pool
	mutex sync.RWMutex
}

func (r *Call) Serve(req *corespb.Request) (*corespb.Response, error) {
	var (
		err  error
		addr string = "default"
	)

	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	if m, ok := req.Header["Host"]; ok && m != "" {
		addr = m
	}

	r.mutex.RLock()
	p := r.pool[addr]
	r.mutex.RUnlock()

	if p == nil {
		p, err = r.createPool(addr)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	conn, err := p.Get(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		conn.Close()
	}()

	client := corespb.NewServiceClient(conn.ClientConn)
	resp, err := client.Call(ctx, req)
	return resp, err
}

func (r *Call) ServeHTTP(req *corespb.Request) (*corespb.Response, error) {
	return nil, errors.New("notsupported")
}

func (r *Call) createPool(addr string) (*grpcpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	fn := r.selectFn(addr)
	p, err := grpcpool.NewWithContext(ctx, fn, r.c.Pool.Init, r.c.Pool.Capacity, time.Duration(r.c.Pool.IdleTimeout)*time.Minute, time.Duration(r.c.Pool.MaxLife)*time.Minute)
	if err != nil {
		return nil, err
	}

	r.mutex.Lock()
	r.pool[addr] = p
	r.mutex.Unlock()
	return p, nil
}

func (r *Call) selectFn(addr string) func(ctx context.Context) (*grpc.ClientConn, error) {
	if addr == "default" {
		return func(ctx context.Context) (*grpc.ClientConn, error) {
			conn, err := rpc.NewConnectContext(ctx, r.c)
			if err != nil {
				return nil, err
			}

			return conn, nil
		}
	}

	return func(ctx context.Context) (*grpc.ClientConn, error) {
		opts := []grpc.DialOption{grpc.WithBlock()}
		if len(r.c.Key) > 0 && len(r.c.Salt) > 0 {
			creds, err := credentials.NewClientTLSFromFile(r.c.Salt, r.c.Key)
			if err != nil {
				return nil, err
			}

			opts = append(opts, grpc.WithTransportCredentials(creds))
			opts = append(opts, grpc.WithPerRPCCredentials(oauth.NewOauthAccess(
				&oauth2.Token{AccessToken: r.c.ServiceSecurityKey},
			)))
		} else {
			opts = append(opts, grpc.WithInsecure())
		}

		conn, err := grpc.DialContext(ctx, addr, opts...)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

// NewCall 创建Call
func NewCall(c conf.RPCClient) *Call {
	return &Call{c: c, pool: make(map[string]*grpcpool.Pool)}
}
