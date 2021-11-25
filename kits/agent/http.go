package agent

import (
	"context"
	"net/http"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/internal/conf"
	"github.com/gorilla/mux"
)

// NewHttpProcessActor 创建http
func NewHttpProcessActor(httpConfig conf.Http, routerConfig RouterConfig) (*os.ProcessActor, error) {
	if httpConfig.Addr == "" {
		return &os.ProcessActor{}, nil
	}

	r := mux.NewRouter()
	// 跨域支持
	if httpConfig.CORS {
		r.Use(mux.CORSMethodMiddleware(r), func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				if req.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				next.ServeHTTP(w, req)
			})
		})
	}

	// 路由
	httpRouterV1(r, routerConfig.HttpConfigV1)

	s := &http.Server{
		Addr:           httpConfig.Addr,
		Handler:        r,
		ReadTimeout:    time.Duration(httpConfig.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(httpConfig.WriteTimeout) * time.Second,
		MaxHeaderBytes: httpConfig.MaxHeaderBytes,
	}

	return &os.ProcessActor{
		Exec: func() error {
			Logger().Log("transport", "http", "on", httpConfig.Addr, "ssl", httpConfig.SSL)
			if httpConfig.SSL {
				return s.ListenAndServeTLS(httpConfig.Cert, httpConfig.Key)
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

			log.Error(Logger()).Log("transport", "websocket", "error", err)
		},

		Close: func() {
			s.Shutdown(context.Background())
			Logger().Log("transport", "http", "on", "shutdown")
		},
	}, nil
}
