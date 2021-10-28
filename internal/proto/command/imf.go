package command

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/kit"
)

const (
	// IMFReload 重新加载
	IMFReload coresproto.Command = kit.IMF + iota + 1

	// IMFDirtyWords 脏词操作
	IMFDirtyWords

	// IMFCheck 检查内容
	IMFCheck
)
