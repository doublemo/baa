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
	"github.com/doublemo/baa/kits/auth/platform"
	"github.com/doublemo/baa/kits/im/cache"
	"github.com/doublemo/baa/kits/im/dao"
	"github.com/doublemo/baa/kits/im/errcode"
	"github.com/doublemo/baa/kits/im/mime"
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

	// 开始存储数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := dao.WriteInboxC(ctx, ttid, ftid, msg); err != nil {
		return nil, &pb.IM_Msg_AckFailed{
			SeqID:      frame.SeqID,
			ErrCode:    errcode.ErrInternalServer.Code(),
			ErrMessage: err.Error(),
		}
	}

	// 消息送检
	msgInspectionReport(msg, ttid, ftid)

	// 推送信息
	pushMessage(*msg)
	return &pb.IM_Msg_AckReceived{
		Id:       msg.ID,
		SeqID:    msg.SeqId,
		NewSeqID: ftid,
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
	for _, id := range users {
		infos, ok := userStatus[id]
		if !ok {
			continue
		}

		if _, ok := addrs[id]; !ok {
			addrs[id] = make([]string, 0)
		}

		for k, v := range infos {
			switch k {
			case platform.PC, platform.Pad, platform.Phone:
				addrs[id] = append(addrs[id], v)
			}
		}
	}
	return addrs, nil
}

func pushMessage(msg ...dao.Messages) error {
	users := make([]uint64, len(msg))
	for i, m := range msg {
		users[i] = m.To
	}

	// agents, err := gatherUsersAgent(users...)
	// if err != nil {
	// 	return err
	// }

	// frame := pb.IM_
	return nil
}
