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
	if router.IsHTTP(req) {
		return contactToHTTP(req, c)
	}

	var frame pb.User_Contacts_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
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

func contactToHTTP(req *corespb.Request, c UserConfig) (resp *corespb.Response, err error) {
	var frame pb.User_Contacts_Request
	{
		if err := jsonpb.UnmarshalString(string(req.Payload), &frame); err != nil {
			return nil, err
		}
	}

	switch payload := frame.Payload.(type) {
	case *pb.User_Contacts_Request_Add:
		resp, err = addContact(req, payload, c)

	case *pb.User_Contacts_Request_Accept:
		resp, err = acceptContact(req, payload, c)

	case *pb.User_Contacts_Request_Refuse:
		resp, err = refuseContact(req, payload, c)

	case *pb.User_Contacts_Request_Cancel:
		resp, err = cancelContact(req, payload, c)

	case *pb.User_Contacts_Request_SearchFriend:
		resp, err = searchFriend(req, payload, c)

	default:
		return nil, errors.New("notsupported")
	}

	switch w := resp.Payload.(type) {
	case *corespb.Response_Content:
		var frame pb.User_Contacts_Reply
		grpcproto.Unmarshal(w.Content, &frame)
		jsonpbM := &jsonpb.Marshaler{}
		b, _ := jsonpbM.MarshalToString(&frame)
		resp.Payload = &corespb.Response_Content{Content: []byte(b)}
	case *corespb.Response_Error:
		return
	}
	return
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

	if _, err = dao.FindContactsByUserIDAndFriendID(uid, friendId); err == nil {
		return errcode.Bad(w, errcode.ErrUserAlreadyContact), nil
	}

	resp := &pb.User_Contacts_Reply{OK: true}
	respBytes, _ := grpcproto.Marshal(resp)
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

	user, err := dao.FindUsersByID(uid)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUserNotfound, err.Error()), nil
	}

	friend, err := dao.FindUsersByID(friendId)
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
	respBytes, _ := grpcproto.Marshal(resp)
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
	respBytes, _ := grpcproto.Marshal(resp)
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
	respBytes, _ := grpcproto.Marshal(resp)
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
	respBytes, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: respBytes}
	return w, nil
}
