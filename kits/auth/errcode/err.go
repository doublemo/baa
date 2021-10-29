package errcode

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/types"
	"github.com/doublemo/baa/internal/proto/kit"
)

var (
	// ErrCommandInvalid 非法的调用命令号
	ErrCommandInvalid = types.NewErrCode(kit.MakeErrCode(kit.Auth, 1), "Invalid call command code in request")

	// ErrPhoneNumberInvalid 非法的手机号码
	ErrPhoneNumberInvalid = types.NewErrCode(kit.MakeErrCode(kit.Auth, 2), "Phone number is invalid")

	// ErrVerificationCodeIncorrect 错误的验证码
	ErrVerificationCodeIncorrect = types.NewErrCode(kit.MakeErrCode(kit.Auth, 3), "Verification code is incorrect")

	// ErrVerificationCodeExists 电话验证码已经存在
	ErrVerificationCodeExists = types.NewErrCode(kit.MakeErrCode(kit.Auth, 4), "Verification code is already sended")

	// ErrVerificationCodeCouldntCreated 电话验证码无法创建
	ErrVerificationCodeCouldntCreated = types.NewErrCode(kit.MakeErrCode(kit.Auth, 5), "Phone verification couldn't be created")

	// ErrAccountExpired 账户已经过期
	ErrAccountExpired = types.NewErrCode(kit.MakeErrCode(kit.Auth, 6), "The Account is expired")

	// ErrAccountDisabled 账户已经被禁用
	ErrAccountDisabled = types.NewErrCode(kit.MakeErrCode(kit.Auth, 7), "The Account is disabled")

	// ErrAccountIDInvalid 非法账户ID
	ErrAccountIDInvalid = types.NewErrCode(kit.MakeErrCode(kit.Auth, 8), "Account ID is invalid")

	// ErrAccountNameLettersInvalid 非法账户名称
	ErrAccountNameLettersInvalid = types.NewErrCode(kit.MakeErrCode(kit.Auth, 9), "The name contains illegal characters. The account name can only contain: A-Z 0-9 and Chinese characters. Special symbols support:@ . _")

	// ErrPasswordLettersInvalid 密码字符无效
	ErrPasswordLettersInvalid = types.NewErrCode(kit.MakeErrCode(kit.Auth, 10), "The password must contain uppercase and lowercase letters, numbers or punctuation, and must be %d-%d digits long.")

	// ErrAccountIsExists 账户名称已经存在
	ErrAccountIsExists = types.NewErrCode(kit.MakeErrCode(kit.Auth, 11), "The Account is exist")

	// ErrUsernameOrPasswordIncorrect 账户名或密码错误
	ErrUsernameOrPasswordIncorrect = types.NewErrCode(kit.MakeErrCode(kit.Auth, 12), "Username or Password is incorrect")

	// ErrInternalServer 服务器内部错误
	ErrInternalServer = types.NewErrCode(kit.MakeErrCode(kit.Auth, 500), "Internal Server Error")
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
