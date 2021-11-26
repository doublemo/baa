package im

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/doublemo/baa/cores/crypto/id"
	log "github.com/doublemo/baa/cores/log/level"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/im/cache"
	"github.com/doublemo/baa/kits/im/dao"
	"github.com/doublemo/baa/kits/im/errcode"
	"github.com/doublemo/baa/kits/im/mime"
	"github.com/doublemo/baa/kits/im/worker"
	grpcproto "github.com/golang/protobuf/proto"
)

// ChatConfig 聊天配置
type ChatConfig struct {
	// IDSecret 用户ID 加密key 16位
	IDSecret string `alias:"idSecret" default:"7581BDD8E8DA3839"`
}

func send(req *corespb.Request, c ChatConfig) (*corespb.Response, error) {
	var frame pb.IM_Send
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	if req.Header == nil {
		return nil, errors.New("Header is nil")
	}

	// 如果两条以上的消息，需要排序
	if len(frame.Messages.Values) > 1 {
		sort.Slice(frame.Messages.Values, func(i, j int) bool {
			return frame.Messages.Values[i].SeqID < frame.Messages.Values[j].SeqID
		})
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	var userid uint64
	if uid, ok := req.Header["UserID"]; ok {
		if m, err := strconv.ParseUint(uid, 10, 64); err == nil {
			userid = m
		}
	}

	ret := &pb.IM_Msg_AckListReceived{
		Successed: make([]*pb.IM_Msg_AckReceived, 0),
		Failed:    make([]*pb.IM_Msg_AckFailed, 0),
	}

	for _, message := range frame.Messages.Values {
		received, failed := sendto(userid, message, c)
		if received != nil {
			ret.Successed = append(ret.Successed, received)
		}

		if failed != nil {
			ret.Failed = append(ret.Failed, failed)
		}
	}

	bytes, _ := grpcproto.Marshal(&pb.IM_Notify{Payload: &pb.IM_Notify_Received{Received: ret}})
	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
}

func sendto(userid uint64, frame *pb.IM_Msg_Content, c ChatConfig) (*pb.IM_Msg_AckReceived, *pb.IM_Msg_AckFailed) {
	// 消息安全检查
	if !validationMsg(frame) {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      frame.SeqID,
			ErrCode:    errcode.ErrInvalidCharacter.Code(),
			ErrMessage: errcode.ErrInvalidCharacter.Error(),
		}
	}

	var err error
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	frame.Id, err = cache.GetSnowflakeID(ctx)
	if err != nil {
		cancel()
		log.Error(Logger()).Log("action", "getSnowflakeID", "error", err)
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      frame.SeqID,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}
	cancel()

	msg, err := makeMessage(frame, []byte(c.IDSecret))
	if err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      frame.SeqID,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}

	if msg.From != userid {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      frame.SeqID,
			ErrCode:    errcode.ErrInvalidUserIdToken.Code(),
			ErrMessage: errcode.ErrInvalidUserIdToken.Error(),
		}
	}

	if frame.Group == pb.IM_Msg_ToG {
		return sendtoG(msg, c)
	}

	return sendtoC(msg, c)
}

func sendtoC(msg *dao.Messages, c ChatConfig) (*pb.IM_Msg_AckReceived, *pb.IM_Msg_AckFailed) {
	// todo 检查是否彼此为好友
	if ok, err := checkIsMyFriend(msg.From, msg.To, msg.Topic); !ok || err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      msg.SeqId,
			ErrCode:    errcode.ErrNotFriend.Code(),
			ErrMessage: errcode.ErrNotFriend.Error(),
		}
	}

	timelines, err := getTimelines(false, msg.From, msg.To)
	if err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      msg.SeqId,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}

	ttid, ok1 := timelines[msg.To]
	ftid, ok2 := timelines[msg.From]
	fmt.Println(ttid, ftid, ok1, ok2)
	if !ok1 || !ok2 {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      msg.SeqId,
			ErrCode:    errcode.ErrInvalidUserIdToken.Code(),
			ErrMessage: errcode.ErrInvalidUserIdToken.Error(),
		}
	}

	msg.TSeqId = ttid
	msg.FSeqId = ftid

	// 开始存储数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := dao.WriteInboxC(ctx, msg); err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      msg.SeqId,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}

	// 消息送检
	msgInspectionReport(msg)

	// 推送信息
	worker.Submit(func() {
		data, err := makeBroadcastMessages([]byte(c.IDSecret), *msg)
		if err != nil {
			log.Error(Logger()).Log("action", "makeBroadcastMessages", "error", err)
		}

		for addr, messages := range data {
			if err := pushMessages(addr, messages...); err != nil {
				log.Error(Logger()).Log("action", "pushMessages", "error", err)
			}
		}
	})

	return &pb.IM_Msg_AckReceived{
		Id:       msg.ID,
		SeqID:    msg.SeqId,
		NewSeqID: ttid,
	}, nil
}

