package user

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/doublemo/baa/cores/crypto/id"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/user/dao"
	"github.com/doublemo/baa/kits/user/errcode"
	grpcproto "github.com/golang/protobuf/proto"
)

// UserConfig 配置
type UserConfig struct {
	// IDSecret 用户ID 加密key 16位
	IDSecret string `alias:"idSecret" default:"7581BDD8E8DA3839"`

	// IdxSecret 用户生成用户默认索引号
	IdxSecret string `alias:"idxSecret" default:"A5C11DD8E8DA3830"`

	// NicknameMaxLength 昵称最大长度
	NicknameMaxLength int `alias:"nicknameMaxLength" default:"34"`

	// ContactRequestExpireAt 联系人增加请求过期时间，单位: 小时
	ContactRequestExpireAt int `alias:"contactRequestExpireAt" default:"168"`
}

func register(req *corespb.Request, c UserConfig) (*corespb.Response, error) {
	var frame pb.User_Register_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	var (
		accountId string
		userId    string
	)

	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	if m, ok := req.Header["AccountID"]; ok {
		accountId = m
	}

	if m, ok := req.Header["UserID"]; ok {
		userId = m
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	fmt.Println(accountId, frame.AccountId)

	if accountId != frame.AccountId {
		return errcode.Bad(w, errcode.ErrInvalidAccountID), nil
	}

	reg := regexp.MustCompile("^(1[3-9])\\d{9}$")
	if frame.Info.Phone != "" && !reg.MatchString(frame.Info.Phone) {
		return errcode.Bad(w, errcode.ErrPhoneNumberInvalid), nil
	}

	_, userUint64Id, err := getAccountInfo(accountId, userId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInvalidAccountID, err.Error()), nil
	}

	if len(frame.Info.Nickname) > c.NicknameMaxLength {
		return errcode.Bad(w, errcode.ErrNickNameLettersInvalid), nil
	}

	userId, _ = id.Encrypt(userUint64Id, []byte(c.IDSecret))
	indexNo, _ := id.Encrypt(userUint64Id, []byte(c.IdxSecret))
	user := dao.Users{
		ID:       userUint64Id,
		IndexNo:  "ID_" + indexNo,
		Nickname: frame.Info.Nickname,
		Headimg:  frame.Info.Headimg,
		Age:      int8(frame.Info.Age),
		Sex:      int8(frame.Info.Sex),
		Idcard:   frame.Info.Idcard,
		Phone:    frame.Info.Phone,
	}

	if err := dao.CreateUsers(&user); err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.User_Register_Reply{}
	resp.UserId, _ = id.Encrypt(user.ID, []byte(c.IDSecret))
	respBytes, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: respBytes}
	return w, nil
}

func getAccountInfo(accountId, userId string, secret []byte) (uint64, uint64, error) {
	auid, err := id.Decrypt(accountId, secret)
	if err != nil {
		return 0, 0, err
	}

	var uuid uint64
	if userId != "" {
		uuid, err = id.Decrypt(userId, secret)
		if err != nil {
			return 0, 0, err
		}
	}

	if auid < 1 {
		return 0, 0, errors.New("Invalid Account ID")
	}

	if uuid > 0 {
		return auid, uuid, nil
	}

	req := &corespb.Request{
		Command: command.AuthAccountInfo.Int32(),
		Header:  make(map[string]string),
	}

	req.Payload, _ = grpcproto.Marshal(&pb.Authentication_Account_Request{AccountId: accountId})
	resp, err := muxRouter.Handler(kit.Auth.Int32(), req)
	if err != nil {
		return 0, 0, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		var frame pb.Authentication_Account_Info
		{
			if err := grpcproto.Unmarshal(payload.Content, &frame); err != nil {
				return 0, 0, err
			}
		}

		m, err := id.Decrypt(frame.UserID, secret)
		if err != nil {
			return 0, 0, err
		}
		uuid = m

	case *corespb.Response_Error:
		return 0, 0, errors.New(payload.Error.String())
	}

	if uuid < 1 {
		return 0, 0, errors.New("Invalid Account ID")
	}

	return auid, uuid, nil
}
