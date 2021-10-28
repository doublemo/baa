package command

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/kit"
)

const (
	// SFUNegotiate sfu沟通
	SFUNegotiate coresproto.Command = kit.SFU + iota + 1

	// SFUJoin 加入
	SFUJoin
)
