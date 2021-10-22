package im

import (
	"context"
	"fmt"
	"time"

	"github.com/doublemo/baa/cores/crypto/id"
	log "github.com/doublemo/baa/cores/log/level"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/im/cache"
	"github.com/doublemo/baa/kits/im/dao"
	"github.com/doublemo/baa/kits/im/errcode"
	"github.com/doublemo/baa/kits/im/proto/pb"
	"github.com/doublemo/baa/kits/imf"
	imfproto "github.com/doublemo/baa/kits/imf/proto"
	imfpb "github.com/doublemo/baa/kits/imf/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
	natsgo "github.com/nats-io/nats.go"
)

// ChatConfig 聊天配置
type ChatConfig struct {

	// TopicsSecret 聊天主题加密key 16位
	TopicsSecret string `alias:"topicsSecret" default:"7581BDD8E8DA3839"`

	// IDSecret 用户ID 加密key 16位
	IDSecret string `alias:"idSecret" default:"7581BDD8E8DA3839"`
}

func send(req *corespb.Request, c ChatConfig) (*corespb.Response, error) {
	var frame pb.IM_Msg_Body
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	// 消息安全检查
	if !validationMsg(&frame) {
		return errcode.Bad(w, errcode.ErrInvalidCharacter), nil
	}

	// 消息送检
	//chatSubmissionInspection(&frame)
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	frame.MsgId, err = cache.GetSnowflakeID(ctx)
	if err != nil {
		cancel()
		log.Error(Logger()).Log("action", "getSnowflakeID", "error", err)
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	cancel()
	if frame.ToType == pb.IM_Msg_ToG {
		return sendtoG(req, &frame, c)
	}
	return sendtoC(req, &frame, c)
}

func sendtoC(req *corespb.Request, frame *pb.IM_Msg_Body, c ChatConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	// 检查是否是好友
	topicId, err := id.Decrypt(frame.To, []byte(c.TopicsSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInvalidTopicToken, err.Error()), nil
	}

	fromUserId, err := id.Decrypt(frame.From, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInvalidUserIdToken, err.Error()), nil
	}

	fmt.Println(topicId, fromUserId)
	topic := fmt.Sprintf("%d_%d", frame.ToType, topicId)
	// 开始存储数据

	bytes, _ := grpcproto.Marshal(frame)
	dao.SaveMsgtoRedis(topic, bytes)

	//
	res, err := dao.GetMsgFromRedis(topic)
	if err != nil {
		return nil, err
	}

	fmt.Println(res)
	m := pb.IM_Msg_Body{}
	for _, vv := range res {
		if err := grpcproto.Unmarshal([]byte(vv), &m); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(m, vv)
		}
	}
	return nil, nil
}

func sendtoG(req *corespb.Request, frame *pb.IM_Msg_Body, c ChatConfig) (*corespb.Response, error) {
	return nil, nil
}

func validationMsg(msg *pb.IM_Msg_Body) bool {
	return true
}

// chatSubmissionInspection 信息送检
func chatSubmissionInspection(frame *pb.IM_Msg_Body) error {
	nc := nats.Conn()
	if nc == nil {
		return nil
	}

	req := &corespb.Request{
		Command: imfproto.CheckCommand.Int32(),
		Header:  map[string]string{"service": ServiceName, "addr": sd.Endpoint().Addr(), "id": sd.Endpoint().ID()},
	}

	msg := imfpb.IMF_Request{
		MsgId:  frame.MsgId,
		Topic:  frame.To,
		ToType: int32(frame.ToType),
	}

	switch payload := frame.Payload.(type) {
	case *pb.IM_Msg_Body_Text:
		msg.Payload = &imfpb.IMF_Request_Text{
			Text: &imfpb.IMF_Content_Text{
				Content: payload.Text.Content,
			},
		}

	case *pb.IM_Msg_Body_Image:
		msg.Payload = &imfpb.IMF_Request_Image{
			Image: &imfpb.IMF_Content_Image{
				ID:      payload.Image.ID,
				Content: payload.Image.Content,
				Name:    payload.Image.Name,
				Url:     "",
			},
		}

	case *pb.IM_Msg_Body_Video:
		msg.Payload = &imfpb.IMF_Request_Video{
			Video: &imfpb.IMF_Content_Video{
				ID:   payload.Video.ID,
				Name: payload.Video.Name,
				Url:  "",
			},
		}

	case *pb.IM_Msg_Body_Voice:
		msg.Payload = &imfpb.IMF_Request_Voice{
			Voice: &imfpb.IMF_Content_Voice{
				Content: payload.Voice.Content,
			},
		}

	case *pb.IM_Msg_Body_File:
		msg.Payload = &imfpb.IMF_Request_File{
			File: &imfpb.IMF_Content_File{
				ID:    payload.File.ID,
				Types: payload.File.Types,
				Name:  payload.File.Name,
				Url:   "",
			},
		}

	default:
		return nil
	}

	req.Payload, _ = grpcproto.Marshal(&msg)
	bytes, _ := grpcproto.Marshal(req)
	err := nc.PublishMsg(&natsgo.Msg{
		Subject: imf.NatsGroupSubject(),
		Reply:   sd.Endpoint().ID(),
		Data:    bytes,
	})
	return err
}