func sendtoG(msg *dao.Messages, c ChatConfig) (*pb.IM_Msg_AckReceived, *pb.IM_Msg_AckFailed) {
	// todo 检查是否在群中
	if ok, err := checkInGroup(msg.From, msg.Topic); !ok || err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      msg.SeqId,
			ErrCode:    errcode.ErrNotFriend.Code(),
			ErrMessage: errcode.ErrNotFriend.Error(),
		}
	}

	timelines, err := getTimelines(false, msg.To)
	if err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      msg.SeqId,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}

	ttid, ok := timelines[msg.To]
	if !ok {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      msg.SeqId,
			ErrCode:    errcode.ErrInvalidUserIdToken.Code(),
			ErrMessage: errcode.ErrInvalidUserIdToken.Error(),
		}
	}

	msg.TSeqId = ttid
	msg.FSeqId = ttid

	// 开始存储数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := dao.WriteInboxG(ctx, msg); err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      msg.SeqId,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}

	// 消息送检
	msgInspectionReport(msg)

	gid, _ := id.Encrypt(msg.To, []byte(c.IDSecret))

	// 推送信息
	fn := func(users []uint64) func() {
		return func() {
			data, err := makeGroupBroadcastMessages(*msg, []byte(c.IDSecret), users...)
			if err != nil {
				log.Error(Logger()).Log("action", "makeBroadcastMessages", "error", err)
			}

			for addr, message := range data {
				if err := pushMessages(addr, message); err != nil {
					log.Error(Logger()).Log("action", "pushMessages", "error", err)
				}
			}
		}
	}
	var (
		page      int32 = 0
		size      int32 = 100
		pageCount int32 = 0
	)
loop:
	page++
	members, count, err := groupMembersID(gid, page, size)
	if err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      msg.SeqId,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}

	if pageCount == 0 {
		pageCount = int32(math.Ceil(float64(count) / float64(size)))
	}

	if len(members) > 0 {
		worker.Submit(fn(members))
	}

	if page <= pageCount {
		goto loop
	}

	return &pb.IM_Msg_AckReceived{
		Id:       msg.ID,
		SeqID:    msg.SeqId,
		NewSeqID: ttid,
	}, nil
}

func validationMsg(msg *pb.IM_Msg_Content) bool {
	return true
}

func makeMessage(frame *pb.IM_Msg_Content, secret []byte) (*dao.Messages, error) {
	tid, err := id.Decrypt(frame.To, []byte(secret))
	if err != nil {
		return nil, err
	}

	fid, err := id.Decrypt(frame.From, []byte(secret))
	if err != nil {
		return nil, err
	}

	msg := &dao.Messages{
		ID:    frame.Id,
		SeqId: frame.SeqID,
		To:    tid,
		From:  fid,
		Group: int32(frame.Group),
		Topic: frame.Topic,
	}

	var content []byte
	switch payload := frame.Payload.(type) {
	case *pb.IM_Msg_Content_Text:
		content, err = []byte(payload.Text.Content), nil
		msg.ContentType = mime.Text
	case *pb.IM_Msg_Content_Image:
		content, err = json.Marshal(payload)
		msg.ContentType = mime.Image
	case *pb.IM_Msg_Content_Video:
		content, err = json.Marshal(payload)
		msg.ContentType = mime.Video
	case *pb.IM_Msg_Content_Voice:
		content, err = json.Marshal(payload)
		msg.ContentType = mime.Voice
	case *pb.IM_Msg_Content_File:
		content, err = json.Marshal(payload)
		msg.ContentType = mime.File
	case *pb.IM_Msg_Content_Emoticon:
		content, err = json.Marshal(payload)
		msg.ContentType = mime.Emoticon
	default:
		return nil, errors.New("")
	}

	if err != nil {
		return nil, err
	}

	msg.Content = string(content)
	return msg, nil
}

func gatherUsersAgent(users ...uint64) (map[uint64][]string, error) {
	status, err := getCacheUsersStatus(false, users...)
	if err != nil {
		return nil, err
	}

	servers := make(map[uint64]map[string]bool)
	for _, id := range users {
		if server, ok := findServersID(id, kit.AgentServiceName, status); ok {
			for _, s := range server {
				if _, ok := servers[id]; !ok {
					servers[id] = make(map[string]bool)
				}
				servers[id][s] = true
			}
		}
	}

	addrs := make(map[uint64][]string)
	for id, sers := range servers {
		addrs[id] = make([]string, 0)
		for addr := range sers {
			addrs[id] = append(addrs[id], addr)
		}
	}
	return addrs, nil
}

