package agent

import (
	"errors"
	"strconv"

	log "github.com/doublemo/baa/cores/log/level"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/kits/agent/adapter/router"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/session"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// sfuRouter sfu 服务
func sfuRouter(config *conf.RPCClient) router.Callback {
	if config == nil {
		panic("Invalid config in sfuRouter")
	}

	conn, err := rpc.NewConnect(config)
	if err != nil {
		panic(err)
	}

	// return
	return func(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
		client, ok := sfuRouterBidirectionalStreamingClient(peer, conn)
		if !ok {
			return nil, errors.New("BidirectionalStreamingClient is nil")
		}

		err := client.Send(&corespb.Request{
			Header:  map[string]string{"PeerId": peer.ID(), "seqno": strconv.FormatUint(uint64(req.SID()), 10)},
			Command: req.SubCommand().Int32(),
			Payload: req.Body(),
		})
		return nil, err
	}
}

func sfuRouterBidirectionalStreamingClient(peer session.Peer, conn ...*grpc.ClientConn) (*rpc.BidirectionalStreamingClient, bool) {
	if m, ok := peer.Params("SFUBidirectionalStreamingClient"); ok && m != nil {
		if v, ok := m.(*rpc.BidirectionalStreamingClient); ok {
			return v, true
		}
	}

	if len(conn) < 1 || conn[0] == nil {
		return nil, false
	}

	client := rpc.NewBidirectionalStreamingClient(conn[0], Logger())
	client.OnReceive = func(r *corespb.Response) {
		w := proto.NewResponseBytes(proto.SFUCommand, r)
		bytes, _ := w.Marshal()
		if err := peer.Send(session.PeerMessagePayload{Type: websocket.BinaryMessage, Data: bytes}); err != nil {
			log.Error(Logger()).Log("error", err)
		}
	}

	client.OnClose = func() {
		peer.SetParams("SFUBidirectionalStreamingClient", nil)
	}

	md := metadata.Pairs("PeerId", peer.ID())
	if err := client.Connect(md); err != nil {
		log.Error(Logger()).Log("error", err)
		return nil, false
	}

	peer.SetParams("SFUBidirectionalStreamingClient", client)
	return client, true
}
