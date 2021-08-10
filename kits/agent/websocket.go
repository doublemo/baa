package agent

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/os"
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/conf"
	midPeer "github.com/doublemo/baa/kits/agent/middlewares/peer"
	"github.com/doublemo/baa/kits/agent/session"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// NewWebsocketProcessActor 创建Websocket
func NewWebsocketProcessActor(config *conf.Webscoket) (*os.ProcessActor, error) {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

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
	r.GET("/websocket", func(ctx *gin.Context) {
		if !ctx.IsWebsocket() {
			ctx.AbortWithError(http.StatusNotFound, errors.New("404 Not found"))
			return
		}

		webscoketHandler(ctx.Writer, ctx.Request, webSocketUpgrader, config, &wg, exitChan)
	})

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

func webscoketHandler(w http.ResponseWriter, req *http.Request, upgrader websocket.Upgrader, config *conf.Webscoket, wg *sync.WaitGroup, exitChan chan struct{}) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Error(Logger()).Log("error", err)
		return
	}

	wg.Add(1)
	peer := session.NewPeerWebsocket(conn, time.Duration(config.ReadDeadline)*time.Second, time.Duration(config.WriteDeadline)*time.Second, config.MaxMessageSize, exitChan)
	peer.OnReceive(func(p session.Peer, m session.PeerMessagePayload) error {
		var (
			err  error
			resp coresproto.Response
		)

		switch m.Type {
		case websocket.BinaryMessage:
			resp, err = handleBinaryMessage(p, m.Data)

		case websocket.TextMessage:
			resp, err = handleTextMessage(p, m.Data)
		}

		if resp != nil {
			bytes, err := resp.Marshal()
			if err != nil {
				return err
			}
			p.Send(session.PeerMessagePayload{Type: m.Type, Data: bytes})
		}
		return err
	})

	peer.OnClose(func(p session.Peer) {
		wg.Done()
		session.RemovePeer(p)

		// close sfu
		if m, ok := sfuRouterBidirectionalStreamingClient(p); ok {
			m.Close()
		}
	})

	peer.Use(midPeer.NewRPMLimiter(config.RPMLimit, Logger()))
	session.AddPeer(peer)
}
