package command

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/kit"
)

const (
	// SNIDSnowflake 获取雪花ID
	SNIDSnowflake coresproto.Command = kit.SNID + iota + 1

	// SNIDAutoincrement 自增ID
	SNIDAutoincrement

	// SNIDClearAutoincrement 清除自增ID
	SNIDClearAutoincrement

	// SNIDMoreAutoincrement 多K同时获取自增ID
	SNIDMoreAutoincrement
)
