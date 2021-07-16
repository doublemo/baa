package proto

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	// NegotiateCommand sfu沟通
	NegotiateCommand coresproto.Command = 1

	// JoinCommand 加入
	JoinCommand coresproto.Command = 2
)
