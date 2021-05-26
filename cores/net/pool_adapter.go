// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package net

// PoolAdapter 池连接适配器
type PoolAdapter interface {
	// Close 连接关闭方法
	Close()

	// Ok 确认连接是否有效
	Ok() bool
}
