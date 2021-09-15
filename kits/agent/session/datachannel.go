package session

import (
	kitlog "github.com/doublemo/baa/cores/log/level"
	awebrtc "github.com/doublemo/baa/kits/agent/webrtc"
	"github.com/pion/webrtc/v3"
)

// DataChannel 数据通道
type DataChannel struct {
	pc         *webrtc.PeerConnection
	candidates []webrtc.ICECandidateInit
	peer       Peer
}

// OnTrack call
func (dc *DataChannel) OnTrack(f func(*webrtc.TrackRemote, *webrtc.RTPReceiver)) {
	dc.pc.OnTrack(f)
}

// OnDataChannel call
func (dc *DataChannel) OnDataChannel(f func(*webrtc.DataChannel)) {
	dc.pc.OnDataChannel(f)
}

// OnICEConnectionStateChange call
func (dc *DataChannel) OnICEConnectionStateChange(f func(webrtc.ICEConnectionState)) {
	dc.pc.OnICEConnectionStateChange(f)
}

// Answer get/set Description
func (dc *DataChannel) Answer(offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	if err := dc.pc.SetRemoteDescription(offer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	if dc.candidates != nil {
		for _, candidate := range dc.candidates {
			if err := dc.pc.AddICECandidate(candidate); err != nil {
				kitlog.Error(Logger()).Log("Add publisher ice candidate to peer err:", err)
			}
		}
	}

	dc.candidates = nil
	answer, err := dc.pc.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	if err := dc.pc.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	return answer, nil
}

// AddICECandidate add/set
func (dc *DataChannel) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	if dc.pc.RemoteDescription() != nil {
		return dc.pc.AddICECandidate(candidate)
	}

	if dc.candidates == nil {
		dc.candidates = []webrtc.ICECandidateInit{candidate}
	} else {
		dc.candidates = append(dc.candidates, candidate)
	}
	return nil
}

// NewDataChannel 创建数据通道
func NewDataChannel(peer Peer, w awebrtc.WebRTCTransportConfig) (*DataChannel, error) {
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&webrtc.MediaEngine{}), webrtc.WithSettingEngine(w.Setting))
	pc, err := api.NewPeerConnection(w.Configuration)
	if err != nil {
		return nil, err
	}

	return &DataChannel{
		pc: pc,
	}, nil
}
