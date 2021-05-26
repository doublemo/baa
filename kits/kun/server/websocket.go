// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package server

import (
	"context"
	"net/http"
	"time"

	"github.com/doublemo/baa/cores/log"
	kitlog "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/kits/kun/mem"
	"github.com/gin-gonic/gin"
)

// NewServerWebsocket websocket
func NewServerWebsocket(o *mem.Parameters, router func(*gin.Engine), logger log.Logger) (*os.ProcessActor, error) {
	if o.Websocket == nil {
		return nil, nil
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	router(r)
	// http server
	s := &http.Server{
		Addr:           o.Websocket.Addr,
		Handler:        r,
		ReadTimeout:    time.Duration(o.Websocket.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(o.Websocket.WriteTimeout) * time.Second,
		MaxHeaderBytes: o.Websocket.MaxHeaderBytes,
	}

	return &os.ProcessActor{
		Exec: func() error {
			logger.Log("transport", "websocket", "on", o.Websocket.Addr, "ssl", o.Websocket.SSL)
			if o.Websocket.SSL {
				return s.ListenAndServeTLS(o.Websocket.Cert, o.Websocket.Key)
			}

			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		},
		Interrupt: func(err error) {
			if err == nil {
				return
			}

			kitlog.Error(logger).Log("transport", "websocket", "error", err)
		},

		Close: func() {
			logger.Log("transport", "websocket", "on", "shutdown")
			s.Shutdown(context.Background())
		},
	}, nil
}
