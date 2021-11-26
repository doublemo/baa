package robot

import (
	"errors"
	"fmt"
	"time"

	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/robot/session"
	grpcproto "github.com/golang/protobuf/proto"
)

func SendChatMessage(peer session.Peer, messages ...*pb.IM_Msg_Content) error {

	frame := &pb.IM_Send{
		Messages: &pb.IM_Msg_List{Values: messages},
	}

	bytes, err := grpcproto.Marshal(frame)
	if err != nil {
		return err
	}

	req := &coresproto.RequestBytes{
		Ver:     1,
		Cmd:     kit.IM,
		SubCmd:  command.IMSend,
		Content: bytes,
		SeqID:   1,
	}

	r, err := req.Marshal()
	if err != nil {
		return err
	}

	return peer.Send(session.PeerMessagePayload{Data: r, Channel: session.PeerMessageChannelWebrtc})
}

func sendTestChat(peer session.Peer) error {
	userid, ok := peer.Params("UserID")
	if !ok {
		return errors.New("invalid UserID")
	}

	uid := userid.(string)
	message := &pb.IM_Msg_Content{
		To:      "NJlPF3UZcr0",
		From:    uid,
		Group:   pb.IM_Msg_ToC,
		Topic:   15771692631559582519,
		Payload: &pb.IM_Msg_Content_Text{Text: &pb.IM_Msg_ContentType_Text{Content: "垃圾企业垃圾股新一轮下跌又开始了4.8企稳"}},
	}

	return SendChatMessage(peer, message)
}

func handleIMNotify(peer session.Peer, w coresproto.Response, c RobotConfig) error {
	if w.StatusCode() != 0 {
		return fmt.Errorf("<%d> %s", w.StatusCode(), w.Body())
	}

	var frame pb.IM_Notify
	{
		if err := grpcproto.Unmarshal(w.Body(), &frame); err != nil {
			return err
		}
	}

	switch payload := frame.Payload.(type) {
	case *pb.IM_Notify_Confirmed:
		fmt.Println("IM_Notify_Confirmed", payload.Confirmed.Values)
	case *pb.IM_Notify_Readed:
		fmt.Println("IM_Notify_Readed", payload.Readed.Values)
	case *pb.IM_Notify_Received:
		fmt.Println("IM_Notify_Received", payload.Received.Failed, payload.Received.Successed)
	case *pb.IM_Notify_List:
		for _, message := range payload.List.Values {
			switch v := message.Payload.(type) {
			case *pb.IM_Msg_Content_Text:
				fmt.Printf("<%s>id: %d, seq:%d from: %s message, %s\n", time.Unix(message.SendAt, 0).String(), message.Id, message.SeqID, message.From, v.Text.Content)
			}
		}
	}

	return nil
}
