package agent

import (
	"errors"
	"fmt"
	"strconv"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/agent/router"
	"github.com/doublemo/baa/kits/agent/session"
	grpcproto "github.com/golang/protobuf/proto"
)

func authenticateCall(r *router.Call) {
	r.OnBeforeCall(func(p session.Peer, r1 proto.Request, r2 *corespb.Request) error {
		accountID, _ := p.Params("AccountID")
		userID, _ := p.Params("UserID")
		r2.Header["AccountID"] = strconv.FormatUint(accountID.(uint64), 10)
		r2.Header["UserID"] = strconv.FormatUint(userID.(uint64), 10)
		return nil
	})
}

func authenticateStream(r *router.Stream) {
	r.OnSend(func(p session.Peer, r1 proto.Request, r2 *corespb.Request) error {
		accountID, _ := p.Params("AccountID")
		userID, _ := p.Params("UserID")
		r2.Header["AccountID"] = strconv.FormatUint(accountID.(uint64), 10)
		r2.Header["UserID"] = strconv.FormatUint(userID.(uint64), 10)
		return nil
	})
}

func authenticationHookAfter(r *router.Call) {
	r.OnAfterCall(func(p session.Peer, r *corespb.Response) error {
		onLogin(p, r)
		return nil
	})
}

func authenticationHookDestroy(r *router.Call) {
	r.OnDestroy(func(peer session.Peer) {
		if userID, ok := peer.Params("UserID"); ok {
			session.RemoveDict(userID.(uint64), peer)
		}

		accountID, ok := peer.Params("AccountID")
		if !ok {
			return
		}

		// 处理玩家离线
		frame := pb.Authentication_Form_Logout{
			Payload: &pb.Authentication_Form_Logout_PeerID{PeerID: peer.ID()},
		}

		body, _ := grpcproto.Marshal(&frame)
		_, err := r.Call(&corespb.Request{
			Header:  map[string]string{"PeerId": peer.ID(), "AccountID": strconv.FormatUint(accountID.(uint64), 10)},
			Command: command.AuthOffline.Int32(),
			Payload: body,
		})

		if err != nil {
			log.Error(Logger()).Log("action", "Destroy", "error", err)
		}
	})
}

func onLogin(peer session.Peer, w *corespb.Response) {
	if w.Command != command.AuthLogin.Int32() {
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

	var frame pb.Authentication_Form_LoginReply
	{
		if err := grpcproto.Unmarshal(content, &frame); err != nil {
			log.Error(Logger()).Log("action", "onLogin", "error", err)
			return
		}
	}

	if w.Header == nil {
		return
	}

	if _, ok := w.Header["AccountID"]; !ok {
		return
	}

	if _, ok := w.Header["UnionID"]; !ok {
		return
	}

	if _, ok := w.Header["UserID"]; !ok {
		return
	}

	if m, err := strconv.ParseUint(w.Header["AccountID"], 10, 64); err == nil && m > 0 {
		peer.SetParams("AccountID", m)
	}

	if m, err := strconv.ParseUint(w.Header["UnionID"], 10, 64); err == nil && m > 0 {
		peer.SetParams("AccountUnionID", m)
	}

	if m, err := strconv.ParseUint(w.Header["UserID"], 10, 64); err == nil && m > 0 {
		peer.SetParams("UserID", m)
		session.AddDict(m, peer)
	}

	switch payload := frame.Payload.(type) {
	case *pb.Authentication_Form_LoginReply_Account:
		peer.SetParams("Token", payload.Account.Token)

	default:
		return
	}
}

func authenticateToken(token string) (*pb.Authentication_Form_Authorized_Info, error) {
	req := &corespb.Request{
		Header:  make(map[string]string),
		Command: int32(command.AuthorizedToken),
	}

	req.Payload, _ = grpcproto.Marshal(&pb.Authentication_Form_Authorized_Token{Token: token})
	resp, err := muxRouter.Handler(kit.Auth.Int32(), req)
	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		var info pb.Authentication_Form_Authorized_Info
		{
			if err := grpcproto.Unmarshal(payload.Content, &info); err != nil {
				return nil, err
			}
		}

		return &info, nil
	case *corespb.Response_Error:
		return nil, fmt.Errorf("authenticateToken: code %d error:%s", payload.Error.Code, payload.Error.Message)
	}

	return nil, errors.New("notsupport")
}
