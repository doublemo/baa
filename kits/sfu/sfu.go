package sfu

import (
	"fmt"
	"sync"
	"time"

	sfulog "github.com/doublemo/baa/kits/sfu/pkg/logger"
	"github.com/doublemo/baa/kits/sfu/pkg/middlewares/datachannel"
	"github.com/doublemo/baa/kits/sfu/pkg/relay"
	ionsfu "github.com/doublemo/baa/kits/sfu/pkg/sfu"
	"github.com/pion/turn/v2"
	"github.com/pion/webrtc/v3"
)

var sfuServer2 *ionsfu.SFU

type (

	// Configuration sfu config
	Configuration struct {
		Ballast   int64               `alias:"ballast"`
		WithStats bool                `alias:"withstats"`
		WebRTC    ionsfu.WebRTCConfig `alias:"webrtc"`
		Router    ionsfu.RouterConfig `alias:"router"`
		Turn      ionsfu.TurnConfig   `alias:"turn"`
	}

	sfuServer struct {
		sync.RWMutex
		webrtc       ionsfu.WebRTCTransportConfig
		turn         *turn.Server
		datachannels []*ionsfu.Datachannel
		withStats    bool
	}
)

func (s *sfuServer) NewDatachannel(label string) *ionsfu.Datachannel {
	dc := &ionsfu.Datachannel{Label: label}
	s.datachannels = append(s.datachannels, dc)
	return dc
}

func (s *sfuServer) GetSession(sid string) (ionsfu.Session, ionsfu.WebRTCTransportConfig) {
	return nil, s.webrtc
}

// NewSFUServer 创建sfu 服务器
func NewSFUServer(config *Configuration) *ionsfu.SFU {
	var ionsfuConfig ionsfu.Config
	{
		ionsfuConfig.SFU.Ballast = config.Ballast
		ionsfuConfig.SFU.WithStats = config.WithStats
		ionsfuConfig.WebRTC = config.WebRTC
		ionsfuConfig.Router = config.Router
		ionsfuConfig.Turn = config.Turn
	}

	ionsfu.Logger = sfulog.New()
	ionsfuServer := ionsfu.NewSFU(ionsfuConfig)
	dc := ionsfuServer.NewDatachannel(ionsfu.APIChannelLabel)
	dc.Use(datachannel.SubscriberAPI)

	go func(srv *ionsfu.SFU) {
		timer := time.NewTicker(time.Second * 5)
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				s, _ := srv.GetSession("test")
				dcs := s.GetFanOutDataChannelLabels()
				fmt.Println("dcs", s.GetDataChannels("", "relayPeerChan"), s.RelayPeers())
				if len(dcs) > 0 {
					s.FanOutMessage("", dcs[0], webrtc.DataChannelMessage{IsString: true, Data: []byte("人这生啊")})
				}
			}
		}
	}(ionsfuServer)

	sx, w := ionsfuServer.GetSession("test")
	relayMeta := relay.PeerMeta{
		PeerID:    "realyone",
		SessionID: "test",
	}

	relayConfig := relay.PeerConfig{
		SettingEngine: w.Setting,
		ICEServers:    w.Configuration.ICEServers,
		Logger:        sfulog.New(),
	}

	peer, err := relay.NewPeer(relayMeta, &relayConfig)
	if err != nil {
		fmt.Println(err)
		return ionsfuServer
	}

	fn := func(meta relay.PeerMeta, signal []byte) ([]byte, error) {
		return sx.AddRelayPeer(meta.PeerID, signal)
	}

	// relayPeer := sfu.NewRelayPeer(peer, sx, &w)
	// vpeer, err := relayPeer.Relay()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return ionsfuServer
	// }

	peer.OnDataChannel(func(channel *webrtc.DataChannel) {
		fmt.Println("relayPeer OnDataChannel:", channel.Label())
	})

	peer.OnReady(func() {
		fmt.Println("Relay Peer OnReady")

		dc, err := peer.CreateDataChannel("relayPeerChan")
		if err != nil {
			fmt.Println(err)
			return
		}

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Println("-------we:", msg)
		})

		//sx.AddDatachannel("", dc)
	})

	if err := peer.Offer(fn); err != nil {
		fmt.Println(err)
		return ionsfuServer
	}

	return ionsfuServer
}

func SetSFUServer(s *ionsfu.SFU) {
	sfuServer2 = s
}
