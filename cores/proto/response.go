// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package proto

// Response  定义响应协议
// ----------------------------------------------------------------
// |  CODE |  V   |  SID   | Command | SubCommand |  Payload  |
// | int16 | int8 | uint32 |  int16  |    int16   |   bytes   |
// ----------------------------------------------------------------
type Response interface {
	// StatusCode 状态码
	StatusCode() int16

	// V 版本号
	V() int8

	// SeqID 协议编号
	SeqID() uint32

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
