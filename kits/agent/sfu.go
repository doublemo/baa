package agent

import (
	"context"
	"fmt"
	"io"
	"strconv"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/agent/adapter/router"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/session"
	sfu "github.com/doublemo/baa/kits/sfu"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// sfuRouter sfu 服务
func sfuRouter(config *RouterConfig) router.Callback {
	if config.Etcd == nil || config.Sfu == nil {
		panic("Invalid config in sfuRouter")
	}

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:///%s", sfu.ServiceName, config.Sfu.Group),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), // This sets the initial balancing policy.
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)

	if err != nil {
		panic(err)
	}

	client := corespb.NewServiceClient(conn)

	// return
	return func(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
		var (
			stream corespb.Service_BidirectionalStreamingClient
			err    error
		)
		sfuClient, ok := peer.Params("sfuClient")
		if !ok || sfuClient == nil {
			md := metadata.Pairs("PeerId", peer.ID())
			ctx := metadata.NewOutgoingContext(context.Background(), md)
			stream, err = client.BidirectionalStreaming(ctx)
			if err != nil {
				return nil, err
			}

			go sfuRecv(peer, conn, stream)
			peer.SetParams("sfuClient", stream)
		} else {
			stream, ok = sfuClient.(corespb.Service_BidirectionalStreamingClient)
			if !ok {
				return nil, nil
			}
		}

		err = stream.Send(&corespb.Request{
			Header:  map[string]string{"PeerId": peer.ID(), "seqno": strconv.FormatUint(uint64(req.SID()), 10)},
			Command: req.SubCommand().Int32(),
			Payload: req.Body(),
		})

		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func sfuRecv(peer session.Peer, cc *grpc.ClientConn, stream corespb.Service_BidirectionalStreamingClient) {
	defer func() {
		fmt.Println("ss-------------------------")
		cc.Close()
	}()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return
		}

		if err != nil {
			fmt.Println("recv err:", err)
			return
		}

		if resp == nil {
			continue
		}

		w := proto.NewResponseBytes(proto.SFUCommand, resp)
		bytes, _ := w.Marshal()
		peer.Send(session.PeerMessagePayload{Type: websocket.BinaryMessage, Data: bytes})
	}
}

func stopSFUGRPC(peer session.Peer) {
	sfuXClient, ok := peer.Params("sfuClient")
	if !ok || sfuXClient == nil {
		return
	}

	xclient, ok := sfuXClient.(corespb.Service_BidirectionalStreamingClient)
	if !ok {
		return
	}

	xclient.CloseSend()
	peer.SetParams("sfuClient", nil)
}
