package robot

import (
	"errors"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

func internalAccountLogin(name, secret string) (*pb.Authentication_Form_AccountInfo, error) {
	req := &corespb.Request{
		Command: command.AuthLogin.Int32(),
		Header:  make(map[string]string),
	}

	frame := &pb.Authentication_Form_Login{
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
