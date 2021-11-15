package session

import (
	"sync"

	kitlog "github.com/doublemo/baa/cores/log/level"
	"github.com/pion/webrtc/v3"
)

// DataChannel 数据通道
type DataChannel struct {
	pc         *webrtc.PeerConnection
	dcs        []*webrtc.DataChannel
	candidates []webrtc.ICECandidateInit
	mutex      sync.RWMutex
}

// OnICEConnectionStateChange call
func (dc *DataChannel) OnICEConnectionStateChange(f func(webrtc.ICEConnectionState)) {
	dc.pc.OnICEConnectionStateChange(f)
}

// AddDataChannel add
func (dc *DataChannel) AddDataChannel(d *webrtc.DataChannel) {
	dc.mutex.Lock()
	dc.dcs = append(dc.dcs, d)
	dc.mutex.Unlock()
}

// RemoveDataChannel remove
func (dc *DataChannel) RemoveDataChannel(label string) {
	dcs := make([]*webrtc.DataChannel, 0)
	dc.mutex.RLock()
	for _, d := range dc.dcs {
		if d.Label() == label {
			continue
		}

		dcs = append(dcs, d)
	}
	dc.mutex.RUnlock()

	dc.mutex.Lock()
	dc.dcs = dcs
	dc.mutex.Unlock()
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

// Close 关闭
func (dc *DataChannel) Close() error {
	return dc.pc.Close()
}

// Send 发送信息到数据通道
func (dc *DataChannel) Send(msg []byte) error {
	dc.mutex.RLock()
	for _, d := range dc.dcs {
		dc.mutex.RUnlock()
		if err := d.Send(msg); err != nil {
			return err
		}
		dc.mutex.RLock()
	}
	dc.mutex.RUnlock()
	return nil
}
