package auth

import (
	"strconv"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/auth/dao"
	grpcproto "github.com/golang/protobuf/proto"
)

// 处理账户下线
func offline(r *corespb.Request) (*corespb.Response, error) {
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

	uid, err0 := strconv.ParseUint(accountID, 10, 64)
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

	deleteUserStatus(&pb.USRT_User{ID: accounts.ID, Type: accounts.Schema, Value: sd.Endpoint().ID()})
	dao.UpdatesAccountByID(accounts.ID, "peer_id", "")
	return w, nil
}
