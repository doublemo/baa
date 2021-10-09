package auth

import (
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"

	"github.com/doublemo/baa/cores/crypto/id"
	log "github.com/doublemo/baa/cores/log/level"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/helper"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/agent"
	agentproto "github.com/doublemo/baa/kits/agent/proto"
	agentpb "github.com/doublemo/baa/kits/agent/proto/pb"
	"github.com/doublemo/baa/kits/auth/dao"
	"github.com/doublemo/baa/kits/auth/errcode"
	"github.com/doublemo/baa/kits/auth/nats"
	"github.com/doublemo/baa/kits/auth/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type (
	// SMSConfig 短信验证码配置
	SMSConfig struct {
		// CodeMaxLen 短信验证码长度
		CodeMaxLen int `alias:"codeMaxLen" default:"4"`

		// CodeExpireAt 短信验证有效期 (秒)
		CodeExpireAt int `alias:"codeExpireAt" default:"300"`

		// CodeReplayAt 短信重发时间 (秒)
		CodeReplayAt int `alias:"codeReplayAt" default:"60"`
	}

	// LRConfig 登录注册配置信息
	LRConfig struct {
		// PasswordMinLen 密码最少字符
		PasswordMinLen int `alias:"passwordMinLen" default:"8"`

		// PasswordMaxLen 密码最大字符
		PasswordMaxLen int `alias:"passwordMaxLen" default:"16"`

		// IDSecret 用户ID加密key 16位
		IDSecret string `alias:"idSecret" default:"7581BDD8E8DA3839"`

		// TokenSecret 用户ID加密key 32位
		TokenSecret string `alias:"tokenSecret" default:"7581BDD8E8DA38397581BDD8E8DA3839"`

		// LoginTypesOfValidationCodes 验证代码的类型
		LoginTypesOfValidationCodes int `alias:"loginTypesOfValidationCodes" default:"0"`

		// TokenExpireAt token有效期 单位 s
		TokenExpireAt int `alias:"tokenExpireAt" default:"3600"`

		// SMS 短信配置
		SMS SMSConfig
	}
)

func login(req *corespb.Request, c LRConfig) (*corespb.Response, error) {
	var frame pb.Authentication_Form_Login
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	switch payload := frame.Payload.(type) {
	case *pb.Authentication_Form_Login_Account:
		return loginAccount(req, &frame, payload, c)

	case *pb.Authentication_Form_Login_Phone:
		// todo

	case *pb.Authentication_Form_Login_SMS:
		return loginMobliePhoneSMSSend(req, &frame, payload, c)
	}

	return nil, errors.New("InvalidProtoType")
}

