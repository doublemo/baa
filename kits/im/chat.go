package im

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/doublemo/baa/cores/crypto/id"
	log "github.com/doublemo/baa/cores/log/level"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/sd"
	agentproto "github.com/doublemo/baa/kits/agent/proto"
	agentpb "github.com/doublemo/baa/kits/agent/proto/pb"
	"github.com/doublemo/baa/kits/auth/platform"
	"github.com/doublemo/baa/kits/im/cache"
	"github.com/doublemo/baa/kits/im/dao"
	"github.com/doublemo/baa/kits/im/errcode"
	"github.com/doublemo/baa/kits/im/mime"
	"github.com/doublemo/baa/kits/im/proto"
	"github.com/doublemo/baa/kits/im/proto/pb"
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

	ret := &pb.IM_Msg_AckListReceived{
		Successed: make([]*pb.IM_Msg_AckReceived, 0),
		Failed:    make([]*pb.IM_Msg_AckFailed, 0),
	}

	for _, message := range frame.Messages.Values {
		received, failed := sendto(message, c)
		fmt.Println(received, failed)
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

func sendto(frame *pb.IM_Msg_Content, c ChatConfig) (*pb.IM_Msg_AckReceived, *pb.IM_Msg_AckFailed) {
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
	if frame.Group == pb.IM_Msg_ToG {
		return sendtoG(frame, c)
	}
	return sendtoC(frame, c)
}

func sendtoC(frame *pb.IM_Msg_Content, c ChatConfig) (*pb.IM_Msg_AckReceived, *pb.IM_Msg_AckFailed) {
	// 检查是否是好友
	msg, err := makeMessage(frame, []byte(c.IDSecret))
	if err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      frame.SeqID,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}

	// todo 检查是否彼此为好友
	backLoop := 0
	statusNoCache := false
back:
	userStatus, err := getCacheUserStatus(statusNoCache, msg.To, msg.From)
	if err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      frame.SeqID,
			ErrCode:    errcode.ErrInvalidUserStatus.Code(),
			ErrMessage: err.Error(),
		}
	}

	timeline, nomatch, err := getUID(userStatus, msg.To, msg.From)
	if err == ErrRematchServiceID && len(nomatch) > 0 && backLoop < 10 {
		backLoop++

		// todo auth server rematch
		fmt.Println("-----goto back", nomatch)
		time.Sleep(time.Millisecond)
		statusNoCache = true
		goto back
	}

	if err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      frame.SeqID,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}

	ttid, ok1 := timeline[msg.To]
	ftid, ok2 := timeline[msg.From]
	if !ok1 || !ok2 {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      frame.SeqID,
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
			SeqID:      frame.SeqID,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}

	// 消息送检
	msgInspectionReport(msg)

	// 推送信息
	if err := pushMessage([]byte(c.IDSecret), *msg); err != nil {
		log.Error(Logger()).Log("action", "pushMessage", "error", err)
	}

	return &pb.IM_Msg_AckReceived{
		Id:       msg.ID,
		SeqID:    msg.SeqId,
		NewSeqID: ttid,
	}, nil
}

func sendtoG(frame *pb.IM_Msg_Content, c ChatConfig) (*pb.IM_Msg_AckReceived, *pb.IM_Msg_AckFailed) {
	return nil, nil
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
	userStatus, err := getCacheUserStatus(false, users...)
	if err != nil {
		return nil, err
	}

	addrs := make(map[uint64][]string)
	addrsmap := make(map[uint64]map[string]bool)
	for _, id := range users {
		infos, ok := userStatus[id]
		if !ok {
			continue
		}

		if _, ok := addrs[id]; !ok {
			addrs[id] = make([]string, 0)
			addrsmap[id] = make(map[string]bool)
		}

		for k, v := range infos {
			switch k {
			case platform.PC, platform.Pad, platform.Phone, platform.Web:
				if !addrsmap[id][v] {
					addrs[id] = append(addrs[id], v)
					addrsmap[id][v] = true
				}
			}
		}
	}
	return addrs, nil
}

func pushMessage(idsecret []byte, msg ...dao.Messages) error {
	users := make([]uint64, len(msg))
	for i, m := range msg {
		users[i] = m.To
	}

	agents, err := gatherUsersAgent(users...)
	if err != nil {
		return err
	}

	nc := nats.Conn()
	for _, m := range msg {
		addrs, ok := agents[m.To]
		if !ok {
			continue
		}

		frame, err := makeMessageToPB(m, idsecret)
		if err != nil {
			return err
		}

		frame.SeqID = m.TSeqId
		w := &corespb.Response{
			Command: proto.PushCommand.Int32(),
			Header:  make(map[string]string),
		}

		bytes, _ := grpcproto.Marshal(&pb.IM_Notify{Payload: &pb.IM_Notify_List{List: &pb.IM_Msg_List{Values: []*pb.IM_Msg_Content{frame}}}})
		w.Payload = &corespb.Response_Content{Content: bytes}
		bytesW, _ := grpcproto.Marshal(w)
		req := &corespb.Request{
			Command: agentproto.BroadcastCommand.Int32(),
			Header:  map[string]string{"service": ServiceName, "addr": sd.Endpoint().Addr(), "id": sd.Endpoint().ID()},
		}

		req.Payload, _ = grpcproto.Marshal(&agentpb.Agent_Broadcast{
			Receiver: []string{frame.To},
			Payload:  bytesW,
			Command:  agentproto.IM.Int32(),
		})

		bytesMsg, _ := grpcproto.Marshal(req)
		for _, addr := range addrs {
			if err := nc.Publish(addr, bytesMsg); err != nil {
				return err
			}
		}
	}
	return nil
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
