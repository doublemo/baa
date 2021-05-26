// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/doublemo/baa/cores/log"
	kitlog "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/kits/kun/mem"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewServerHttp 创建http服务
func NewServerHttp(o *mem.Parameters, router func(*gin.Engine), logger log.Logger) (*os.ProcessActor, error) {
	if o.Http == nil {
		return nil, errors.New("Invalid parameters in http options")
	}

	gin.SetMode(gin.ReleaseMode)

	// Disable Console Color
	gin.DisableConsoleColor()

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router(r)
	s := &http.Server{
		Addr:           o.Http.Addr,
		Handler:        r,
		ReadTimeout:    time.Duration(o.Http.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(o.Http.WriteTimeout) * time.Second,
		MaxHeaderBytes: o.Http.MaxHeaderBytes,
	}

	return &os.ProcessActor{
		Exec: func() error {
			logger.Log("transport", "http", "on", o.Http.Addr, "ssl", o.Http.SSL)
			if o.Http.SSL {
				return s.ListenAndServeTLS(o.Http.Cert, o.Http.Key)
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

			kitlog.Error(logger).Log("transport", "http", "error", err)
		},

		Close: func() {
			logger.Log("transport", "http", "on", "shutdown")
			s.Shutdown(context.Background())
		},
	}, nil
}
