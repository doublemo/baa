package router

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	coreslog "github.com/doublemo/baa/cores/log"
	grpcpool "github.com/doublemo/baa/cores/pool/grpc"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/kits/agent/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	// Stream grpc stream
	Stream struct {
		c               conf.RPCClient
		pool            *grpcpool.Pool
		clients         map[string]*rpc.BidirectionalStreamingClient
		logger          coreslog.Logger
		mutex           sync.RWMutex
		onBeforeConnect atomic.Value
		onClose         atomic.Value
		onSend          atomic.Value
		onReceive       atomic.Value
		onDestroy       atomic.Value
		allowCommands   map[int32]bool
	}

	// StreamOptions opts
	StreamOptions func(c *Stream)
)

func (r *Stream) Serve(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	client, err := r.getClient(peer)
	if err != nil {
		return nil, err
	}

	request := &corespb.Request{
		Header:  map[string]string{"PeerId": peer.ID(), "seqno": strconv.FormatUint(uint64(req.SID()), 10)},
		Command: req.SubCommand().Int32(),
		Payload: req.Body(),
	}

	if handler, ok := r.onSend.Load().(func(session.Peer, coresproto.Request, *corespb.Request) error); ok && handler != nil {
		if err := handler(peer, req, request); err != nil {
			return nil, err
		}
	}

	return nil, client.Send(request)
}

// Destroy 清理
func (r *Stream) Destroy(peer session.Peer) {
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

	if handler, ok := r.onDestroy.Load().(func(session.Peer)); ok && handler != nil {
		handler(peer)
	}
}

func (r *Stream) getClient(peer session.Peer) (*rpc.BidirectionalStreamingClient, error) {
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

	client = rpc.NewBidirectionalStreamingClient(conn.ClientConn, r.logger)
	client.OnReceive = func(p session.Peer, s *Stream) func(*corespb.Response) {
		return func(resp *corespb.Response) {
			if handler, ok := s.onReceive.Load().(func(session.Peer, *corespb.Response)); ok && handler != nil {
				handler(p, resp)
			}
		}
	}(peer, r)

	if handler, ok := r.onClose.Load().(func()); ok && handler != nil {
		client.OnClose = handler
	}

	md := metadata.Pairs("PeerId", peer.ID())
	if handler, ok := r.onBeforeConnect.Load().(func(md metadata.MD) metadata.MD); ok && handler != nil {
		md = handler(md)
	}

	if err := client.Connect(md); err != nil {
		return nil, err
	}

	r.mutex.Lock()
	r.clients[peer.ID()] = client
	r.mutex.Unlock()
	return client, nil
}

func (r *Stream) createPool() (*grpcpool.Pool, error) {
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

// OnBeforeConnect Hook
func (r *Stream) OnBeforeConnect(f func(md metadata.MD) metadata.MD) {
	if f == nil {
		return
	}

	r.onBeforeConnect.Store(f)
}

// OnClose Hook
func (r *Stream) OnClose(f func(md metadata.MD) metadata.MD) {
	if f == nil {
		return
	}

	r.onClose.Store(f)
}

// OnSend Hook
func (r *Stream) OnSend(f func(session.Peer, coresproto.Request, *corespb.Request) error) {
	if f == nil {
		return
	}

	r.onSend.Store(f)
}

// OnReceive Hook
func (r *Stream) OnReceive(f func(session.Peer, *corespb.Response)) {
	if f == nil {
		return
	}
	r.onReceive.Store(f)
}

// OnDestroy Hook
func (r *Stream) OnDestroy(f func(session.Peer)) {
	if f == nil {
		return
	}

	r.onDestroy.Store(f)
}

func NewStream(config conf.RPCClient, logger coreslog.Logger, opts ...StreamOptions) *Stream {
	s := &Stream{
		c:             config,
		clients:       make(map[string]*rpc.BidirectionalStreamingClient),
		allowCommands: make(map[int32]bool),
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

// AllowCommandsStreamOptions 设置允许通过的有效命令
func AllowCommandsStreamOptions(commands ...int32) StreamOptions {
	return func(r *Stream) {
		for _, c := range commands {
			r.allowCommands[c] = true
		}
	}
}
