package session

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

type (
	// Router peer
	Router interface {
		New()
		Go(Peer, coresproto.Request) (coresproto.Response, error)
		Close()
	}
)
