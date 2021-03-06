package im

import (
	"context"
	"fmt"
	"time"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/im/dao"
	"github.com/doublemo/baa/kits/imf"
	grpcproto "github.com/golang/protobuf/proto"
	natsgo "github.com/nats-io/nats.go"
)

func msgInspectionReport(msg *dao.Messages, seqs ...uint64) error {
	nc := nats.Conn()
	if nc == nil {
		return nil
	}

	req := &coresproto.RequestBytes{
		Ver:    1, // 版本号不为0，则要求对方回复消息
		Cmd:    kit.IMF,
		SubCmd: command.IMFCheck,
		SeqID:  1,
	}

	data := pb.IMF_Request{
		Values: make([]*pb.IMF_Content, 0),
	}

	data.Values = append(data.Values, &pb.IMF_Content{
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

	req.Content, _ = grpcproto.Marshal(&data)
	bytes, err := req.Marshal()
	if err != nil {
		return err
	}
	err = nc.PublishMsg(&natsgo.Msg{
		Subject: imf.NatsGroupSubject(),
		Reply:   sd.Endpoint().ID(),
		Data:    bytes,
	})
	return err
}

func handleMsgInspectionReport(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.IMF_Reply
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

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		dao.ChangeInboxMessageStatus(ctx, m.ID, 1)
		cancel()
		fmt.Println("chek msg :=>", m)
	}
	return nil, nil
}
