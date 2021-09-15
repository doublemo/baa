package sfu

import (
	sfulog "github.com/doublemo/baa/kits/sfu/pkg/logger"
	"github.com/doublemo/baa/kits/sfu/pkg/middlewares/datachannel"
	ionsfu "github.com/doublemo/baa/kits/sfu/pkg/sfu"
)

type (

	// Configuration sfu config
	Configuration struct {
		Ballast   int64               `alias:"ballast"`
		WithStats bool                `alias:"withstats"`
		WebRTC    ionsfu.WebRTCConfig `alias:"webrtc"`
		Router    ionsfu.RouterConfig `alias:"router"`
		Turn      ionsfu.TurnConfig   `alias:"turn"`
	}
)

// newSFUServer 创建sfu 服务器
func newSFUServer(config *Configuration) *ionsfu.SFU {
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
	return ionsfuServer
}
