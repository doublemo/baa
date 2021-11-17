package robot

import (
	"errors"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/robot/dao"
	"github.com/doublemo/baa/kits/robot/session"
	grpcproto "github.com/golang/protobuf/proto"
)

func internalAccountLogin(name, secret string) (*pb.Authentication_Form_AccountInfo, error) {
	req := &corespb.Request{
		Command: command.AuthLogin.Int32(),
		Header:  map[string]string{"PeerId": "Robot"},
	}

	frame := &pb.Authentication_Form_Login{
		Scheme: "password",
		Payload: &pb.Authentication_Form_Login_Account{
			Account: &pb.Authentication_Form_LoginAccount{
				Username: name,
				Password: secret,
			},
		},
	}

	req.Payload, _ = grpcproto.Marshal(frame)
	resp, err := muxRouter.Handler(kit.Auth.Int32(), req)
	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		var respFrame pb.Authentication_Form_LoginReply
		{
			if err := grpcproto.Unmarshal(payload.Content, &respFrame); err != nil {
				return nil, err
			}
		}

		switch p := respFrame.Payload.(type) {
		case *pb.Authentication_Form_LoginReply_Account:
			return p.Account, nil
		default:
			return nil, errors.New("notsupported")
		}

	case *corespb.Response_Error:
		return nil, errors.New(payload.Error.Message)
	}

	return nil, errors.New("notsupported")
}

func doRequestLogin(peer session.Peer, c RobotConfig) error {
	rob, ok := peer.Params("Robots")
	if !ok {
		return errors.New("cann't find robot")
	}

	robot, ok := rob.(*dao.Robots)
	if !ok || len(robot.Secret) < 1 {
		return errors.New("cann't find robot")
	}

	password, err := decryptPassword(robot.Secret, []byte(c.PasswordSecret))
	if err != nil {
		return err
	}

	frame := &pb.Authentication_Form_Login{
		Scheme: robot.SchemaName,
		Payload: &pb.Authentication_Form_Login_Account{
			Account: &pb.Authentication_Form_LoginAccount{
				Username: robot.Name,
				Password: password,
			},
		},
	}

	bytes, err := grpcproto.Marshal(frame)
	if err != nil {
		return err
	}

	req := &coresproto.RequestBytes{
		Ver:     1,
		Cmd:     kit.Auth,
		SubCmd:  command.AuthLogin,
		Content: bytes,
		SeqID:   1,
	}

	r, err := req.Marshal()
	if err != nil {
		return err
	}

	return peer.Send(session.PeerMessagePayload{Data: r})
}

func login(peer session.Peer, w coresproto.Response, c RobotConfig) error {
	if w.StatusCode() != 0 {
		return errors.New(string(w.Body()))
	}

	var frame pb.Authentication_Form_LoginReply
	{
		if err := grpcproto.Unmarshal(w.Body(), &frame); err != nil {
			return err
		}
	}

	switch payload := frame.Payload.(type) {
	case *pb.Authentication_Form_LoginReply_Account:
		peer.SetParams("AccountID", payload.Account.ID)
		peer.SetParams("UnionID", payload.Account.UnionID)
		peer.SetParams("UserID", payload.Account.UserID)
		peer.SetParams("Token", payload.Account.Token)
	default:
		return nil
	}

	return openDataChannel(peer, c)
}
