package command

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/kit"
)

const (
	// RobotCreate 创建机器人
	RobotCreate coresproto.Command = kit.Robot + (iota + 1)

	// RobotStart 启动机器人
	RobotStart
)
