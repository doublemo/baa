package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"

	coreslog "github.com/doublemo/baa/cores/log"
	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config 指标监控服务
type Config struct {
	Addr   string `alias:"addr" default:":6070"`
	TurnOn bool   `alias:"turnOn" default:"false"`
}

func NewMetricsProcessActor(c Config, logger coreslog.Logger) (*os.ProcessActor, error) {
	if !c.TurnOn {
		return &os.ProcessActor{}, nil
	}
	fmt.Println(c)
	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Handler: m}
	metricsLis, err := net.Listen("tcp", c.Addr)
	if err != nil {
		return nil, fmt.Errorf("cannot bind to metrics endpoint, add:%s, error:%s", c.Addr, err.Error())
	}

	return &os.ProcessActor{
		Exec: func() error {
			logger.Log("transport", "metrics", "on", c.Addr)
			err = srv.Serve(metricsLis)
			if err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("debug server stopped. got err: %s", err.Error())
			}
			return nil
		},
		Interrupt: func(err error) {
			if err == nil {
				return
			}
			log.Error(logger).Log("transport", "metrics", "error", err)
		},

		Close: func() {
			logger.Log("transport", "metrics", "on", "shutdown", "error", srv.Shutdown(context.Background()))
		},
	}, nil
}
