package user

import (
	"errors"
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

func getUserInfo(req *corespb.Request, c UserConfig) (*corespb.Response, error) {
	var frame pb.User_InfoRequest
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	switch payload := frame.Payload.(type) {
	case *pb.User_InfoRequest_UserId:
		return findUserInfo(req, payload.UserId, c)

	case *pb.User_InfoRequest_MoreUserId:
		return findUserInfos(req, payload.MoreUserId.Values, c)

	case *pb.User_InfoRequest_UserIdFromString:
		uid, err := id.Decrypt(payload.UserIdFromString, []byte(c.IDSecret))
		if err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		return findUserInfo(req, uid, c)

	case *pb.User_InfoRequest_MoreUserIdString:
		data := make([]uint64, len(payload.MoreUserIdString.Values))
		for i, value := range payload.MoreUserIdString.Values {
			uid, err := id.Decrypt(value, []byte(c.IDSecret))
			if err != nil {
				return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
			}

			data[i] = uid
		}

		return findUserInfos(req, data, c)
	}

	return nil, errors.New("notsupported")
}

func findUserInfo(req *corespb.Request, userid uint64, c UserConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	users, err := dao.FindUsersByID(userid)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	uid, _ := id.Encrypt(userid, []byte(c.IDSecret))
	frame := &pb.User_InfoReply{
		Payload: &pb.User_InfoReply_Value{
			Value: &pb.User_Info{
				UserId:   uid,
				Nickname: users.Nickname,
				Headimg:  users.Headimg,
				Age:      int32(users.Age),
				Sex:      int32(users.Sex),
				Idcard:   users.Idcard,
				Phone:    users.Phone,
				IndexNo:  users.IndexNo,
			},
		},
	}

	bytes, _ := grpcproto.Marshal(frame)
	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
}

func findUserInfos(req *corespb.Request, userid []uint64, c UserConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	users, err := dao.FindUsersByMoreID(userid...)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	values := &pb.User_MoreInfo{Values: make([]*pb.User_Info, len(users))}
	for i, user := range users {
		uid, _ := id.Encrypt(user.ID, []byte(c.IDSecret))
		values.Values[i] = &pb.User_Info{
			UserId:   uid,
			Nickname: user.Nickname,
			Headimg:  user.Headimg,
			Age:      int32(user.Age),
			Sex:      int32(user.Sex),
			Idcard:   user.Idcard,
			Phone:    user.Phone,
			IndexNo:  user.IndexNo,
		}
	}

	frame := &pb.User_InfoReply{
		Payload: &pb.User_InfoReply_Values{
			Values: values,
		},
	}
	bytes, _ := grpcproto.Marshal(frame)
	w.Payload = &corespb.Response_Content{Content: bytes}
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
