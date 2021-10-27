package im

import (
	"fmt"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/im/dao"
	"github.com/doublemo/baa/kits/imf"
	imfproto "github.com/doublemo/baa/kits/imf/proto"
	imfpb "github.com/doublemo/baa/kits/imf/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
	natsgo "github.com/nats-io/nats.go"
)

func msgInspectionReport(msg *dao.Messages, seqs ...uint64) error {
	nc := nats.Conn()
	if nc == nil {
		return nil
	}

	req := &corespb.Request{
		Command: imfproto.CheckCommand.Int32(),
		Header:  map[string]string{"service": ServiceName, "addr": sd.Endpoint().Addr(), "id": sd.Endpoint().ID()},
	}

	data := imfpb.IMF_Request{
		Values: make([]*imfpb.IMF_Content, 0),
	}

	data.Values = append(data.Values, &imfpb.IMF_Content{
		MsgId:       msg.ID,
		SeqId:       msg.SeqId,
		Topic:       msg.Topic,
		Group:       msg.Group,
		Content:     msg.Content,
		To:          msg.To,
		From:        msg.From,
		ContentType: msg.ContentType,
		TSeqId:      msg.TSeqId,
		FSeqId:      msg.FSeqId,
	})

	req.Payload, _ = grpcproto.Marshal(&data)
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
	for _, value := range frame.Values {
		if !value.Ok {
			continue
		}

		m := dao.Messages{
			ID:          value.Content.MsgId,
			SeqId:       value.Content.SeqId,
			To:          value.Content.To,
			From:        value.Content.From,
			Content:     value.Content.Content,
			ContentType: value.Content.ContentType,
			Topic:       value.Content.Topic,
			TSeqId:      value.Content.TSeqId,
			FSeqId:      value.Content.FSeqId,
			Group:       value.Content.Group,
		}

		fmt.Println("chek msg :=>", m)
	}
	return nil, nil
}
