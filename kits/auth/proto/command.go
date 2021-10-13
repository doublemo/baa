package proto

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	// LoginCommand 登录
	LoginCommand coresproto.Command = 1

	// RegisterCommand 注册
	RegisterCommand coresproto.Command = 2

	// LogoutCommand 退出登录
	LogoutCommand coresproto.Command = 3

	// OfflineCommand 玩家离线
	OfflineCommand coresproto.Command = 4
)
