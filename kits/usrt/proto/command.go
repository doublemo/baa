package proto

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	// UpdateUserStatusCommand 更新用户状态
	UpdateUserStatusCommand coresproto.Command = 1

	// DeleteUserStatusCommand 删除用户状态
	DeleteUserStatusCommand coresproto.Command = 2

	// GetUserStatusCommand 获取用户在线信息
	GetUserStatusCommand coresproto.Command = 3
)
