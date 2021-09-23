package agent

import (
	"fmt"
	"strconv"
	"sync"

	log "github.com/doublemo/baa/cores/log/level"
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
	c       *conf.RPCClient
	conn    *grpc.ClientConn
	clients map[string]*rpc.BidirectionalStreamingClient
	mutex   sync.RWMutex
}

func (r *sfuRouter) Serve(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	r.mutex.Lock()
	if r.conn == nil {
		conn, err := rpc.NewConnect(r.c)
		if err != nil {
			r.mutex.Unlock()
			return nil, err
		}

		r.conn = conn
	}
	r.mutex.Unlock()

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
	fmt.Println("--------x-----xx--", peer.ID())
}

func (r *sfuRouter) getClient(peer session.Peer) (*rpc.BidirectionalStreamingClient, error) {
	r.mutex.RLock()
	client, ok := r.clients[peer.ID()]
	r.mutex.RUnlock()
	if ok {
		return client, nil
	}

	client = rpc.NewBidirectionalStreamingClient(r.conn, Logger())
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

func newSFURouter(config *conf.RPCClient) *sfuRouter {
	if config == nil {
		panic("Invalid config in sfuRouter")
	}

	return &sfuRouter{
		c:       config,
		clients: make(map[string]*rpc.BidirectionalStreamingClient),
	}
}
