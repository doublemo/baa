package agent

import (
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	coresnet "github.com/doublemo/baa/cores/net"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/internal/conf"
	midPeer "github.com/doublemo/baa/kits/agent/middlewares/peer"
	"github.com/doublemo/baa/kits/agent/session"
	awebrtc "github.com/doublemo/baa/kits/agent/webrtc"
	"github.com/pion/webrtc/v3"
)

// NewSocketProcessActor 创建socket服务
func NewSocketProcessActor(config *conf.Scoket) (*os.ProcessActor, error) {
	var wg sync.WaitGroup
	exitChan := make(chan struct{})
	s := coresnet.NewSocket2()
	s.OnConnect(func(conn net.Conn) {
		wg.Add(1)
		peer := session.NewPeerSocket(conn, time.Duration(config.ReadDeadline)*time.Second, time.Duration(config.WriteDeadline)*time.Second, exitChan)
		peer.OnReceive(func(p session.Peer, m session.PeerMessagePayload) error {
			resp, err := handleBinaryMessage(p, m.Data)
			if resp != nil {
				bytes, err := resp.Marshal()
				if err != nil {
					return err
				}
				p.Send(session.PeerMessagePayload{Data: bytes})
			}
			return err
		})

		peer.OnClose(func(p session.Peer) {
			wg.Done()
			session.RemovePeer(p)
		})

		peer.Use(midPeer.NewRPMLimiter(config.RPMLimit, Logger()))
		session.AddPeer(peer)

		dc, err := session.NewDataChannel(peer, awebrtc.Transport())
		if err != nil {
			log.Error(Logger()).Log("error", err)
			peer.Close()
			return
		}

		dc.OnDataChannel(func(dc *webrtc.DataChannel) {
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				fmt.Println("ddc-----", string(msg.Data))
				dc.SendText("dd-lalalal")
			})
		})

		dc.OnICEConnectionStateChange(makeICEConnectionStateChange(peer))
		peer.UseDataChannel(dc)
	})

	s.OnClose(func() {
		close(exitChan)
		wg.Wait()
	})

	return &os.ProcessActor{
		Exec: func() error {
			Logger().Log("transport", "socket", "on", config.Addr)

			return s.Serve(config.Addr, config.ReadBufferSize, config.WriteBufferSize)
		},
		Interrupt: func(err error) {
			if err == nil {
				return
			}
			log.Error(Logger()).Log("transport", "socket", "error", err)
		},

		Close: func() {
			s.Shutdown()
			Logger().Log("transport", "socket", "on", "shutdown")
		},
	}, nil
}
