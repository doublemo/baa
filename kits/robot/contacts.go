package robot

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/doublemo/baa/cores/crypto/id"
	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/robot/dao"
	"github.com/doublemo/baa/kits/robot/session"
	"github.com/golang/protobuf/jsonpb"
)

// syncContacts 同步联系人
func syncContacts(peer session.Peer, userid uint64, ver int64, c RobotConfig) error {
	var (
		page       int32 = 1
		maxVersion int64 = ver
	)
	defer func() {
		if maxVersion > ver {
			dao.UpsertVersionByID(&dao.RobotVersionManagers{
				ID:        "contacts",
				UserID:    userid,
				VersionID: maxVersion,
				Version:   "",
			})
		}
	}()

reload:
	contacts, err := doContactsRequest(peer, page, ver, c)
	if err != nil {
		return err
	}

	if len(contacts.Values) < 1 {
		return nil
	}

	contactsData := make([]*dao.RobotContacts, len(contacts.Values))
	for i, value := range contacts.Values {
		fid, err := id.Decrypt(value.FriendId, []byte(c.IDSecret))
		if err != nil {
			return err
		}

		contactsData[i] = &dao.RobotContacts{
			UserID:      userid,
			FriendID:    fid,
			FNickname:   value.FNickname,
			FHeadimg:    value.FHeadimg,
			FSex:        int8(value.FSex),
			Remark:      value.Remark,
			Mute:        int8(value.Mute),
			StickyOnTop: int8(value.StickyOnTop),
			Type:        int8(value.Type),
			Topic:       value.Topic,
			Status:      0,
			Version:     value.Version,
		}

		if value.Version > maxVersion {
			maxVersion = value.Version
		}
	}

	if err := dao.CreateOrUpdateRobotContacts(contactsData...); err != nil {
		return err
	}

	pageCount := math.Ceil(float64(contacts.RecordCount) / float64(contacts.Size))
	if pageCount > float64(page) {
		page++
		goto reload
	}
	return nil
}

func doContactsRequest(peer session.Peer, page int32, ver int64, c RobotConfig) (*pb.User_Contacts_ListReply, error) {
	userid, ok := peer.Params("UserID")
	if !ok {
		return nil, errors.New("invalid UserID")
	}

	agent, ok := peer.Params("AgentHttp")
	if !ok {
		return nil, errors.New("invalid agent addr")
	}

	tk, ok := peer.Params("Token")
	if !ok {
		return nil, errors.New("invalid token")
	}

	frame := &pb.User_Contacts_FriendRequestList{
		UserId:  userid.(string),
		Page:    page,
		Size:    50,
		Version: ver,
	}

	pm := jsonpb.Marshaler{}
	data, err := pm.MarshalToString(frame)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	body, errcode := RequestPostWithContext(ctx, command.UserContactsList, agent.(string)+"/v1/user", []byte(data), []byte(c.CommandSecret), tk.(string), c.CSRFSecret)
	if errcode != nil {
		return nil, errcode.ToError()
	}

	var frameW pb.User_Contacts_ListReply
	{
		if err := jsonpb.UnmarshalString(string(body), &frameW); err != nil {
			log.Error(Logger()).Log("action", "doAddFriendRequest", "error", err.Error(), "body", string(body))
			return nil, err
		}
	}

	return &frameW, nil
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
