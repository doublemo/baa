package errcode

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/types"
	"github.com/doublemo/baa/internal/proto/kit"
)

var (
	// ErrCommandInvalid 非法的调用命令号
	ErrCommandInvalid = types.NewErrCode(kit.MakeErrCode(kit.SFU, 1), "Invalid call command code in request")

	// ErrInternalServer 服务器内部错误
	ErrInternalServer = types.NewErrCode(kit.MakeErrCode(kit.SFU, 500), "Internal Server Error")
)

// Bad 错误处理
func Bad(resp *corespb.Response, err *types.ErrCode, msg ...string) *corespb.Response {
	errmsg := err.Error()
	if len(msg) > 0 {
		errmsg = msg[0]
	}

	resp.Payload = &corespb.Response_Error{
		Error: &corespb.Error{
			Code:    err.Code(),
			Message: errmsg,
		},
	}
	return resp
}
