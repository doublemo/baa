package command

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/kit"
)

const (
	// AgentHandshake 加密握手
	AgentHandshake coresproto.Command = kit.Agent + (iota + 1)

	// AgentDatachannel 创建数据通道
	AgentDatachannel

	// AgentHeartbeater 心跳
	AgentHeartbeater

	// AgentKickedOut 踢掉玩家
	AgentKickedOut

	// AgentBroadcast 广播消息
	AgentBroadcast
)
