package auth

import (
	"strconv"

	"github.com/doublemo/baa/cores/crypto/id"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/auth/dao"
	"github.com/doublemo/baa/kits/auth/errcode"
	grpcproto "github.com/golang/protobuf/proto"
)

func accountInfo(req *corespb.Request, c LRConfig) (*corespb.Response, error) {
	var frame pb.Authentication_Account_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	var accountId uint64
	switch data := frame.ID.(type) {
	case *pb.Authentication_Account_Request_StringID:
		if m, err := id.Decrypt(data.StringID, []byte(c.IDSecret)); err == nil {
			accountId = m
		}

	case *pb.Authentication_Account_Request_Uint64ID:
		accountId = data.Uint64ID
	}

	if accountId < 1 {
		return errcode.Bad(w, errcode.ErrAccountIDInvalid), nil
	}

	info, err := dao.GetAccoutsByID(accountId)
	if err != nil {
		return errcode.Bad(w, errcode.ErrAccountIDInvalid, err.Error()), nil
	}

	resp := &pb.Authentication_Account_Info{
		Token:     "",
		Schema:    info.SchemaName,
		Name:      info.Name,
		Status:    int32(info.Status),
		ExpiresAt: info.ExpiresAt,
		CreatedAt: info.CreatedAt.Unix(),
	}

	resp.ID, _ = id.Encrypt(info.ID, []byte(c.IDSecret))
	resp.UnionID, _ = id.Encrypt(info.UnionID, []byte(c.IDSecret))
	resp.UserID, _ = id.Encrypt(info.UserID, []byte(c.IDSecret))
	respBytes, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: respBytes}
	w.Header["AccountID"] = strconv.FormatUint(info.ID, 10)
	w.Header["UnionID"] = strconv.FormatUint(info.UnionID, 10)
	w.Header["UserID"] = strconv.FormatUint(info.UserID, 10)
	return w, nil
}

// 处理账户下线
func offline(r *corespb.Request, c LRConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: r.Command,
		Payload: &corespb.Response_Content{Content: r.Payload},
	}

	if r.Header == nil {
		return errcode.Bad(w, errcode.ErrInternalServer, "header is nil"), nil
	}

	peerid, ok := r.Header["PeerId"]
	if !ok {
		return errcode.Bad(w, errcode.ErrInternalServer, "PeerId is undefined"), nil
	}

	var frame pb.Authentication_Form_Logout
	{
		if err := grpcproto.Unmarshal(r.Payload, &frame); err != nil {
			return nil, err
		}
	}

	userid, err := strconv.ParseUint(frame.UserID, 10, 64)
	if err != nil {
		return errcode.Bad(w, errcode.ErrAccountIDInvalid, err.Error()), nil
	}

	// 更新用户在线状态
	online, _ := grpcproto.Marshal(&pb.SM_User_Action_Offline{
		UserId:   userid,
		Platform: "",
		PeerId:   peerid,
	})

	if err := publishUserState(&pb.SM_Event{Action: pb.SM_ActionUserOffline, Data: online}); err != nil {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect, "change status falied"), nil
	}
	return w, nil
}
