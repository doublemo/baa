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
)
