package user

import (
	"errors"
	"regexp"
	"strconv"

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
		accountId uint64
		userId    uint64
	)

	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	if m, ok := req.Header["AccountID"]; ok {
		aid, err := strconv.ParseUint(m, 10, 64)
		if err != nil {
			return errcode.Bad(w, errcode.ErrInvalidAccountID, err.Error()), nil
		}

		accountId = aid
	}

	if m, ok := req.Header["UserID"]; ok {
		uid, err := strconv.ParseUint(m, 10, 64)
		if err != nil {
			return errcode.Bad(w, errcode.ErrInvalidAccountID, err.Error()), nil
		}
		userId = uid
	}

	reg := regexp.MustCompile("^(1[3-9])\\d{9}$")
	if frame.Info.Phone != "" && !reg.MatchString(frame.Info.Phone) {
		return errcode.Bad(w, errcode.ErrPhoneNumberInvalid), nil
	}

	_, userUint64Id, err := getAccountInfo(accountId, userId)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInvalidAccountID, err.Error()), nil
	}

	if len(frame.Info.Nickname) > c.NicknameMaxLength {
		return errcode.Bad(w, errcode.ErrNickNameLettersInvalid), nil
	}

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

	resp := &pb.User_Register_Reply{IndexNo: user.IndexNo}
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

func getAccountInfo(accountId, userId uint64) (uint64, uint64, error) {
	if accountId < 1 {
		return 0, 0, errors.New("Invalid Account ID")
	}

	if userId > 0 {
		return accountId, userId, nil
	}

	req := &corespb.Request{
		Command: command.AuthAccountInfo.Int32(),
		Header:  make(map[string]string),
	}

	req.Payload, _ = grpcproto.Marshal(&pb.Authentication_Account_Request{ID: &pb.Authentication_Account_Request_Uint64ID{Uint64ID: accountId}})
	resp, err := muxRouter.Handler(kit.Auth.Int32(), req)
	if err != nil {
		return 0, 0, err
	}

	if resp.Header == nil {
		return 0, 0, errors.New("Invalid Account ID")
	}

	uid, ok := resp.Header["UserID"]
	if !ok {
		return 0, 0, errors.New("Invalid Account ID")
	}

	uuid, err := strconv.ParseUint(uid, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return accountId, uuid, nil
}
