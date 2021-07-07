package proto

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	// HandshakeCommand 加密握手
	HandshakeCommand coresproto.Command = 1

	// SFUCommand sfu
	SFUCommand coresproto.Command = 2
)
