package errcode

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/types"
)

var (
	// ErrMaxIDNumber 请求ID数量过大
	ErrMaxIDNumber = types.NewErrCode(1, "Invalid call number in request")

	// ErrKeyIsEmpty redis key等于空
	ErrKeyIsEmpty = types.NewErrCode(1, "key is empty")

	// ErrInternalServer 服务器内部错误
	ErrInternalServer = types.NewErrCode(500, "Internal Server Error")
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
