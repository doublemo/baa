package user

import (
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

	im, ok := req.Header["IMServer"]
	if !ok {
		servers, err := getUserServers(userid)
		if err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		server, ok := servers[userid]
		if !ok {
			return errcode.Bad(w, errcode.ErrInvalidUserID, "im server 0"), nil
		}

		if m, ok := server[kit.IMServiceName]; ok {
			endpoints, err := sd.GetEndpointsByID(m)
			if err != nil {
				return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
			}

			if m0, ok := endpoints[m]; ok && m0 != nil {
				im = m0.Addr()
			}
		}

		if im == "" {
			return errcode.Bad(w, errcode.ErrInvalidUserID, "im server 1"), nil
		}
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
			Group:  pb.IM_Msg_ToG,
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
		for _, ack := range acked {
			ids = append(ids, ack.SeqID)
		}

		dao.UpdateGroupMembersStatus(gid, dao.GroupMembersStatusWaitingJoin, ids...)
	}
}
