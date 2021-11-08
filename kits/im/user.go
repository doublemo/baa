package im

import (
	"errors"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

func checkIsMyFriend(userid, friendid, topic uint64) (bool, error) {
	frame := &pb.User_Contacts_IsFriend{
		FriendId: friendid,
		UserId:   userid,
		Topic:    topic,
	}

	b, _ := grpcproto.Marshal(frame)
	resp, err := muxRouter.Handler(kit.User.Int32(), &corespb.Request{Command: command.UserCheckIsMyFriend.Int32(), Payload: b})
	if err != nil {
		return false, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		resp := pb.User_Contacts_Reply{}
		if err := grpcproto.Unmarshal(payload.Content, &resp); err != nil {
			return false, err
		}

		return resp.OK, nil

	case *corespb.Response_Error:
		return false, errors.New(payload.Error.Message)
	}

	return false, nil
}
