package proto

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	// SnowflakeCommand 获取雪花ID
	SnowflakeCommand coresproto.Command = 1

	// AutoincrementCommand 自增ID
	AutoincrementCommand coresproto.Command = 2

	// ClearAutoincrementCommand 清除自增ID
	ClearAutoincrementCommand coresproto.Command = 3

	// MoreAutoincrementCommand 多K同时获取自增ID
	MoreAutoincrementCommand coresproto.Command = 4
)
