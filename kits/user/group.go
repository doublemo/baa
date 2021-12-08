package user

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/doublemo/baa/cores/crypto/id"
	log "github.com/doublemo/baa/cores/log/level"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/helper"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/internal/worker"
	"github.com/doublemo/baa/kits/user/dao"
	"github.com/doublemo/baa/kits/user/errcode"
	"github.com/golang/protobuf/jsonpb"
	grpcproto "github.com/golang/protobuf/proto"
)

// GroupConfig 群配置
type GroupConfig struct {
	// IDSecret 用户ID 加密key 16位
	IDSecret string `alias:"idSecret" default:"7581BDD8E8DA3839"`

	// GroupSecret 群ID 加密key 16位
	GroupSecret string `alias:"idSecret" default:"7581BDD8E8DA3839"`

	// NameMaxLength 昵称最大长度
	NameMaxLength int `alias:"nameMaxLength" default:"34"`
}

func group(req *corespb.Request, c GroupConfig) (*corespb.Response, error) {
	var frame pb.User_Group_Request
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

	switch payload := frame.Payload.(type) {
	case *pb.User_Group_Request_Accept:
		return acceptGroup(req, payload, c)
	}

	return nil, errors.New("notsupported")
}

func acceptGroup(req *corespb.Request, frame *pb.User_Group_Request_Accept, c GroupConfig) (*corespb.Response, error) {
	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	useridstr, ok := req.Header["UserID"]
	if !ok {
		return errcode.Bad(w, errcode.ErrUserNotfound), nil
	}

	userid, err := strconv.ParseUint(useridstr, 10, 64)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	fuserid, err := id.Decrypt(frame.Accept.UserId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	if userid == fuserid {
		return errcode.Bad(w, errcode.ErrUserNotfound), nil
	}

	fgroupid, err := id.Decrypt(frame.Accept.GroupId, []byte(c.GroupSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrGroupIsNotFound, err.Error()), nil
	}

	group, err := dao.FindGroupsByGroupID(fgroupid)
	if err != nil {
		return errcode.Bad(w, errcode.ErrGroupIsNotFound, err.Error()), nil
	}

	// 查看有没有被邀请过
	member, err := dao.FindGroupsMemberByGroupIDAndUserID(group.ID, userid, "status")
	if err != nil {
		return errcode.Bad(w, errcode.ErrGroupIsNotFound, err.Error()), nil
	}

	resp := &pb.User_Group_Notify{
		Payload: &pb.User_Group_Notify_GroupRequest{
			GroupRequest: &pb.User_Group_Info{
				ID:      frame.Accept.GroupId,
				Name:    group.Name,
				Notice:  group.Notice,
				Headimg: group.Notice,
			},
		},
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

	if member.Status == dao.GroupMembersStatusNormal {
		w.Payload = &corespb.Response_Content{Content: respBytes}
		return w, nil
	}

	if err := dao.UpdateGroupMembersStatus(group.ID, dao.GroupMembersStatusNormal, userid); err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	// 添加群到通讯录
	contact := &dao.Contacts{
		UserID:    group.UserID,
		FriendID:  group.ID,
		FNickname: group.Name,
		FHeadimg:  group.Headimg,
		FSex:      0,
		Type:      dao.ContactsTypeGroup,
		Topic:     group.Topic,
		Status:    0,
		Version:   time.Now().Unix(),
	}

	dao.CreateContacts(contact)
	w.Payload = &corespb.Response_Content{Content: respBytes}
	return w, nil
}

func inviteAddGroup(req *corespb.Request, frame *pb.User_Group_Request_Invite, c GroupConfig) (*corespb.Response, error) {
	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	useridstr, ok := req.Header["UserID"]
	if !ok {
		return errcode.Bad(w, errcode.ErrUserNotfound), nil
	}

	userid, err := strconv.ParseUint(useridstr, 10, 64)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	fuserid, err := id.Decrypt(frame.Invite.UserId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	if userid != fuserid || len(frame.Invite.FriendId) < 1 {
		return errcode.Bad(w, errcode.ErrUserNotfound), nil
	}

	im, err := findRequestHeaderIMServer(req, userid)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	friends := make([]uint64, len(frame.Invite.FriendId))
	for i, value := range frame.Invite.FriendId {
		friendid, err := id.Decrypt(value, []byte(c.IDSecret))
		if err != nil {
			return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
		}
		friends[i] = friendid
	}

	fgroupid, err := id.Decrypt(frame.Invite.GroupId, []byte(c.GroupSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrGroupIsNotFound, err.Error()), nil
	}

	group, err := dao.FindGroupsByGroupID(fgroupid)
	if err != nil {
		return errcode.Bad(w, errcode.ErrGroupIsNotFound, err.Error()), nil
	}

	// 查看有没有被邀请过
	members, err := dao.FindGroupsMembersNotNormalByGroupIDAndUserID(group.ID, friends, "status")
	if err != dao.ErrRecordNotFound && err != nil {
		return errcode.Bad(w, errcode.ErrGroupIsNotFound, err.Error()), nil
	}

	resp := &pb.User_Group_Reply{
		OK: true,
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

	invited := make(map[uint64]*dao.GroupMembers)
	for _, member := range members {
		invited[member.UserID] = member
	}

	users, err := dao.FindContactsByUserID(userid, friends, "id", "user_id", "friend_id", "f_nickname", "f_headimg", "f_sex", "type", "topic")
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	contacts := make(map[uint64]*dao.Contacts)
	for _, u := range users {
		contacts[u.FriendID] = u
	}

	gmembers := make([]*dao.GroupMembers, 0)
	messages := make([]*pb.IM_Msg_Content, 0)
	var message *pb.IM_Msg_Content
	for _, fid := range friends {
		contact, ok := contacts[fid]
		if !ok {
			continue
		}

		enid, _ := id.Encrypt(fid, []byte(c.IDSecret))
		message = &pb.IM_Msg_Content{
			SeqID:  fid,
			To:     enid,
			From:   frame.Invite.UserId,
			Group:  pb.IM_Msg_ToC,
			Topic:  contact.Topic,
			SendAt: time.Now().Unix(),
			Origin: pb.IM_Msg_OriginSystem,
		}

		if m, ok := invited[fid]; ok {
			if m.Status == dao.GroupMembersStatusNormal {
				continue
			}

			messages = append(messages, message)
			continue
		}

		messages = append(messages, message)
		gmembers = append(gmembers, &dao.GroupMembers{
			GroupID:       group.ID,
			UserID:        contact.FriendID,
			Nickname:      contact.FNickname,
			Headimg:       contact.FHeadimg,
			Sex:           contact.FSex,
			Topic:         group.Topic,
			OfficialTitle: dao.GroupMemberOfficialTitleNone,
			Status:        dao.GroupMembersStatusInvitationNotSent,
			JoinAt:        0,
			Origin:        dao.GroupMembersOriginInvite,
			Handler:       userid,
			Version:       time.Now().Unix(),
		})
	}

	if err := dao.CreateGroupsMember(group.ID, gmembers...); err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	// 发送邀请信息
	// 从聊天信息发送
	gid, _ := id.Encrypt(group.ID, []byte(c.GroupSecret))
	waitSendMssages := make([]*pb.IM_Msg_Content, 0)
	i := 0
	for _, message := range messages {
		message.Payload = &pb.IM_Msg_Content_JoinGroupInvite{
			JoinGroupInvite: &pb.IM_Msg_ContentType_JoinGroupInvite{
				GroupID: gid,
				Name:    group.Name,
				Headimg: group.Headimg,
			},
		}

		waitSendMssages = append(waitSendMssages, message)
		i++
		if i >= 10 {
			worker.Submit(makeJoinGroupInvite(im, frame.Invite.GroupId, group.ID, waitSendMssages...))
			waitSendMssages = make([]*pb.IM_Msg_Content, 0)
			i = 0
		}
	}

	if len(waitSendMssages) > 0 {
		worker.Submit(makeJoinGroupInvite(im, frame.Invite.GroupId, group.ID, waitSendMssages...))
	}

	w.Payload = &corespb.Response_Content{Content: respBytes}
	return w, nil
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

func groupCreate(req *corespb.Request, c GroupConfig) (*corespb.Response, error) {
	var frame pb.User_Group_Create_Request
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

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	if len(frame.Members) < 1 {
		return errcode.Bad(w, errcode.ErrGroupMembersLessThen1), nil
	}

	useridString, ok := req.Header["UserID"]
	if !ok {
		return errcode.Bad(w, errcode.ErrInvalidUserID), nil
	}

	userid, err := strconv.ParseUint(useridString, 10, 64)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInvalidUserID, err.Error()), nil
	}

	muserid, err := id.Decrypt(frame.UserId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInvalidUserID, err.Error()), nil
	}

	if userid != muserid {
		return errcode.Bad(w, errcode.ErrInvalidUserID), nil
	}

	im, err := findRequestHeaderIMServer(req, userid)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	muser, err := dao.FindUsersByID(userid, "id", "nickname", "headimg", "sex")
	if err != nil {
		return errcode.Bad(w, errcode.ErrInvalidUserID, err.Error()), nil
	}

	membersUsers := make([]uint64, len(frame.Members))
	for i, value := range frame.Members {
		m, err := id.Decrypt(value, []byte(c.IDSecret))
		if err != nil {
			return errcode.Bad(w, errcode.ErrGroupMemnersIDIncorrect, err.Error()), nil
		}

		membersUsers[i] = m
	}

	users, err := dao.FindContactsByUserID(userid, membersUsers, "id", "user_id", "friend_id", "f_nickname", "f_headimg", "f_sex", "type", "topic")
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	if len(users) < 1 {
		return errcode.Bad(w, errcode.ErrGroupMemnersIDIncorrect), nil
	}

	sid, err := getSNID(1)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	if len(sid) < 1 {
		return errcode.Bad(w, errcode.ErrInternalServer), nil
	}

	group := &dao.Groups{
		ID:      sid[0],
		Name:    "",
		Headimg: "",
		Topic:   helper.GenerateTopic(sid[0]),
		UserID:  userid,
	}

	gid, err := id.Encrypt(group.ID, []byte(c.GroupSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	members := make([]*dao.GroupMembers, 0)
	name := []string{muser.Nickname}
	membersMap := make(map[uint64]bool)
	version := time.Now().Unix()
	messages := make([]*pb.IM_Msg_Content, 0)
	for _, contact := range users {
		if membersMap[contact.FriendID] || contact.Type != dao.ContactsTypePerson {
			continue
		}

		name = append(name, contact.FNickname)
		membersMap[contact.FriendID] = true
		members = append(members, &dao.GroupMembers{
			GroupID:       group.ID,
			UserID:        contact.FriendID,
			Nickname:      contact.FNickname,
			Headimg:       contact.FHeadimg,
			Sex:           contact.FSex,
			Topic:         group.Topic,
			OfficialTitle: dao.GroupMemberOfficialTitleNone,
			Status:        dao.GroupMembersStatusInvitationNotSent,
			JoinAt:        0,
			Origin:        dao.GroupMembersOriginInvite,
			Handler:       userid,
			Version:       version,
		})

		enid, err := id.Encrypt(contact.FriendID, []byte(c.IDSecret))
		if err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		messages = append(messages, &pb.IM_Msg_Content{
			SeqID:  0,
			To:     enid,
			From:   frame.UserId,
			Group:  pb.IM_Msg_ToC,
			Topic:  contact.Topic,
			SendAt: time.Now().Unix(),
			Origin: pb.IM_Msg_OriginSystem,
		})
	}

	if len(members) < 1 {
		return errcode.Bad(w, errcode.ErrGroupMemnersIDIncorrect), nil
	}

	members = append(members, &dao.GroupMembers{
		GroupID:       group.ID,
		UserID:        muser.ID,
		Nickname:      muser.Nickname,
		Headimg:       muser.Headimg,
		Sex:           muser.Sex,
		Topic:         group.Topic,
		OfficialTitle: dao.GroupMemberOfficialTitleOwner,
		Status:        dao.GroupMembersStatusNormal,
		JoinAt:        time.Now().Unix(),
		Origin:        dao.GroupMembersOriginVolunteer,
		Handler:       userid,
		Version:       version,
	})

	// 生成群头像
	// 生成群名称
	names := strings.Join(name, ",")
	if len(names) > c.NameMaxLength {
		names = names[0:c.NameMaxLength]
	}

	group.Name = names
	if err := dao.CreateAndJoinGroup(group, members...); err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	seqidmaps := make(map[string]uint64)
	for _, member := range members {
		enid, err := id.Encrypt(member.UserID, []byte(c.IDSecret))
		if err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}
		seqidmaps[enid] = member.UserID
	}

	// 发送邀请信息
	// 从聊天信息发送
	waitSendMssages := make([]*pb.IM_Msg_Content, 0)
	i := 0
	for _, message := range messages {
		message.SeqID = seqidmaps[message.To]
		message.Payload = &pb.IM_Msg_Content_JoinGroupInvite{
			JoinGroupInvite: &pb.IM_Msg_ContentType_JoinGroupInvite{
				GroupID: gid,
				Name:    group.Name,
				Headimg: group.Headimg,
			},
		}

		waitSendMssages = append(waitSendMssages, message)
		i++
		if i >= 10 {
			worker.Submit(makeJoinGroupInvite(im, useridString, group.ID, waitSendMssages...))
			waitSendMssages = make([]*pb.IM_Msg_Content, 0)
			i = 0
		}
	}

	if len(waitSendMssages) > 0 {
		worker.Submit(makeJoinGroupInvite(im, useridString, group.ID, waitSendMssages...))
	}

	resp := &pb.User_Group_Create_Reply{
		Info: &pb.User_Group_Info{
			ID:      gid,
			Name:    group.Name,
			Notice:  "",
			Headimg: group.Headimg,
		},
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

func makeJoinGroupInvite(im, userid string, gid uint64, messages ...*pb.IM_Msg_Content) func() {
	return func() {
		acked, err := sendChatMessages(im, userid, messages...)
		if err != nil {
			log.Error(Logger()).Log("action", "sendChatMessages", "error", err.Error())
			return
		}

		ids := make([]uint64, len(acked))
		for i, ack := range acked {
			ids[i] = ack.SeqID
		}

		if len(ids) < 1 {
			return
		}

		dao.UpdateGroupMembersStatus(gid, dao.GroupMembersStatusWaitingJoin, ids...)
	}
}

func findRequestHeaderIMServer(req *corespb.Request, userid uint64) (string, error) {
	if req.Header == nil {
		return "", errors.New("request header is nil")
	}

	im, ok := req.Header["IMServer"]
	if !ok {
		servers, err := getUserServers(userid)
		if err != nil {
			return "", err
		}

		server, ok := servers[userid]
		if !ok {
			return "", errors.New("im server 0")
		}

		if m, ok := server[kit.IMServiceName]; ok {
			endpoints, err := sd.GetEndpointsByID(m)
			if err != nil {
				return "", err
			}

			if m0, ok := endpoints[m]; ok && m0 != nil {
				im = m0.Addr()
			}
		}
	}

	if im == "" {
		return "", errors.New("Could not find im server")
	}

	return im, nil
}
