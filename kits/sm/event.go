package sm

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

func eventHandler(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SM_Event
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	switch frame.Action {
	case pb.SM_ActionUserOnline:
		return nil, online(&frame)

	case pb.SM_ActionUserOffline:
		return nil, offline(&frame)

	case pb.SM_ActionUserStatusUpdate:
		return nil, updateUserStatus(&frame)
	}

	return nil, nil
}
