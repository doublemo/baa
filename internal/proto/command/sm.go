package command

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/kit"
)

const (
	// SMEvent 状态事件
	SMEvent coresproto.Command = kit.SM + iota + 1

	// SMUserStatus 获取用户状态
	SMUserStatus

	// SMBroadcastMessagesToAgent 广播消息到网关
	SMBroadcastMessagesToAgent

	// SMUserServers 获取用户分配的服务器
	SMUserServers

	// SMAssginServers 分配指定服务器
	SMAssginServers
)
