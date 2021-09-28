package proto

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	// SnowflakeCommand 获取雪花ID
	SnowflakeCommand coresproto.Command = 1

	// AutoincrementCommand 自增ID
	AutoincrementCommand coresproto.Command = 2
)
