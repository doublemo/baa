package sfu

import (
	"context"
	"fmt"
	"net"
	"sync"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/kits/sfu/adapter/router"
	"github.com/doublemo/baa/kits/sfu/session"
	ionsfu "github.com/pion/ion-sfu/pkg/sfu"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	baseserver struct {
		corespb.UnimplementedServiceServer
		mutex sync.Mutex
	}
)

func (s *baseserver) Call(ctx context.Context, req *corespb.Request) (*corespb.Response, error) {
	return nil, nil
}

func (s *baseserver) BidirectionalStreaming(stream corespb.Service_BidirectionalStreamingServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.DataLoss, "BidirectionalStreaming: failed to get metadata")
	}

	peermd, ok := md["peerid"]
	if !ok || len(peermd) < 1 {
		return status.Errorf(codes.DataLoss, "BidirectionalStreaming: failed to get metadata PeerId")
	}

	peer := session.NewPeerLocal(peermd[0])

	// create ion sfu peer
	p := ionsfu.NewPeer(sfuServer)
	peer.Peer(p)
	defer p.Close()

	datach := make(chan *corespb.Request, 1)
	go func(ss corespb.Service_BidirectionalStreamingServer, dataChan chan *corespb.Request) {
		for {
			r, err := ss.Recv()
			if err != nil {
				return
			}

			dataChan <- r
		}
	}(stream, datach)

	for {
		select {
		case frame, ok := <-datach:
			if !ok {
				return status.Errorf(codes.DataLoss, "BidirectionalStreaming: failed to get data")
			}

			fn, err := router.Fn(coresproto.Command(frame.Command))
			if err != nil {
				return err
			}

			resp, err := fn(peer, frame)
			if err != nil {
				return err
			}

			if resp != nil {
				stream.Send(resp)
			}

		case w, ok := <-peer.DataChannel():
			if !ok {
				return status.Errorf(codes.DataLoss, "BidirectionalStreaming: failed to get data")
			}

			if w != nil {
				switch p := w.Payload.(type) {
				case *corespb.Response_Content:
					fmt.Println("send to :", w.Command, string(p.Content))
				}

				stream.Send(w)
			}
		}
	}
}

func NewServerActor(config *conf.RPC) (*os.ProcessActor, error) {
	lis, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return nil, err
	}

	s, err := rpc.NewServer(config)
	if err != nil {
		return nil, err
	}

	service.RegisterChannelzServiceToServer(s)
	corespb.RegisterServiceServer(s, &baseserver{})
	return &os.ProcessActor{
		Exec: func() error {
			Logger().Log("transport", "rpc", "on", config.Addr)
			return s.Serve(lis)
		},
		Interrupt: func(err error) {
			if err == nil {
				return
			}

			log.Error(Logger()).Log("transport", "rpc", "error", err)
		},

		Close: func() {
			Logger().Log("transport", "rpc", "on", "shutdown")
			s.Stop()
		},
	}, nil
}
