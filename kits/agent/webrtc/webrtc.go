package webrtc

import (
	"net"
	"time"

	"github.com/doublemo/baa/kits/agent/log"
	"github.com/pion/ice/v2"
	"github.com/pion/webrtc/v3"
)

var w WebRTCTransportConfig

type (

	// ICEServerConfig defines parameters for ice servers
	ICEServerConfig struct {
		URLs       []string `alias:"urls"`
		Username   string   `alias:"username"`
		Credential string   `alias:"credential"`
	}

	// Candidates setting
	Candidates struct {
		IceLite    bool     `alias:"icelite"`
		NAT1To1IPs []string `alias:"nat1to1"`
	}

	// WebRTCTransportConfig represents Configuration options
	WebRTCTransportConfig struct {
		Configuration webrtc.Configuration
		Setting       webrtc.SettingEngine
	}

	WebRTCTimeoutsConfig struct {
		ICEDisconnectedTimeout int `alias:"disconnected"`
		ICEFailedTimeout       int `alias:"failed"`
		ICEKeepaliveInterval   int `alias:"keepalive"`
	}

	// WebRTCConfig defines parameters for ice
	WebRTCConfig struct {
		ICESinglePort int                  `alias:"singleport"`
		ICEPortRange  []uint16             `alias:"portrange"`
		ICEServers    []ICEServerConfig    `alias:"iceserver"`
		Candidates    Candidates           `alias:"candidates"`
		SDPSemantics  string               `alias:"sdpsemantics"`
		MDNS          bool                 `alias:"mdns"`
		Timeouts      WebRTCTimeoutsConfig `alias:"timeouts"`
	}
)

func Init(c WebRTCConfig) error {
	se := webrtc.SettingEngine{}
	se.DisableMediaEngineCopy(true)

	if c.ICESinglePort != 0 {
		log.Logger().Log("transport", "ice [single-port]", "on", c.ICESinglePort)
		udpListener, err := net.ListenUDP("udp", &net.UDPAddr{
			IP:   net.IP{0, 0, 0, 0},
			Port: c.ICESinglePort,
		})

		if err != nil {
			return err
		}
		se.SetICEUDPMux(webrtc.NewICEUDPMux(nil, udpListener))
	} else {
		if len(c.ICEPortRange) == 2 && c.ICEPortRange[0] != 0 && c.ICEPortRange[1] != 0 {
			if err := se.SetEphemeralUDPPortRange(c.ICEPortRange[0], c.ICEPortRange[1]); err != nil {
				return err
			}
		}
	}

	var iceServers []webrtc.ICEServer
	if c.Candidates.IceLite {
		se.SetLite(c.Candidates.IceLite)
	} else {
		for _, iceServer := range c.ICEServers {
			s := webrtc.ICEServer{
				URLs:       iceServer.URLs,
				Username:   iceServer.Username,
				Credential: iceServer.Credential,
			}
			iceServers = append(iceServers, s)
		}
	}

	sdpSemantics := webrtc.SDPSemanticsUnifiedPlan
	switch c.SDPSemantics {
	case "unified-plan-with-fallback":
		sdpSemantics = webrtc.SDPSemanticsUnifiedPlanWithFallback
	case "plan-b":
		sdpSemantics = webrtc.SDPSemanticsPlanB
	}

	if c.Timeouts.ICEDisconnectedTimeout == 0 &&
		c.Timeouts.ICEFailedTimeout == 0 &&
		c.Timeouts.ICEKeepaliveInterval == 0 {
		log.Logger().Log("webrtc", "No webrtc timeouts found in config, using default ones")
	} else {
		se.SetICETimeouts(
			time.Duration(c.Timeouts.ICEDisconnectedTimeout)*time.Second,
			time.Duration(c.Timeouts.ICEFailedTimeout)*time.Second,
			time.Duration(c.Timeouts.ICEKeepaliveInterval)*time.Second,
		)
	}

	w.Configuration = webrtc.Configuration{
		ICEServers:   iceServers,
		SDPSemantics: sdpSemantics,
	}

	w.Setting = se
	if len(c.Candidates.NAT1To1IPs) > 0 {
		w.Setting.SetNAT1To1IPs(c.Candidates.NAT1To1IPs, webrtc.ICECandidateTypeHost)
	}

	if !c.MDNS {
		w.Setting.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
	}
	return nil
}

func Transport() WebRTCTransportConfig {
	return w
}
