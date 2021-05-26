// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package os

// Server 服务接口
type Server interface {
	// Start 启动服务
	Start() error

	// Readyed 服务是否已经准备就绪
	Readyed() bool

	// Shutdown 关闭服务
	Shutdown()

	// Reload 重新加载服务
	Reload()

	// ServiceName 返回服务名称
	ServiceName() string

	// OtherCommand 响应其他自定义命令
	OtherCommand(int)

	// QuitCh 退出信息号
	QuitCh() <-chan struct{}
}
