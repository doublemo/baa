package agent

import (
	"context"
	"strconv"
	"sync"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	grpcpool "github.com/doublemo/baa/cores/pool/grpc"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type sfuRouter struct {
	c       conf.RPCClient
	pool    *grpcpool.Pool
	clients map[string]*rpc.BidirectionalStreamingClient
	mutex   sync.RWMutex
}

func (r *sfuRouter) Serve(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	client, err := r.getClient(peer)
	if err != nil {
		return nil, err
	}

	err = client.Send(&corespb.Request{
		Header:  map[string]string{"PeerId": peer.ID(), "seqno": strconv.FormatUint(uint64(req.SID()), 10)},
		Command: req.SubCommand().Int32(),
		Payload: req.Body(),
	})

	return nil, err
}

// Destroy 清理
func (r *sfuRouter) Destroy(peer session.Peer) {
	r.mutex.RLock()
	client, ok := r.clients[peer.ID()]
	r.mutex.RUnlock()
	if !ok {
		return
	}

	client.Close()
	r.mutex.Lock()
	delete(r.clients, peer.ID())
	r.mutex.Unlock()
}

func (r *sfuRouter) getClient(peer session.Peer) (*rpc.BidirectionalStreamingClient, error) {
	r.mutex.RLock()
	client, ok := r.clients[peer.ID()]
	r.mutex.RUnlock()
	if ok {
		return client, nil
	}

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

	defer conn.Close()

	client = rpc.NewBidirectionalStreamingClient(conn.ClientConn, Logger())
	client.OnReceive = func(r *corespb.Response) {
		w := proto.NewResponseBytes(proto.SFU, r)
		bytes, _ := w.Marshal()
		if err := peer.Send(session.PeerMessagePayload{Data: bytes}); err != nil {
			log.Error(Logger()).Log("error", err)
		}
	}

	md := metadata.Pairs("PeerId", peer.ID())
	if err := client.Connect(md); err != nil {
		return nil, err
	}

	r.mutex.Lock()
	r.clients[peer.ID()] = client
	r.mutex.Unlock()
	return client, nil
}

func (r *sfuRouter) createPool() (*grpcpool.Pool, error) {
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

func newSFURouter(config conf.RPCClient) *sfuRouter {
	return &sfuRouter{
		c:       config,
		clients: make(map[string]*rpc.BidirectionalStreamingClient),
	}
}
