package errcode

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/types"
	"github.com/doublemo/baa/internal/proto/kit"
)

var (
	// ErrInvalidAccountID 错误的账户ID
	ErrInvalidAccountID = types.NewErrCode(kit.MakeErrCode(kit.User, 1), "Account ID is incorrect")

	// ErrNickNameLettersInvalid 非法账户名称
	ErrNickNameLettersInvalid = types.NewErrCode(kit.MakeErrCode(kit.User, 2), "The name contains illegal characters. The account name can only contain: A-Z 0-9 and Chinese characters. Special symbols support:@ . _")

	// ErrPhoneNumberInvalid 非法的手机号码
	ErrPhoneNumberInvalid = types.NewErrCode(kit.MakeErrCode(kit.User, 3), "Phone number is invalid")

	// ErrUserNotfound 用户不存在
	ErrUserNotfound = types.NewErrCode(kit.MakeErrCode(kit.User, 4), "User is notfound")

	// ErrUserAlreadyContact 已经是联系人
	ErrUserAlreadyContact = types.NewErrCode(kit.MakeErrCode(kit.User, 5), "Contact is already in your address book")

	// ErrMessageTooLong 信息太长
	ErrMessageTooLong = types.NewErrCode(kit.MakeErrCode(kit.User, 6), "Your message is too long")

	// ErrContactsRequestExpired 增加联系人请求已经过期
	ErrContactsRequestExpired = types.NewErrCode(kit.MakeErrCode(kit.User, 7), "The request to add a contact has expired")

	// ErrContactsRequestNotFound 增加联系人请求不存在
	ErrContactsRequestNotFound = types.NewErrCode(kit.MakeErrCode(kit.User, 8), "The request to add a contact has deleted")

	// ErrInternalServer 服务器内部错误
	ErrInternalServer = types.NewErrCode(kit.MakeErrCode(kit.User, 500), "Internal Server Error")
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
