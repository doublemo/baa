package agent

import (
	"context"
	"strconv"
	"sync"
	"time"

	grpcpool "github.com/doublemo/baa/cores/pool/grpc"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/session"
	"github.com/doublemo/baa/kits/sfu/errcode"
	"google.golang.org/grpc"
)

type authenticationRouter struct {
	c     conf.RPCClient
	pool  *grpcpool.Pool
	mutex sync.RWMutex
}

func (r *authenticationRouter) Serve(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
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

	client := corespb.NewServiceClient(conn.ClientConn)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		conn.Close()
		cancel2()
	}()

	resp, err := client.Call(ctx2, &corespb.Request{
		Header:  map[string]string{"PeerId": peer.ID(), "seqno": strconv.FormatUint(uint64(req.SID()), 10)},
		Command: req.SubCommand().Int32(),
		Payload: req.Body(),
	})
	if err != nil {
		return proto.NewResponseBytes(req.Command(), errcode.Bad(&corespb.Response{Command: req.SubCommand().Int32()}, errcode.ErrInternalServer, grpc.ErrorDesc(err))), nil
	}

	w := proto.NewResponseBytes(req.Command(), resp)
	return w, nil
}

func (r *authenticationRouter) createPool() (*grpcpool.Pool, error) {
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

// Destroy 清理
func (r *authenticationRouter) Destroy(peer session.Peer) {
}

func newAuthenticationRouter(config conf.RPCClient) *authenticationRouter {
	return &authenticationRouter{
		c: config,
	}
}