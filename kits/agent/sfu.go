package agent

import (
	log "github.com/doublemo/baa/cores/log/level"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/router"
	"github.com/doublemo/baa/kits/agent/session"
)

func sfuHookOnReceive(s *router.Stream) {
	s.OnReceive(func(peer session.Peer, r *corespb.Response) {
		w := proto.NewResponseBytes(proto.SFU, r)
		bytes, _ := w.Marshal()
		if err := peer.Send(session.PeerMessagePayload{Data: bytes}); err != nil {
			log.Error(Logger()).Log("error", err)
		}
	})
}
