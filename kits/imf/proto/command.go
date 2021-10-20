package proto

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	// ReloadCommand 重新加载
	ReloadCommand coresproto.Command = 1

	// DirtyWordsCommand 脏词操作
	DirtyWordsCommand coresproto.Command = 2

	// CheckCommand 检查内容
	CheckCommand coresproto.Command = 3
)
