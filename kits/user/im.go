package user

import (
	"errors"
	"fmt"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

func sendChatMessages(im, userid string, messages ...*pb.IM_Msg_Content) ([]*pb.IM_Msg_AckReceived, error) {
	frame := &pb.IM_Send{
		Messages: &pb.IM_Msg_List{
			Values: messages,
		},
	}

	bytes, err := grpcproto.Marshal(frame)
	if err != nil {
		return nil, err
	}

	req := &corespb.Request{
		Header:  map[string]string{"Host": im, "UserID": userid},
		Command: command.IMSend.Int32(),
		Payload: bytes,
	}

	resp, err := muxRouter.Handler(kit.IM.Int32(), req)
	if err != nil {
		return nil, err
	}

	var respFrame = &pb.IM_Notify{}
	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		if err := grpcproto.Unmarshal(payload.Content, respFrame); err != nil {
			return nil, err
		}

	case *corespb.Response_Error:
		return nil, fmt.Errorf("<%d> %s", payload.Error.Code, payload.Error.Message)
	}

	if respFrame == nil {
		return nil, errors.New("send message failed")
	}

	switch payload := respFrame.Payload.(type) {
	case *pb.IM_Notify_Received:
		return payload.Received.Successed, nil
	}

	return nil, errors.New("send message failed")
}
