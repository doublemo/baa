package errcode

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/types"
)

var (
	// ErrInvalidCharacter 包含非法字符
	ErrInvalidCharacter = types.NewErrCode(50001, "Illegal character")

	// ErrInvalidUserIdToken 无效用户ID
	ErrInvalidUserIdToken = types.NewErrCode(50002, "Invalid user id token in request")

	// ErrInvalidTopicToken 无效用户ID
	ErrInvalidTopicToken = types.NewErrCode(50003, "Invalid topic token in request")

	// ErrInvalidUserStatus 无法获取用户实时状态
	ErrInvalidUserStatus = types.NewErrCode(50004, "User status couldn't be created")

	// ErrInternalServer 服务器内部错误
	ErrInternalServer = types.NewErrCode(50500, "Internal Server Error")
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
