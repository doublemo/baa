package proto

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	Agent coresproto.Command = 1
	SFU   coresproto.Command = 2
)

const (
	// HandshakeCommand 加密握手
	HandshakeCommand coresproto.Command = 1

	//DatachannelCommand 创建数据通道
	DatachannelCommand coresproto.Command = 2
)
