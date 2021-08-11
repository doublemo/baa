package session

import (
	"errors"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	ionsfu "github.com/pion/ion-sfu/pkg/sfu"
)

type (
	Peer interface {
		ID() string
		Send(*corespb.Response) error
		Peer(...ionsfu.Peer) ionsfu.Peer
		DataChannel() <-chan *corespb.Response
	}

	PeerLocal struct {
		id       string
		sfuPeer  ionsfu.Peer
		dataChan chan *corespb.Response
	}
)

func (peer *PeerLocal) ID() string {
	return peer.id
}

func (peer *PeerLocal) Send(w *corespb.Response) error {
	select {
	case peer.dataChan <- w:
	default:
		return errors.New("peer datachannel fulled")
	}

	return nil
}

func (peer *PeerLocal) Peer(p ...ionsfu.Peer) ionsfu.Peer {
	if len(p) < 1 {
		return peer.sfuPeer
	}

	peer.sfuPeer = p[0]
	return p[0]
}

func (peer *PeerLocal) DataChannel() <-chan *corespb.Response {
	return peer.dataChan
}

func NewPeerLocal(id string) *PeerLocal {
	return &PeerLocal{
		id:       id,
		dataChan: make(chan *corespb.Response, 1),
	}
}
