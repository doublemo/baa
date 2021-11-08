package user

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/doublemo/baa/cores/crypto/id"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/kits/user/dao"
	"github.com/doublemo/baa/kits/user/errcode"
	"github.com/golang/protobuf/jsonpb"
	grpcproto "github.com/golang/protobuf/proto"
)

func contact(req *corespb.Request, c UserConfig) (*corespb.Response, error) {
	var frame pb.User_Contacts_Request
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
	case *pb.User_Contacts_Request_Add:
		return addContact(req, payload, c)

	case *pb.User_Contacts_Request_Accept:
		return acceptContact(req, payload, c)

	case *pb.User_Contacts_Request_Refuse:
		return refuseContact(req, payload, c)

	case *pb.User_Contacts_Request_Cancel:
		return cancelContact(req, payload, c)

	case *pb.User_Contacts_Request_SearchFriend:
		return searchFriend(req, payload, c)
	}

	return nil, errors.New("notsupported")
}

func addContact(req *corespb.Request, frame *pb.User_Contacts_Request_Add, c UserConfig) (*corespb.Response, error) {
	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	var userId string
	if m, ok := req.Header["UserId"]; ok {
		userId = m
	}

	if userId != frame.Add.UserId {
		return errcode.Bad(w, errcode.ErrUserNotfound, frame.Add.UserId), nil
	}

	if utf8.RuneCount([]byte(frame.Add.Message)) > 30 {
		return errcode.Bad(w, errcode.ErrMessageTooLong), nil
	}

	uid, err := id.Decrypt(frame.Add.UserId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	friendId, err := id.Decrypt(frame.Add.FriendId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	if _, err = dao.FindContactsByUserIDAndFriendID(uid, friendId, "id"); err == nil {
		return errcode.Bad(w, errcode.ErrUserAlreadyContact), nil
	}

	resp := &pb.User_Contacts_Reply{OK: true}
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

	// 检查对方信箱里面有我的请求
	contactsRequest, err := dao.FindContactsRequestByUserIDAndFriendID(friendId, uid)
	if err == nil {
		if contactsRequest.Status == 0 {
			return w, nil
		}

		// 如果有直接删除
		if err := dao.DeleteContactsRequestByUserIDAndFriendID(friendId, uid); err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}
	}

	user, err := dao.FindUsersByID(uid, "id", "nickname", "heading", "sex")
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	friend, err := dao.FindUsersByID(friendId, "id", "nickname", "heading", "sex")
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	// 将请求发送到对方信息
	request := dao.ContactsRequest{
		UserID:    friend.ID,
		FriendID:  uid,
		FNickname: user.Nickname,
		FHeadimg:  user.Headimg,
		FSex:      user.Sex,
		Remark:    frame.Add.Remark,
		Messages:  "A:" + frame.Add.Message,
		Status:    0,
		FromID:    uid,
	}

	if err := dao.CreateContactsRequest(&request); err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	return w, nil
}

func acceptContact(req *corespb.Request, frame *pb.User_Contacts_Request_Accept, c UserConfig) (*corespb.Response, error) {
	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	var userId string
	if m, ok := req.Header["UserId"]; ok {
		userId = m
	}

	if userId != frame.Accept.UserId {
		return errcode.Bad(w, errcode.ErrUserNotfound, frame.Accept.UserId), nil
	}

	uid, err := id.Decrypt(frame.Accept.UserId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	friendId, err := id.Decrypt(frame.Accept.FriendId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	contactsRequest, err := dao.FindContactsRequestByUserIDAndFriendID(uid, friendId)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	if contactsRequest.Status != 0 {
		return errcode.Bad(w, errcode.ErrContactsRequestExpired), nil
	}

	at := time.Now().Sub(time.Unix(contactsRequest.CreatedAt, 0))
	if at.Hours() > float64(c.ContactRequestExpireAt) {
		dao.UpdateContactsRequestStatusByID(uid, friendId, -1)
		return errcode.Bad(w, errcode.ErrContactsRequestExpired), nil
	}

	user, err := dao.FindUsersByID(contactsRequest.UserID)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	friend, err := dao.FindUsersByID(contactsRequest.FriendID)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	resp := &pb.User_Contacts_Reply{OK: true}
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
	if err := dao.AddContactsFromRequest(user, friend, contactsRequest, frame.Accept.Remark); err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}
	return w, nil
}

func refuseContact(req *corespb.Request, frame *pb.User_Contacts_Request_Refuse, c UserConfig) (*corespb.Response, error) {
	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	var userId string
	if m, ok := req.Header["UserId"]; ok {
		userId = m
	}

	if userId != frame.Refuse.UserId {
		return errcode.Bad(w, errcode.ErrUserNotfound, frame.Refuse.UserId), nil
	}

	uid, err := id.Decrypt(frame.Refuse.UserId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	friendId, err := id.Decrypt(frame.Refuse.FriendId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	contactsRequest, err := dao.FindContactsRequestByUserIDAndFriendID(uid, friendId)
	if err != nil {
		return errcode.Bad(w, errcode.ErrContactsRequestNotFound, err.Error()), nil
	}

	if contactsRequest.Status < 0 {
		return errcode.Bad(w, errcode.ErrContactsRequestExpired), nil
	}

	user, err := dao.FindUsersByID(contactsRequest.UserID)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	friend, err := dao.FindUsersByID(contactsRequest.FriendID)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	resp := &pb.User_Contacts_Reply{OK: true}
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
	if frame.Refuse.Message == "" {
		dao.UpdateContactsRequestStatusByID(uid, friendId, 1)
		return w, nil
	}

	if contactsRequest.Messages == "" {
		contactsRequest.Messages = "B:" + frame.Refuse.Message
	} else {
		contactsRequest.Messages += "|B:" + frame.Refuse.Message
	}

	if err := dao.RefuseAddContact(user, friend, contactsRequest); err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}
	return w, nil
}

func cancelContact(req *corespb.Request, frame *pb.User_Contacts_Request_Cancel, c UserConfig) (*corespb.Response, error) {
	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	var userId string
	if m, ok := req.Header["UserId"]; ok {
		userId = m
	}

	if userId != frame.Cancel.UserId {
		return errcode.Bad(w, errcode.ErrUserNotfound, frame.Cancel.UserId), nil
	}

	uid, err := id.Decrypt(frame.Cancel.UserId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	friendId, err := id.Decrypt(frame.Cancel.FriendId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	resp := &pb.User_Contacts_Reply{OK: true}
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
	dao.DeleteContactsRequestByUserIDAndFriendID(uid, friendId)
	//dao.DeleteContactsRequestByUserIDAndFriendID(friendId, uid)
	return w, nil
}

func searchFriend(req *corespb.Request, frame *pb.User_Contacts_Request_SearchFriend, c UserConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	users, err := dao.FindUsersByIndexNo(frame.SearchFriend)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	resp := &pb.User_Info{
		Nickname: users.Nickname,
		Headimg:  users.Headimg,
		Age:      int32(users.Age),
		Sex:      int32(users.Sex),
	}

	resp.UserId, _ = id.Encrypt(users.ID, []byte(c.IDSecret))
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

func friendRequestList(req *corespb.Request, c UserConfig) (*corespb.Response, error) {
	var frame pb.User_Contacts_FriendRequestList
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

	var userId string
	if m, ok := req.Header["UserId"]; ok {
		userId = m
	}

	if userId != frame.UserId {
		return errcode.Bad(w, errcode.ErrUserNotfound, frame.UserId), nil
	}

	uid, err := id.Decrypt(frame.UserId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}
	list, count, err := dao.FindContactsRequestByUserID(uid, frame.Page, frame.Size, frame.Version)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.User_Contacts_FriendRequestListReply{
		Values:      make([]*pb.User_Contacts_FriendRequestInfo, len(list)),
		Page:        frame.Page,
		Size:        frame.Size,
		RecordCount: int32(count),
	}

	for k, v := range list {
		enuid, _ := id.Encrypt(v.UserID, []byte(c.IDSecret))
		enfid, _ := id.Encrypt(v.FriendID, []byte(c.IDSecret))
		resp.Values[k] = &pb.User_Contacts_FriendRequestInfo{
			UserId:    enuid,
			FriendId:  enfid,
			FNickname: v.FNickname,
			FHeadimg:  v.FHeadimg,
			FSex:      int32(v.FSex),
			Remark:    v.Remark,
			Messages:  v.Messages,
			Status:    int32(v.Status),
			Version:   v.Version,
			CreatedAt: v.CreatedAt,
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

func contacts(req *corespb.Request, c UserConfig) (*corespb.Response, error) {
	var frame pb.User_Contacts_ListRequest
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

	var userId string
	if m, ok := req.Header["UserId"]; ok {
		userId = m
	}

	if userId != frame.UserId {
		return errcode.Bad(w, errcode.ErrUserNotfound, frame.UserId), nil
	}

	uid, err := id.Decrypt(frame.UserId, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	list, count, err := dao.FindContactsByUserID(uid, frame.Page, frame.Size, frame.Version)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.User_Contacts_ListReply{
		Values:      make([]*pb.User_Contacts_Info, len(list)),
		Page:        frame.Page,
		Size:        frame.Size,
		RecordCount: int32(count),
	}

	for k, v := range list {
		enfid, _ := id.Encrypt(v.FriendID, []byte(c.IDSecret))
		resp.Values[k] = &pb.User_Contacts_Info{
			FriendId:    enfid,
			FNickname:   v.FNickname,
			FHeadimg:    v.FHeadimg,
			FSex:        int32(v.FSex),
			Remark:      v.Remark,
			Mute:        int32(v.Mute),
			StickyOnTop: int32(v.StickyOnTop),
			Type:        int32(v.Type),
			Topic:       v.Topic,
			Version:     v.Version,
			CreatedAt:   v.CreateAt,
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

func checkIsMyFriend(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.User_Contacts_IsFriend
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	resp := pb.User_Contacts_Reply{
		OK: false,
	}

	bytes, _ := grpcproto.Marshal(&resp)
	w.Payload = &corespb.Response_Content{Content: bytes}
	contacts, err := dao.FindContactsByUserIDAndTopic(frame.UserId, frame.Topic, "friend_id", "status")
	if err != nil {
		return w, nil
	}

	if contacts.FriendID != frame.FriendId || contacts.Status != 0 {
		return w, nil
	}

	resp.OK = true
	bytes, _ = grpcproto.Marshal(&resp)
	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
}
