package agent

import (
	log "github.com/doublemo/baa/cores/log/level"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/agent/router"
	"github.com/doublemo/baa/kits/agent/session"
	authproto "github.com/doublemo/baa/kits/auth/proto"
	authpb "github.com/doublemo/baa/kits/auth/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

func authenticationHookAfter(r *router.Call) {
	r.OnAfterCall(func(p session.Peer, r *corespb.Response) error {
		onLogin(p, r)
		return nil
	})
}

func authenticationHookDestroy(r *router.Call) {
	r.OnDestroy(func(peer session.Peer) {
		if userID, ok := peer.Params("UserID"); ok {
			session.RemoveDict(userID.(string), peer)
		}

		accountID, ok := peer.Params("AccountID")
		if !ok {
			return
		}

		// 处理玩家离线
		frame := authpb.Authentication_Form_Logout{
			Payload: &authpb.Authentication_Form_Logout_PeerID{PeerID: peer.ID()},
		}

		body, _ := grpcproto.Marshal(&frame)
		_, err := r.Call(&corespb.Request{
			Header:  map[string]string{"PeerId": peer.ID(), "AccountID": accountID.(string)},
			Command: authproto.OfflineCommand.Int32(),
			Payload: body,
		})

		if err != nil {
			log.Error(Logger()).Log("action", "Destroy", "error", err)
		}
	})
}

func onLogin(peer session.Peer, w *corespb.Response) {
	if w.Command != authproto.LoginCommand.Int32() {
		return
	}

	var content []byte
	switch payload := w.Payload.(type) {
	case *corespb.Response_Content:
		content = payload.Content
	case *corespb.Response_Error:
		return
	default:
		return
	}

	var frame authpb.Authentication_Form_LoginReply
	{
		if err := grpcproto.Unmarshal(content, &frame); err != nil {
			log.Error(Logger()).Log("action", "onLogin", "error", err)
			return
		}
	}

	switch payload := frame.Payload.(type) {
	case *authpb.Authentication_Form_LoginReply_Account:
		peer.SetParams("AccountID", payload.Account.ID)
		peer.SetParams("AccountUnionID", payload.Account.UnionID)
		peer.SetParams("UserID", payload.Account.UserID)
		peer.SetParams("Token", payload.Account.Token)
		session.AddDict(payload.Account.UserID, peer)

	default:
		return
	}
}