func loginAccount(req *corespb.Request, reqFrame *pb.Authentication_Form_Login, form *pb.Authentication_Form_Login_Account, c LRConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
	}

	if req.Header == nil {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect), nil
	}

	var peerID string
	if m, ok := req.Header["PeerID"]; ok {
		peerID = m
	}

	if peerID == "" {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect), nil
	}

	// ^[\u4e00-\u9fa5]
	reg := regexp.MustCompile(`^[a-z0-9A-Z\p{Han}]+([\.|_|@][a-z0-9A-Z\p{Han}]+)*$`)
	if !reg.MatchString(form.Account.Username) {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect), nil
	}

	if !isValidPassword(form.Account.Password, c.PasswordMinLen, c.PasswordMaxLen) {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect), nil
	}

	if c.LoginTypesOfValidationCodes != 0 && !isValidValidationCodes(form, c) {
		return errcode.Bad(w, errcode.ErrVerificationCodeIncorrect), nil
	}

	account, err := dao.GetAccoutsBySchemeAName(reqFrame.Scheme, form.Account.Username)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect), nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Secret), []byte(form.Account.Username))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect), nil
	}

	if account.Status != 0 {
		return errcode.Bad(w, errcode.ErrAccountDisabled), nil
	}

	if account.ExpiresAt != 0 && account.ExpiresAt < time.Now().Unix() {
		return errcode.Bad(w, errcode.ErrAccountExpired), nil
	}

	// 防止相同账户重复登录
	if peerID == account.PeerID && account.PeerID != "" {
		kickedOut(account.PeerID)
	}

	// 更新
	db := dao.DB()
	if db == nil {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect), nil
	}

	db.Model(account).Updates(&dao.Accounts{PeerID: peerID})

	accountInfo, err := makeAuthenticationFormAccountInfo(account, c)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect, err.Error()), nil
	}

	resp := &pb.Authentication_Form_LoginReply{
		Scheme: reqFrame.Scheme,
		Payload: &pb.Authentication_Form_LoginReply_Account{
			Account: accountInfo,
		},
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func isValidValidationCodes(form *pb.Authentication_Form_Login_Account, c LRConfig) bool {
	switch payload := form.Account.ValidationCodes.(type) {
	case *pb.Authentication_Form_LoginAccount_Phone:
		return isValidValidationCodesPhone(payload.Phone.Phone, payload.Phone.Code, c)
	case *pb.Authentication_Form_LoginAccount_Code:
		// todo code
	}
	return false
}

func isValidValidationCodesPhone(phone string, code string, c LRConfig) bool {
	vcode, err := dao.GetSMSCode(phone, "login")
	if err != nil {
		return false
	}

	mcode, expire, err := dao.ParseSMSVerificationCode(vcode, c.SMS.CodeMaxLen)
	if err != nil {
		return false
	}

	if time.Now().Unix() > expire || code != mcode {
		return false
	}

	dao.RemoveSMSCode(phone, "login")
	return true
}

func loginMobliePhoneSMSSend(req *corespb.Request, reqFrame *pb.Authentication_Form_Login, sms *pb.Authentication_Form_Login_SMS, c LRConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
	}

	regular := "^(1[3-9])\\d{9}$"
	reg := regexp.MustCompile(regular)
	if !reg.MatchString(sms.SMS.Phone) {
		return errcode.Bad(w, errcode.ErrPhoneNumberInvalid), nil
	}

	vcode, err := dao.GetSMSCode(sms.SMS.Phone, "login")
	if err == nil {
		_, expire, err := dao.ParseSMSVerificationCode(vcode, c.SMS.CodeMaxLen)
		if err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		expireAt := expire - int64(c.SMS.CodeExpireAt)
		if time.Now().Sub(time.Unix(expireAt, 0)).Seconds() < float64(c.SMS.CodeReplayAt) {
			return errcode.Bad(w, errcode.ErrVerificationCodeExists), nil
		}

		dao.RemoveSMSCode(sms.SMS.Phone, "login")
	}

	code, err := dao.GenerateSMSCode(sms.SMS.Phone, c.SMS.CodeMaxLen, time.Duration(c.SMS.CodeExpireAt)*time.Second, "login")
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	// todo sms api
	fmt.Println("code:", code)
	resp := &pb.Authentication_Form_LoginReply{
		Scheme: reqFrame.Scheme,
		Payload: &pb.Authentication_Form_LoginReply_SMS{
			SMS: &pb.Authentication_Form_MobilePhoneSMSCode{
				Phone:    sms.SMS.Phone,
				ReplayAt: int32(time.Duration(c.SMS.CodeReplayAt) * time.Second),
				ExpireAt: int32(time.Duration(c.SMS.CodeExpireAt) * time.Second),
			},
		},
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func register(req *corespb.Request, c LRConfig) (*corespb.Response, error) {
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

	case *pb.Authentication_Form_Register_CheckUsername:
		return registerCheckUsername(req, &frame, payload, c)
	}

	return nil, errors.New("InvalidProtoType")
}

func registerMobliePhoneSMSSend(req *corespb.Request, reqFrame *pb.Authentication_Form_Register, sms *pb.Authentication_Form_Register_SMS, c LRConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
	}

	regular := "^(1[3-9])\\d{9}$"
	reg := regexp.MustCompile(regular)
	if !reg.MatchString(sms.SMS.Phone) {
		return errcode.Bad(w, errcode.ErrPhoneNumberInvalid), nil
	}

	vcode, err := dao.GetSMSCode(sms.SMS.Phone, "register")
	if err == nil {
		_, expire, err := dao.ParseSMSVerificationCode(vcode, c.SMS.CodeMaxLen)
		if err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		expireAt := expire - int64(c.SMS.CodeExpireAt)
		if time.Now().Sub(time.Unix(expireAt, 0)).Seconds() < float64(c.SMS.CodeReplayAt) {
			return errcode.Bad(w, errcode.ErrVerificationCodeExists), nil
		}

		dao.RemoveSMSCode(sms.SMS.Phone, "register")
	}

	code, err := dao.GenerateSMSCode(sms.SMS.Phone, c.SMS.CodeMaxLen, time.Duration(c.SMS.CodeExpireAt)*time.Second, "register")
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	// todo sms api
	fmt.Println("code:", code)
	resp := &pb.Authentication_Form_RegisterReply{
		Scheme: reqFrame.Scheme,
		Payload: &pb.Authentication_Form_RegisterReply_SMS{
			SMS: &pb.Authentication_Form_MobilePhoneSMSCode{
				Phone:    sms.SMS.Phone,
				ReplayAt: int32(time.Duration(c.SMS.CodeReplayAt) * time.Second),
				ExpireAt: int32(time.Duration(c.SMS.CodeExpireAt) * time.Second),
			},
		},
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func registerAccount(req *corespb.Request, reqFrame *pb.Authentication_Form_Register, r *pb.Authentication_Form_Register_Account, c LRConfig) (*corespb.Response, error) {
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

	if len(r.Account.PhoneCode) != c.SMS.CodeMaxLen {
		return errcode.Bad(w, errcode.ErrVerificationCodeIncorrect), nil
	}

	vcode, err := dao.GetSMSCode(r.Account.Phone)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	code, expire, err := dao.ParseSMSVerificationCode(vcode, c.SMS.CodeMaxLen)
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

	result := dao.DB().Create(&accounts)
	if result.Error != nil || result.RowsAffected != 1 {
		return errcode.Bad(w, errcode.ErrInternalServer, result.Error.Error()), nil
	}

	accountID, err := id.Encrypt(accounts.ID, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	unionID, err := id.Encrypt(accounts.UnionID, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.Authentication_Form_RegisterReply{
		Scheme: reqFrame.Scheme,
		Payload: &pb.Authentication_Form_RegisterReply_Account{
			Account: &pb.Authentication_Form_AccountInfo{
				ID:      accountID,
				UnionID: unionID,
			},
		},
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func registerCheckUsername(req *corespb.Request, reqFrame *pb.Authentication_Form_Register, r *pb.Authentication_Form_Register_CheckUsername, c LRConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
	}

	reply := pb.Authentication_Form_RegisterReply_CheckUsername{
		CheckUsername: &pb.Authentication_Form_RegisterCheckUsernameReply{
			OK: false,
		},
	}

	resp := &pb.Authentication_Form_RegisterReply{
		Scheme: reqFrame.Scheme,
	}

	// ^[\u4e00-\u9fa5]
	reg := regexp.MustCompile(`^[a-z0-9A-Z\p{Han}]+([\.|_|@][a-z0-9A-Z\p{Han}]+)*$`)
	if !reg.MatchString(r.CheckUsername.Username) {
		resp.Payload = &reply
		b, _ := grpcproto.Marshal(resp)
		w.Payload = &corespb.Response_Content{Content: b}
		return w, nil
	}

	_, err := dao.GetAccoutsBySchemeAName(reqFrame.Scheme, r.CheckUsername.Username)
	if err == gorm.ErrRecordNotFound {
		reply.CheckUsername.OK = true
	}

	resp.Payload = &reply
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

func makeAuthenticationFormAccountInfo(account *dao.Accounts, c LRConfig) (*pb.Authentication_Form_AccountInfo, error) {
	accountID, err := id.Encrypt(account.ID, []byte(c.IDSecret))
	if err != nil {
		return nil, err
	}

	unionID, err := id.Encrypt(account.UnionID, []byte(c.IDSecret))
	if err != nil {
		return nil, err
	}

	userID, err := id.Encrypt(account.UserID, []byte(c.IDSecret))
	if err != nil {
		return nil, err
	}

	token, err := helper.GenerateToken(account.ID, account.UnionID, time.Duration(c.TokenExpireAt)*time.Second, []byte(c.TokenSecret))
	if err != nil {
		return nil, err
	}

	return &pb.Authentication_Form_AccountInfo{
		ID:      accountID,
		UnionID: unionID,
		UserID:  userID,
		Token:   token,
	}, nil
}

// kickedOut 踢出用户
func kickedOut(peerID string) {
	endpointer := sd.Endpointer()
	if endpointer == nil {
		return
	}

	endpoints, err := endpointer.Endpoints()
	if err != nil {
		log.Error(Logger()).Log("action", "kickedOut", "error", err)
		return
	}

	nc := nats.Conn()
	if nc == nil {
		return
	}

	frame := agentpb.Agent_KickedOut{
		PeerID: []string{peerID},
	}

	frameBytes, _ := grpcproto.Marshal(&frame)
	w := corespb.Response{
		Command: agentproto.KickedOutCommand.Int32(),
		Payload: &corespb.Response_Content{Content: frameBytes},
		Header:  make(map[string]string),
	}

	w.Header["service"] = ServiceName
	w.Header["addr"] = sd.Endpoint().Addr()
	wBytes, _ := grpcproto.Marshal(&w)
	for _, endpoint := range endpoints {
		if endpoint.Name() != agent.ServiceName {
			continue
		}

		nc.Publish(endpoint.ID(), wBytes)
	}
}
