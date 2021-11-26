package robot

import (
	"context"
	"errors"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/robot/session"
	"github.com/golang/protobuf/jsonpb"
)

// syncContacts 同步联系人
func syncContacts(peer session.Peer) {

}

// doCheckFriendRequest 检查好友请求
func doCheckFriendRequest(peer session.Peer, page int32, c RobotConfig) error {
	userid, ok := peer.Params("UserID")
	if !ok {
		return errors.New("invalid UserID")
	}

	uid := userid.(string)
	agent, ok := peer.Params("AgentHttp")
	if !ok {
		return errors.New("invalid agent addr")
	}

	tk, ok := peer.Params("Token")
	if !ok {
		return errors.New("invalid token")
	}

	frame := &pb.User_Contacts_FriendRequestList{
		UserId:  userid.(string),
		Page:    page,
		Size:    10,
		Version: 0,
	}

	pm := jsonpb.Marshaler{}
	data, err := pm.MarshalToString(frame)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	body, errcode := RequestPostWithContext(ctx, command.UserContactsRequest, agent.(string)+"/v1/user", []byte(data), []byte(c.CommandSecret), tk.(string), c.CSRFSecret)
	if errcode != nil {
		return errcode.ToError()
	}

	var frameW pb.User_Contacts_FriendRequestListReply
	{
		if err := jsonpb.UnmarshalString(string(body), &frameW); err != nil {
			log.Error(Logger()).Log("action", "doAddFriendRequest", "error", err.Error(), "body", string(body))
			return err
		}
	}

	for _, v := range frameW.Values {
		// 过滤掉自己发送的请求
		if v.FromID == uid {
			continue
		}

		if v.Status == 0 {
			// 接受
			log.Info(Logger()).Log("action", "doacceptFriendRequest", "peer", peer.ID(), "friend", v.FriendId)
			if err := doAcceptFriendRequest(peer, v.FriendId, c); err != nil {
				return err
			}
		}
	}

	return nil
}

func doAcceptFriendRequest(peer session.Peer, friendid string, c RobotConfig) error {
	userid, ok := peer.Params("UserID")
	if !ok {
		return errors.New("invalid UserID")
	}

	agent, ok := peer.Params("AgentHttp")
	if !ok {
		return errors.New("invalid agent addr")
	}

	tk, ok := peer.Params("Token")
	if !ok {
		return errors.New("invalid token")
	}

	frame := &pb.User_Contacts_Request{
		Payload: &pb.User_Contacts_Request_Accept{
			Accept: &pb.User_Contacts_Accept{
				FriendId: friendid,
				Remark:   "",
				UserId:   userid.(string),
			},
		},
	}

	pm := jsonpb.Marshaler{}
	data, err := pm.MarshalToString(frame)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	body, errcode := RequestPostWithContext(ctx, command.UserContacts, agent.(string)+"/v1/user", []byte(data), []byte(c.CommandSecret), tk.(string), c.CSRFSecret)
	if errcode != nil {
		return errcode.ToError()
	}

	var frameW pb.User_Contacts_Reply
	{
		if err := jsonpb.UnmarshalString(string(body), &frameW); err != nil {
			log.Error(Logger()).Log("action", "doAddFriendRequest", "error", err.Error(), "body", string(body))
			return err
		}
	}

	if frameW.OK {
		return nil
	}

	return errors.New("accept friend failed")
}

func doAddFriendRequest(peer session.Peer, friendid string, c RobotConfig) error {
	userid, ok := peer.Params("UserID")
	if !ok {
		return errors.New("invalid UserID")
	}

	uid := userid.(string)
	agent, ok := peer.Params("AgentHttp")
	if !ok {
		return errors.New("invalid agent addr")
	}

	tk, ok := peer.Params("Token")
	if !ok {
		return errors.New("invalid token")
	}

	frame := &pb.User_Contacts_Request{
		Payload: &pb.User_Contacts_Request_Add{
			Add: &pb.User_Contacts_Add{
				FriendId: friendid,
				Remark:   "",
				Message:  "你好，我是小莉。想问点事儿",
				UserId:   uid,
			},
		},
	}

	pm := jsonpb.Marshaler{}
	data, err := pm.MarshalToString(frame)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	body, errcode := RequestPostWithContext(ctx, command.UserContacts, agent.(string)+"/v1/user", []byte(data), []byte(c.CommandSecret), tk.(string), c.CSRFSecret)
	if errcode != nil {
		return errcode.ToError()
	}

	var frameW pb.User_Contacts_Reply
	{
		if err := jsonpb.UnmarshalString(string(body), &frameW); err != nil {
			log.Error(Logger()).Log("action", "doAddFriendRequest", "error", err.Error(), "body", string(body))
			return err
		}
	}

	if frameW.OK {
		return nil
	}

	return errors.New("add friend failed")
}

func doFindFriendRequest(peer session.Peer, friend string, c RobotConfig) (*pb.User_Info, error) {
	agent, ok := peer.Params("AgentHttp")
	if !ok {
		return nil, errors.New("invalid agent addr")
	}

	tk, ok := peer.Params("Token")
	if !ok {
		return nil, errors.New("invalid token")
	}

	frame := &pb.User_Contacts_Request{
		Payload: &pb.User_Contacts_Request_SearchFriend{
			SearchFriend: friend,
		},
	}

	pm := jsonpb.Marshaler{}
	data, err := pm.MarshalToString(frame)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	body, errcode := RequestPostWithContext(ctx, command.UserContacts, agent.(string)+"/v1/user", []byte(data), []byte(c.CommandSecret), tk.(string), c.CSRFSecret)
	if errcode != nil {
		return nil, errcode.ToError()
	}

	var frameW pb.User_Info
	{
		if err := jsonpb.UnmarshalString(string(body), &frameW); err != nil {
			log.Error(Logger()).Log("action", "doAddFriendRequest", "error", err.Error(), "body", string(body))
			return nil, err
		}
	}

	return &frameW, nil
}

func doFindRefuseRequest(peer session.Peer, friendid string, c RobotConfig) error {
	userid, ok := peer.Params("UserID")
	if !ok {
		return errors.New("invalid UserID")
	}

	uid := userid.(string)

	agent, ok := peer.Params("AgentHttp")
	if !ok {
		return errors.New("invalid agent addr")
	}

	tk, ok := peer.Params("Token")
	if !ok {
		return errors.New("invalid token")
	}

	frame := &pb.User_Contacts_Request{
		Payload: &pb.User_Contacts_Request_Refuse{
			Refuse: &pb.User_Contacts_Refuse{
				FriendId: friendid,
				UserId:   uid,
			},
		},
	}

	pm := jsonpb.Marshaler{}
	data, err := pm.MarshalToString(frame)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	body, errcode := RequestPostWithContext(ctx, command.UserContacts, agent.(string)+"/v1/user", []byte(data), []byte(c.CommandSecret), tk.(string), c.CSRFSecret)
	if errcode != nil {
		return errcode.ToError()
	}

	var frameW pb.User_Contacts_Reply
	{
		if err := jsonpb.UnmarshalString(string(body), &frameW); err != nil {
			log.Error(Logger()).Log("action", "doAddFriendRequest", "error", err.Error(), "body", string(body))
			return err
		}
	}

	if frameW.OK {
		return nil
	}

	return errors.New("refuse friend failed")
}
