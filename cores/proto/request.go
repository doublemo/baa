// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package proto

// Request 定义请求协议
// ----------------------------------------------------------------
// |  V   |  SID   | Command | SubCommand |  Payload  |
// | int8 | uint32 |  int16  |    int16   |   bytes   |
// ----------------------------------------------------------------
type Request interface {
	// V 版本号
	V() int8

	// SID 协议编号
	SID() uint32

	// Command 主命令
	Command() Command

	// SubCommand 子命令
	SubCommand() Command

	// Body 内容
	Body() []byte

	// Marshal 组包
	Marshal() ([]byte, error)

	// Unmarshal 解包
	Unmarshal([]byte) error
}
