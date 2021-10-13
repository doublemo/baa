package auth

import (
	"errors"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/auth/dao"
	"github.com/doublemo/baa/kits/auth/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

// 处理账户下线
func offline(r *corespb.Request) (*corespb.Response, error) {
	var peerID string
	{
		if r.Header != nil {
			if m, ok := r.Header["PeerId"]; ok {
				peerID = m
			}
		}
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
		if peerID != payload.PeerID {
			return nil, errors.New("InvalidPeerID")
		}

		accounts, err = dao.GetAccoutsByPeerID(payload.PeerID)
		if err != nil {
			return nil, err
		}

	case *pb.Authentication_Form_Logout_Token:
	}

	w := &corespb.Response{
		Command: r.Command,
		Payload: &corespb.Response_Content{Content: r.Payload},
	}

	if accounts == nil {
		return w, nil
	}

	dao.UpdatesAccountByID(accounts.ID, "peer_id", "")
	return w, nil
}
