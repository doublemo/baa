// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

// +build !windows

package os

// runService 运行服务
func runService(s Server) error {
	return s.Start()
}

// IsWindowsService 确认是否已windows服务方式进行运行
func IsWindowsService() bool {
	return false
}
