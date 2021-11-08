package errcode

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/types"
	"github.com/doublemo/baa/internal/proto/kit"
)

var (
	// ErrInvalidCharacter 包含非法字符
	ErrInvalidCharacter = types.NewErrCode(kit.MakeErrCode(kit.IM, 1), "Illegal character")

	// ErrInvalidUserIdToken 无效用户ID
	ErrInvalidUserIdToken = types.NewErrCode(kit.MakeErrCode(kit.IM, 2), "Invalid user id token in request")

	// ErrInvalidTopicToken 无效用户ID
	ErrInvalidTopicToken = types.NewErrCode(kit.MakeErrCode(kit.IM, 3), "Invalid topic token in request")

	// ErrInvalidUserStatus 无法获取用户实时状态
	ErrInvalidUserStatus = types.NewErrCode(kit.MakeErrCode(kit.IM, 4), "User status couldn't be created")

	// ErrNotFriend 还不是好朋友
	ErrNotFriend = types.NewErrCode(kit.MakeErrCode(kit.IM, 5), "You haven’t established a friendship yet, so you can’t trust each other yet")

	// ErrInternalServer 服务器内部错误
	ErrInternalServer = types.NewErrCode(kit.MakeErrCode(kit.IM, 500), "Internal Server Error")
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
