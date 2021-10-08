package errcode

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/types"
)

var (
	// ErrCommandInvalid 非法的调用命令号
	ErrCommandInvalid = types.NewErrCode(30001, "Invalid call command code in request")

	// ErrPhoneNumberInvalid 非法的手机号码
	ErrPhoneNumberInvalid = types.NewErrCode(30002, "Phone number is invalid")

	// ErrVerificationCodeIncorrect 错误的验证码
	ErrVerificationCodeIncorrect = types.NewErrCode(30003, "Verification code is incorrect")

	// ErrVerificationCodeExists 电话验证码已经存在
	ErrVerificationCodeExists = types.NewErrCode(30004, "Verification code is already sended")

	// ErrVerificationCodeCouldntCreated 电话验证码无法创建
	ErrVerificationCodeCouldntCreated = types.NewErrCode(30005, "Phone verification couldn't be created")

	// ErrAccountExpired 账户已经过期
	ErrAccountExpired = types.NewErrCode(30006, "The Account is expired")

	// ErrAccountDisabled 账户已经被禁用
	ErrAccountDisabled = types.NewErrCode(30007, "The Account is disabled")

	// ErrAccountIDInvalid 非法账户ID
	ErrAccountIDInvalid = types.NewErrCode(30008, "Account ID is invalid")

	// ErrAccountNameLettersInvalid 非法账户名称
	ErrAccountNameLettersInvalid = types.NewErrCode(30009, "The name contains illegal characters. The account name can only contain: A-Z 0-9 and Chinese characters. Special symbols support:@ . _")

	// ErrPasswordLettersInvalid 密码字符无效
	ErrPasswordLettersInvalid = types.NewErrCode(30010, "The password must contain uppercase and lowercase letters, numbers or punctuation, and must be %d-%d digits long.")

	// ErrAccountIsExists 账户名称已经存在
	ErrAccountIsExists = types.NewErrCode(30006, "The Account is exist")

	// ErrUsernameOrPasswordIncorrect 账户名或密码错误
	ErrUsernameOrPasswordIncorrect = types.NewErrCode(30007, "Username or Password is incorrect")

	// ErrInternalServer 服务器内部错误
	ErrInternalServer = types.NewErrCode(30500, "Internal Server Error")
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
