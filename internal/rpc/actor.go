package rpc

import (
	"net"

	coreslog "github.com/doublemo/baa/cores/log"
	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// NewRPCServerActor 创建RPC服务
func NewRPCServerActor(config conf.RPC, srv corespb.ServiceServer, logger coreslog.Logger) (*os.ProcessActor, error) {
	lis, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return nil, err
	}

	s, err := NewServer(config)
	if err != nil {
		return nil, err
	}

	corespb.RegisterServiceServer(s, srv)
	service.RegisterChannelzServiceToServer(s)
	healthpb.RegisterHealthServer(s, health.NewServer())
	grpc_prometheus.Register(s)
	return &os.ProcessActor{
		Exec: func() error {
			logger.Log("transport", "rpc", "on", config.Addr)
			return s.Serve(lis)
		},
		Interrupt: func(err error) {
			if err == nil {
				return
			}

			log.Error(logger).Log("transport", "rpc", "error", err)
		},

		Close: func() {
			logger.Log("transport", "rpc", "on", "shutdown")
			s.Stop()
		},
	}, nil
}
