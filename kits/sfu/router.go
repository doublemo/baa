package sfu

import (
	"github.com/doublemo/baa/kits/sfu/adapter/router"
	"github.com/doublemo/baa/kits/sfu/proto"
)

// InitRouter init
func InitRouter() {
	router.On(proto.JoinCommand, join)
	router.On(proto.NegotiateCommand, negotiate)
}
