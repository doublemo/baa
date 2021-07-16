package session

import (
	"github.com/pion/webrtc/v3"
)

type (
	Peer interface {
		ID() string
	}

	PeerLocal struct {
		id             string
		pc             *webrtc.PeerConnection
		onOffer        func(*webrtc.SessionDescription)
		onIceCandidate func(*webrtc.ICECandidateInit, int)
		onNotify       func(payload []byte) error
	}
)

func (peer *PeerLocal) ID() string {
	return peer.id
}

func (peer *PeerLocal) OnOffer(fn func(*webrtc.SessionDescription)) {
	peer.onOffer = fn
}

func (peer *PeerLocal) OnIceCandidate(fn func(*webrtc.ICECandidateInit, int)) {
	peer.onIceCandidate = fn
}

func (peer *PeerLocal) OnNotify(fn func([]byte) error) {
	peer.onNotify = fn
}

func NewPeerLocal(id string) *PeerLocal {
	return &PeerLocal{
		id: id,
	}
}
