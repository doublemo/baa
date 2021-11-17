package auth

import (
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

	accountId, err := id.Decrypt(frame.AccountId, []byte(c.IDSecret))
	if err != nil {
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
	return w, nil
}

// 处理账户下线
func offline(r *corespb.Request, c LRConfig) (*corespb.Response, error) {
	var (
		accountID string
		peerID    string
	)

	if r.Header != nil {
		if m, ok := r.Header["PeerId"]; ok {
			peerID = m
		}

		if m, ok := r.Header["AccountID"]; ok {
			accountID = m
		}
	}

	uid, err0 := id.Decrypt(accountID, []byte(c.IDSecret))
	if err0 != nil {
		return nil, err0
	}

	var frame pb.Authentication_Form_Logout
	{
		if err := grpcproto.Unmarshal(r.Payload, &frame); err != nil {
			return nil, err
		}
	}

	var (
		accounts *dao.Accounts
		err      error
	)

	switch payload := frame.Payload.(type) {
	case *pb.Authentication_Form_Logout_ID:
	case *pb.Authentication_Form_Logout_PeerID:
		peerID = payload.PeerID
		accounts, err = dao.GetAccoutsByID(uid)
		if err != nil {
			return nil, err
		}

	case *pb.Authentication_Form_Logout_Token:
	}

	w := &corespb.Response{
		Command: r.Command,
		Payload: &corespb.Response_Content{Content: r.Payload},
	}

	if accounts == nil || accounts.PeerID != peerID {
		return w, nil
	}

	// 更新用户在线状态
	online, _ := grpcproto.Marshal(&pb.SM_User_Action_Offline{
		UserId:   accounts.UserID,
		Platform: "pc",
	})

	if err := publishUserState(&pb.SM_Event{Action: pb.SM_ActionUserOffline, Data: online}); err != nil {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect, "change status falied"), nil
	}
	dao.UpdatesAccountByID(accounts.ID, "peer_id", "")
	return w, nil
}
