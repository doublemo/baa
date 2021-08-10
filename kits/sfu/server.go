package sfu

import (
	"context"
	"fmt"
	"io"
	"net"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/kits/sfu/adapter/router"
	"github.com/doublemo/baa/kits/sfu/proto"
	sfulog "github.com/pion/ion-sfu/pkg/logger"
	"github.com/pion/ion-sfu/pkg/sfu"
	ionsfu "github.com/pion/ion-sfu/pkg/sfu"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ionsfuServer *sfu.SFU
)

type (

	// Configuration k
	Configuration struct {
		Ballast   int64               `alias:"ballast"`
		WithStats bool                `alias:"withstats"`
		WebRTC    ionsfu.WebRTCConfig `alias:"webrtc"`
		Router    ionsfu.RouterConfig `alias:"router"`
		Turn      ionsfu.TurnConfig   `alias:"turn"`
	}

	baseserver struct {
		corespb.UnimplementedServiceServer
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

	peerId, ok := md["peerid"]
	if !ok {
		return status.Errorf(codes.DataLoss, "BidirectionalStreaming: failed to get metadata PeerId")
	}

	fmt.Println("peerId:", peerId)

	var (
		resp *corespb.Response
		errr error
	)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if req.Command == proto.JoinCommand.Int32() {
			resp, errr = join(stream, peerId[0], req)
		} else {
			fn, err := router.Fn(coresproto.Command(req.Command))
			if err != nil {
				return err
			}
			resp, errr = fn(req)
		}

		if errr != nil {
			return errr
		}

		if resp == nil {
			continue
		}

		stream.Send(resp)
	}
}

func NewServerActor(config *conf.RPC, etcd *conf.Etcd, sfuconfig *Configuration) (*os.ProcessActor, error) {
	var c ionsfu.Config
	{
		c.SFU.Ballast = sfuconfig.Ballast
		c.SFU.WithStats = sfuconfig.WithStats
		c.WebRTC = sfuconfig.WebRTC
		c.Router = sfuconfig.Router
		c.Turn = sfuconfig.Turn
	}
	ionsfu.Logger = sfulog.New()
	ionsfuServer = ionsfu.NewSFU(c)

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
