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
)

// MakeErrCode 为不同组件创建对应的错误代码
func MakeErrCode(cmd coresproto.Command, seq int32) int32 {
	return cmd.Int32()*10000 + seq
}
