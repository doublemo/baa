package im

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/im/proto/pb"
	"github.com/doublemo/baa/kits/imf"
	imfproto "github.com/doublemo/baa/kits/imf/proto"
	imfpb "github.com/doublemo/baa/kits/imf/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
	natsgo "github.com/nats-io/nats.go"
)

func msgInspectionReport(frame *pb.IM_Msg_Body) error {
	nc := nats.Conn()
	if nc == nil {
		return nil
	}

	req := &corespb.Request{
		Command: imfproto.CheckCommand.Int32(),
		Header:  map[string]string{"service": ServiceName, "addr": sd.Endpoint().Addr(), "id": sd.Endpoint().ID()},
	}

	msg := imfpb.IMF_Request{
		MsgId:         frame.MsgId,
		Topic:         "",
		ToType:        int32(frame.ToType),
		RequiredReply: false,
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

func handleMsgInspectionReport(req *corespb.Request) (*corespb.Response, error) {
	var frame imfpb.IMF_Reply
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	// todo 处理非法的消息
	return nil, nil
}
