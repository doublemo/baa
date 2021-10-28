package command

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/kit"
)

const (
	// IMSend 发送聊天信息
	IMSend coresproto.Command = kit.IM + iota + 1

	// IMPull 拉取聊天信息
	IMPull

	// IMPush 推送聊天聊天信息
	IMPush

	// IMAction 信息事件
	IMAction
)
