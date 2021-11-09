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

func checkInGroup(userid, topic uint64) (bool, error) {
	frame := &pb.User_Group_In{
		UserId:  userid,
		GroupId: topic,
	}

	b, _ := grpcproto.Marshal(frame)
	resp, err := muxRouter.Handler(kit.User.Int32(), &corespb.Request{Command: command.UserCheckInGroup.Int32(), Payload: b})
	if err != nil {
		return false, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		resp := pb.User_Group_Reply{}
		if err := grpcproto.Unmarshal(payload.Content, &resp); err != nil {
			return false, err
		}

		return resp.OK, nil

	case *corespb.Response_Error:
		return false, errors.New(payload.Error.Message)
	}

	return false, nil
}

func groupMembersID(groupid string, page, size int32) ([]uint64, int64, error) {
	frame := &pb.User_Group_MembersIDListRequest{
		GroupId: groupid,
		Page:    page,
		Size:    size,
		Version: 0,
	}

	b, _ := grpcproto.Marshal(frame)
	resp, err := muxRouter.Handler(kit.User.Int32(), &corespb.Request{Command: command.UserGroupMembersValidID.Int32(), Payload: b})
	if err != nil {
		return nil, 0, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		resp := pb.User_Group_MembersIDListReply{}
		if err := grpcproto.Unmarshal(payload.Content, &resp); err != nil {
			return nil, 0, err
		}

		return resp.Values, int64(resp.RecordCount), nil

	case *corespb.Response_Error:
		return nil, 0, errors.New(payload.Error.Message)
	}

	return []uint64{}, 0, nil
}
