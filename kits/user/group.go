package user

import (
	"github.com/doublemo/baa/cores/crypto/id"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/kits/user/dao"
	"github.com/doublemo/baa/kits/user/errcode"
	"github.com/golang/protobuf/jsonpb"
	grpcproto "github.com/golang/protobuf/proto"
)

// GroupConfig 群配置
type GroupConfig struct {
	// IDSecret 用户ID 加密key 16位
	IDSecret string `alias:"idSecret" default:"7581BDD8E8DA3839"`
}

func checkInGroup(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.User_Group_In
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	resp := pb.User_Group_Reply{
		OK: false,
	}

	bytes, _ := grpcproto.Marshal(&resp)
	w.Payload = &corespb.Response_Content{Content: bytes}
	contacts, err := dao.FindContactsByUserIDAndTopic(frame.UserId, frame.GroupId, "friend_id", "status")
	if err != nil {
		return w, nil
	}

	if contacts.FriendID != frame.GroupId || contacts.Status != 0 {
		return w, nil
	}

	resp.OK = true
	bytes, _ = grpcproto.Marshal(&resp)
	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
}

func groupMembers(req *corespb.Request, c GroupConfig) (*corespb.Response, error) {
	var frame pb.User_Group_MembersListRequest
	{
		if router.IsHTTP(req) {
			if err := jsonpb.UnmarshalString(string(req.Payload), &frame); err != nil {
				return nil, err
			}
		} else {
			if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
				return nil, err
			}
		}
	}

	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	gid, err := id.Decrypt(frame.GroupId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	list, count, err := dao.FindGroupsMembersByGroupID(gid, frame.Page, frame.Size, frame.Version)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.User_Group_MembersListReply{
		Values:      make([]*pb.User_Group_Member, len(list)),
		Page:        frame.Page,
		Size:        frame.Size,
		RecordCount: int32(count),
	}

	for k, v := range list {
		enfid, _ := id.Encrypt(v.UserID, []byte(c.IDSecret))
		resp.Values[k] = &pb.User_Group_Member{
			UserId:   enfid,
			Nickname: v.Nickname,
			Headimg:  v.Headimg,
			Sex:      int32(v.Sex),
			Topic:    v.Topic,
			Alias:    v.Alias,
			Version:  v.Version,
		}
	}

	var respBytes []byte
	{
		if router.IsHTTP(req) {
			jsonpbM := &jsonpb.Marshaler{}
			json, _ := jsonpbM.MarshalToString(resp)
			respBytes = []byte(json)
		} else {
			respBytes, _ = grpcproto.Marshal(resp)
		}
	}
	w.Payload = &corespb.Response_Content{Content: respBytes}
	return w, nil
}

func groupMembersID(req *corespb.Request, c GroupConfig) (*corespb.Response, error) {
	var frame pb.User_Group_MembersIDListRequest
	{
		if router.IsHTTP(req) {
			if err := jsonpb.UnmarshalString(string(req.Payload), &frame); err != nil {
				return nil, err
			}
		} else {
			if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
				return nil, err
			}
		}
	}

	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	gid, err := id.Decrypt(frame.GroupId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	list, count, err := dao.FindGroupsMembersIDByGroupID(gid, frame.Page, frame.Size)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.User_Group_MembersIDListReply{
		Values:      list,
		Page:        frame.Page,
		Size:        frame.Size,
		RecordCount: int32(count),
	}

	var respBytes []byte
	{
		if router.IsHTTP(req) {
			jsonpbM := &jsonpb.Marshaler{}
			json, _ := jsonpbM.MarshalToString(resp)
			respBytes = []byte(json)
		} else {
			respBytes, _ = grpcproto.Marshal(resp)
		}
	}
	w.Payload = &corespb.Response_Content{Content: respBytes}
	return w, nil
}
