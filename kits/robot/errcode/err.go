package errcode

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/types"
	"github.com/doublemo/baa/internal/proto/kit"
)

var (
	// ErrUsernameOrPasswordIncorrect 账户名或密码错误
	ErrUsernameOrPasswordIncorrect = types.NewErrCode(kit.MakeErrCode(kit.Robot, 1), "Username or Password is incorrect")

	// ErrAccountNameLettersInvalid 非法账户名称
	ErrAccountNameLettersInvalid = types.NewErrCode(kit.MakeErrCode(kit.Robot, 2), "The name contains illegal characters. The account name can only contain: A-Z 0-9 and Chinese characters. Special symbols support:@ . _")

	// ErrPasswordLettersInvalid 密码字符无效
	ErrPasswordLettersInvalid = types.NewErrCode(kit.MakeErrCode(kit.Robot, 3), "The password must contain uppercase and lowercase letters, numbers or punctuation, and must be %d-%d digits long.")

	// ErrRobotsIsExists 账户名称已经存在
	ErrRobotsIsExists = types.NewErrCode(kit.MakeErrCode(kit.Auth, 4), "The Robot is exist")

	// ErrRobotsTooMany 机器人太多
	ErrRobotsTooMany = types.NewErrCode(kit.MakeErrCode(kit.Auth, 5), "Too many robots")

	// ErrInternalServer 服务器内部错误
	ErrInternalServer = types.NewErrCode(kit.MakeErrCode(kit.Robot, 500), "Internal Server Error")
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