func pushMessages(addr string, msg ...*pb.Agent_BroadcastMessage) error {
	nc := nats.Conn()
	if nc == nil {
		return errors.New("conn is nil")
	}

	frame := &pb.Agent_Broadcast{Messages: msg}
	req := coresproto.RequestBytes{
		Cmd:    kit.Agent,
		SubCmd: command.AgentBroadcast,
		SeqID:  1,
	}

	req.Content, _ = grpcproto.Marshal(frame)
	bytes, err := req.Marshal()
	if err != nil {
		return err
	}

	if err := nc.Publish(addr, bytes); err != nil {
		return err
	}

	return nc.FlushTimeout(time.Second * 10)
}

func makeBroadcastMessages(idsecret []byte, msg ...dao.Messages) (map[string][]*pb.Agent_BroadcastMessage, error) {
	users := make([]uint64, len(msg))
	for i, m := range msg {
		users[i] = m.To
	}

	agents, err := gatherUsersAgent(users...)
	if err != nil {
		return nil, err
	}

	data := make(map[string][]*pb.Agent_BroadcastMessage)
	for _, m := range msg {
		addrs, ok := agents[m.To]
		if !ok {
			continue
		}

		frame, err := makeMessageToPB(m, idsecret)
		if err != nil {
			return nil, err
		}

		frame.SeqID = m.TSeqId
		bytes, _ := grpcproto.Marshal(&pb.IM_Notify{Payload: &pb.IM_Notify_List{List: &pb.IM_Msg_List{Values: []*pb.IM_Msg_Content{frame}}}})
		called := make(map[string]bool)
		for _, addr := range addrs {
			if _, ok := data[addr]; !ok {
				data[addr] = make([]*pb.Agent_BroadcastMessage, 0)
			}

			if called[addr] {
				continue
			}
			data[addr] = append(data[addr], &pb.Agent_BroadcastMessage{Receiver: []uint64{m.To}, Command: kit.IM.Int32(), SubCommand: command.IMPush.Int32(), Payload: bytes})
		}
	}

	return data, nil
}

func makeGroupBroadcastMessages(msg dao.Messages, idsecret []byte, users ...uint64) (map[string]*pb.Agent_BroadcastMessage, error) {
	agents, err := gatherUsersAgent(users...)
	if err != nil {
		return nil, err
	}

	frame, err := makeMessageToPB(msg, idsecret)
	if err != nil {
		return nil, err
	}

	frame.SeqID = msg.TSeqId
	bytes, _ := grpcproto.Marshal(&pb.IM_Notify{Payload: &pb.IM_Notify_List{List: &pb.IM_Msg_List{Values: []*pb.IM_Msg_Content{frame}}}})
	data := make(map[string]*pb.Agent_BroadcastMessage)
	for _, userid := range users {
		addrs, ok := agents[userid]
		if !ok {
			continue
		}

		called := make(map[string]bool)
		for _, addr := range addrs {
			if called[addr] {
				continue
			}

			called[addr] = true
			if _, ok := data[addr]; ok {
				data[addr].Receiver = append(data[addr].Receiver, userid)
			} else {
				data[addr] = &pb.Agent_BroadcastMessage{Receiver: []uint64{userid}, Command: kit.IM.Int32(), SubCommand: command.IMPush.Int32(), Payload: bytes}
			}
		}
	}
	return data, nil
}

func makeMessageToPB(m dao.Messages, secret []byte) (*pb.IM_Msg_Content, error) {
	frameMsg := pb.IM_Msg_Content{
		Id:     m.ID,
		SeqID:  m.SeqId,
		Group:  pb.IM_Msg_Group(m.Group),
		Topic:  m.Topic,
		SendAt: m.CreatedAt.Unix(),
	}
	frameMsg.To, _ = id.Encrypt(m.To, secret)
	frameMsg.From, _ = id.Encrypt(m.From, secret)

	switch m.ContentType {
	case mime.Text:
		frameMsg.Payload = &pb.IM_Msg_Content_Text{Text: &pb.IM_Msg_ContentType_Text{Content: m.Content}}

	case mime.Image:
		content := pb.IM_Msg_Content_Image{}
		if err := json.Unmarshal([]byte(m.Content), &content); err != nil {
			return nil, err
		}
		frameMsg.Payload = &content

	case mime.Video:
		content := pb.IM_Msg_Content_Video{}
		if err := json.Unmarshal([]byte(m.Content), &content); err != nil {
			return nil, err
		}
		frameMsg.Payload = &content

	case mime.Voice:
		content := pb.IM_Msg_Content_Voice{}
		if err := json.Unmarshal([]byte(m.Content), &content); err != nil {
			return nil, err
		}
		frameMsg.Payload = &content

	case mime.File:
		content := pb.IM_Msg_Content_File{}
		if err := json.Unmarshal([]byte(m.Content), &content); err != nil {
			return nil, err
		}
		frameMsg.Payload = &content

	case mime.Emoticon:
		content := pb.IM_Msg_Content_Emoticon{}
		if err := json.Unmarshal([]byte(m.Content), &content); err != nil {
			return nil, err
		}
		frameMsg.Payload = &content
	}

	return &frameMsg, nil
}
