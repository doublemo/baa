package agent

import (
	"context"
	"net/http"
	"sync"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/internal/conf"
	midPeer "github.com/doublemo/baa/kits/agent/middlewares/peer"
	"github.com/doublemo/baa/kits/agent/session"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// NewWebsocketProcessActor 创建Websocket
func NewWebsocketProcessActor(config *conf.Webscoket) (*os.ProcessActor, error) {
	r := mux.NewRouter()
	var webSocketUpgrader websocket.Upgrader
	{
		webSocketUpgrader = websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}

	var wg sync.WaitGroup
	exitChan := make(chan struct{})
	r.HandleFunc("/websocket", func(rw http.ResponseWriter, r *http.Request) {
		serveWebsocket(rw, r, webSocketUpgrader, config, &wg, exitChan)
	}).Methods("GET")

	s := &http.Server{
		Addr:           config.Addr,
		Handler:        r,
		ReadTimeout:    time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(config.WriteTimeout) * time.Second,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	return &os.ProcessActor{
		Exec: func() error {
			Logger().Log("transport", "websocket", "on", config.Addr, "ssl", config.SSL)
			if config.SSL {
				return s.ListenAndServeTLS(config.Cert, config.Key)
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
			close(exitChan)
			wg.Wait()
			s.Shutdown(context.Background())
			Logger().Log("transport", "websocket", "on", "shutdown")
		},
	}, nil
}

func serveWebsocket(w http.ResponseWriter, req *http.Request, upgrader websocket.Upgrader, config *conf.Webscoket, wg *sync.WaitGroup, exitChan chan struct{}) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Error(Logger()).Log("error", err)
		return
	}

	wg.Add(1)
	peer := session.NewPeerWebsocket(conn, time.Duration(config.ReadDeadline)*time.Second, time.Duration(config.WriteDeadline)*time.Second, config.MaxMessageSize, exitChan)
	peer.OnReceive(onMessage)
	peer.OnClose(func(p session.Peer) {
		wg.Done()
		session.RemovePeer(p)
		sRouter.Destroy(p)
	})

	peer.Use(midPeer.NewRPMLimiter(config.RPMLimit, Logger()))
	peer.Go()
	session.AddPeer(peer)

	// bind datachannel
	if err := useDataChannel(peer); err != nil {
		log.Error(Logger()).Log("error", err)
		peer.Close()
		return
	}
}
