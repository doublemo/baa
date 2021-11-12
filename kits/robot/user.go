package robot

import (
	"errors"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

func internalUserinfo(id string) (*pb.User_Info, error) {
	req := &corespb.Request{
		Command: command.UserInfo.Int32(),
		Header:  make(map[string]string),
	}

	frame := &pb.User_InfoRequest{
		Payload: &pb.User_InfoRequest_UserIdFromString{
			UserIdFromString: id,
		},
	}

	req.Payload, _ = grpcproto.Marshal(frame)
	resp, err := muxRouter.Handler(kit.Auth.Int32(), req)
	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		var respFrame pb.User_InfoReply
		{
			if err := grpcproto.Unmarshal(payload.Content, &respFrame); err != nil {
				return nil, err
			}
		}

		switch p := respFrame.Payload.(type) {
		case *pb.User_InfoReply_Value:
			return p.Value, nil
		default:
			return nil, errors.New("notsupported")
		}

	case *corespb.Response_Error:
		return nil, errors.New(payload.Error.Message)
	}

	return nil, errors.New("notsupported")
}
