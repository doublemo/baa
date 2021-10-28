package command

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/kit"
)

const (
	// USRTUpdateUserStatus 更新用户状态
	USRTUpdateUserStatus coresproto.Command = kit.USRT + iota + 1

	// USRTDeleteUserStatus 删除用户状态
	USRTDeleteUserStatus

	// USRTGetUserStatus 获取用户在线信息
	USRTGetUserStatus
)
