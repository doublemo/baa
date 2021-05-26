// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package os

// Command 定义服务命令
type Command string

const (
	// CommandStop 停止服务
	CommandStop Command = "stop"

	// CommandQuit 退出服务
	CommandQuit Command = "quit"

	// CommandReload 重载服务
	CommandReload Command = "reload"

	// CommandUSR1 自定义命令1
	CommandUSR1 Command = "usr1"

	// CommandUSR2 自定义命令2
	CommandUSR2 Command = "usr2"
)
