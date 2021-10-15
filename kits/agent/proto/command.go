package proto

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	Agent coresproto.Command = 1
	SFU   coresproto.Command = 2
	Auth  coresproto.Command = 3
	SNID  coresproto.Command = 4
)

const (
	// HandshakeCommand 加密握手
	HandshakeCommand coresproto.Command = 1

	//DatachannelCommand 创建数据通道
	DatachannelCommand coresproto.Command = 2

	// HeartbeaterCommand 心跳
	HeartbeaterCommand coresproto.Command = 3

	// KickedOutCommand 踢掉玩家
	KickedOutCommand coresproto.Command = 4
)
