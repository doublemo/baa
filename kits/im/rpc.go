package im

import (
	"net"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// NewRPCServerActor 创建RPC服务
func NewRPCServerActor(config conf.RPC) (*os.ProcessActor, error) {
	lis, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return nil, err
	}

	s, err := rpc.NewServer(config)
	if err != nil {
		return nil, err
	}

	healthcheck := health.NewServer()
	service.RegisterChannelzServiceToServer(s)
	healthpb.RegisterHealthServer(s, healthcheck)
	corespb.RegisterServiceServer(s, &baseServer{})
	stopCheckChan := make(chan struct{})
	go func(name string, s *health.Server, exit chan struct{}) {
		// asynchronously inspect dependencies and toggle serving status as needed
		next := healthpb.HealthCheckResponse_SERVING
		timer := time.NewTicker(time.Second * 5)
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				healthcheck.SetServingStatus(name, next)
				if next == healthpb.HealthCheckResponse_SERVING {
					next = healthpb.HealthCheckResponse_NOT_SERVING
				} else {
					next = healthpb.HealthCheckResponse_SERVING
				}

			case <-exit:
				return
			}
		}

	}(config.Name, healthcheck, stopCheckChan)

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
			close(stopCheckChan)
		},
	}, nil
}
