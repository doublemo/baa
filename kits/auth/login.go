package auth

import (
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/auth/dao"
	"github.com/doublemo/baa/kits/auth/errcode"
	"github.com/doublemo/baa/kits/auth/proto/pb"
	snpb "github.com/doublemo/baa/kits/snid/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SMSConfig 短信验证码配置
type SMSConfig struct {
	// CodeMaxLen 短信验证码长度
	CodeMaxLen int `alias:"codeMaxLen" default:"4"`

	// CodeExpireAt 短信验证有效期 (秒)
	CodeExpireAt int `alias:"codeExpireAt" default:"300"`

	// CodeReplayAt 短信重发时间 (秒)
	CodeReplayAt int `alias:"codeReplayAt" default:"60"`

	// PasswordMinLen 密码最少字符
	PasswordMinLen int `alias:"passwordMinLen" default:"8"`

	// PasswordMaxLen 密码最大字符
	PasswordMaxLen int `alias:"passwordMaxLen" default:"16"`
}

func login(req *corespb.Request, c SMSConfig) (*corespb.Response, error) {
	frame := snpb.SNID_Request{
		N: 99,
	}

	b, _ := grpcproto.Marshal(&frame)
	resp, err := ir.Handler(&corespb.Request{Command: internalSnidRouter, Payload: b})
	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		fmt.Println(payload)
	}
	return resp, nil
}

func register(req *corespb.Request, c SMSConfig) (*corespb.Response, error) {
	var frame pb.Authentication_Form_Register
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	switch payload := frame.Payload.(type) {
	case *pb.Authentication_Form_Register_Account:
		return registerAccount(req, &frame, payload, c)

	case *pb.Authentication_Form_Register_SMS:
		return registerMobliePhoneSMSSend(req, &frame, payload, c)
	}

	return nil, errors.New("InvalidProtoType")
}

func registerMobliePhoneSMSSend(req *corespb.Request, reqFrame *pb.Authentication_Form_Register, sms *pb.Authentication_Form_Register_SMS, c SMSConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
	}

	regular := "^(1[3-9])\\d{9}$"
	reg := regexp.MustCompile(regular)
	if !reg.MatchString(sms.SMS.Phone) {
		return errcode.Bad(w, errcode.ErrPhoneNumberInvalid), nil
	}

	vcode, err := dao.GetSMSCode(sms.SMS.Phone)
	if err == nil {
		_, expire, err := dao.ParseSMSVerificationCode(vcode, c.CodeMaxLen)
		if err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		expireAt := expire - int64(c.CodeExpireAt)
		if time.Now().Sub(time.Unix(expireAt, 0)).Seconds() < float64(c.CodeReplayAt) {
			return errcode.Bad(w, errcode.ErrVerificationCodeExists), nil
		}

		dao.RemoveSMSCode(sms.SMS.Phone)
	}

	code, err := dao.GenerateSMSCode(sms.SMS.Phone, c.CodeMaxLen, time.Duration(c.CodeExpireAt)*time.Second)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	// todo sms api
	fmt.Println("code:", code)
	resp := &pb.Authentication_Form_RegisterReply{
		Scheme: reqFrame.Scheme,
		Payload: &pb.Authentication_Form_RegisterReply_SMS{
			SMS: &pb.Authentication_Form_MobilePhoneSMSCode{Phone: sms.SMS.Phone},
		},
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func registerAccount(req *corespb.Request, reqFrame *pb.Authentication_Form_Register, r *pb.Authentication_Form_Register_Account, c SMSConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
	}

	// ^[\u4e00-\u9fa5]
	reg := regexp.MustCompile(`^[a-z0-9A-Z\p{Han}]+([\.|_|@][a-z0-9A-Z\p{Han}]+)*$`)
	if !reg.MatchString(r.Account.Username) {
		return errcode.Bad(w, errcode.ErrAccountNameLettersInvalid), nil
	}

	if !isValidPassword(r.Account.Password, c.PasswordMinLen, c.PasswordMaxLen) {
		return errcode.Bad(w, errcode.ErrAccountNameLettersInvalid, fmt.Sprintf(errcode.ErrAccountNameLettersInvalid.Error(), c.PasswordMinLen, c.PasswordMaxLen)), nil
	}

	reg = regexp.MustCompile("^(1[3-9])\\d{9}$")
	if !reg.MatchString(r.Account.Phone) {
		return errcode.Bad(w, errcode.ErrPhoneNumberInvalid), nil
	}

	if len(r.Account.PhoneCode) != c.CodeMaxLen {
		return errcode.Bad(w, errcode.ErrVerificationCodeIncorrect), nil
	}

	vcode, err := dao.GetSMSCode(r.Account.Phone)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	code, expire, err := dao.ParseSMSVerificationCode(vcode, c.CodeMaxLen)
	if err != nil {
		return errcode.Bad(w, errcode.ErrVerificationCodeIncorrect, err.Error()), nil
	}

	if time.Now().Unix() > expire || code != r.Account.PhoneCode {
		return errcode.Bad(w, errcode.ErrVerificationCodeIncorrect), nil
	}

	dao.RemoveSMSCode(r.Account.Phone)
	_, err = dao.GetAccoutsBySchemeAName(reqFrame.Scheme, r.Account.Username)
	if err != gorm.ErrRecordNotFound {
		return errcode.Bad(w, errcode.ErrAccountIsExists), nil
	}

	idvalues, err := getSNID(3)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	password, err := bcrypt.GenerateFromPassword([]byte(r.Account.Password), 16)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	accounts := dao.Accounts{
		ID:      idvalues[0],
		UnionID: idvalues[1],
		UserID:  idvalues[2],
		Scheme:  "password",
		Name:    r.Account.Username,
		Secret:  string(password),
	}

	dao.DB().Create(&accounts)
	resp := &pb.Authentication_Form_RegisterReply{
		Scheme: reqFrame.Scheme,
		Payload: &pb.Authentication_Form_RegisterReply_Account{
			Account: &pb.Authentication_Form_AccountInfo{
				ID:      "accounts.ID",
				UnionID: "accounts.UnionID",
			},
		},
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func isValidPassword(str string, minLen, maxLen int) bool {
	var (
		isUpper   = false
		isLower   = false
		isNumber  = false
		isSpecial = false
	)

	if len(str) < minLen || len(str) > maxLen {
		return false
	}

	for _, s := range str {
		switch {
		case unicode.IsUpper(s):
			isUpper = true
		case unicode.IsLower(s):
			isLower = true
		case unicode.IsNumber(s):
			isNumber = true
		case unicode.IsPunct(s) || unicode.IsSymbol(s):
			isSpecial = true
		default:
		}
	}
	return (isUpper && isLower) && (isNumber || isSpecial)
}
