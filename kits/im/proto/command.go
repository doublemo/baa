package proto

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	// SendCommand 发送聊天信息
	SendCommand coresproto.Command = 1

	// PullCommand 拉取聊天信息
	PullCommand coresproto.Command = 2

	// PushCommand 推送聊天聊天信息
	PushCommand coresproto.Command = 3

	// ActionCommand 信息事件
	ActionCommand coresproto.Command = 4
)
