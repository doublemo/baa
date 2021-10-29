package kit

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	Agent coresproto.Command = 1 + (iota * 1000)
	SFU
	Auth
	SNID
	IM
	IMF
	USRT
	User
	Robot
)

// MakeErrCode 为不同组件创建对应的错误代码
// * <pre>{@code
// * +--------------------------+------------+
// * |     kit id    | sequence |   extends  |
// * +--------------------------+------------+
// *       16bits       13bits       3bits
// * }</pre>
//  kit -32767 ~ 32767
//  seq 0 ~ 4096
//  extends ：0 ~ 3
func MakeErrCode(cmd coresproto.Command, seq uint32, extends ...uint32) int32 {
	if len(extends) < 1 {
		extends = []uint32{0}
	}
	return (int32(cmd) << uint(16)) | (int32(seq) << uint(3)) | int32(extends[0])
}

// ParseErrCode 解析errcode
func ParseErrCode(code int32) (int32, uint32, uint32) {
	extends := (code << uint(29)) >> uint(29)
	sequence := (code << uint(16)) >> uint(19)
	kit := code >> 16
	return kit, uint32(sequence), uint32(extends)
}
