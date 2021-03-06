package auth

import (
	"errors"
	"fmt"
	"time"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

const (
	// channelStateEventBroadcast 状态变更时通知通道
	channelStateEventBroadcast = "sm.broadcast"

	// channelStateEventReceiver 接收状态事件通道
	channelStateEventReceiver = "sm.receiver"
)

func publishUserState(frame *pb.SM_Event) error {
	nc := nats.Conn()
	if nc == nil {
		return errors.New("nats is nil")
	}

	req := coresproto.RequestBytes{
		Cmd:    kit.SM,
		SubCmd: command.SMEvent,
		SeqID:  1,
	}

	req.Content, _ = grpcproto.Marshal(frame)
	data, _ := req.Marshal()
	if err := nc.Publish(channelStateEventReceiver, data); err != nil {
		return err
	}

	return nc.FlushTimeout(time.Second)
}

func getUsersStatus(noCache bool, values ...uint64) ([]*pb.SM_User_Status, error) {
	if len(values) > 100 {
		return nil, errors.New("the value length cannot be greater then 100")
	}

	header := make(map[string]string)
	if noCache {
		header["no-cache"] = "true"
	}

	frame := &pb.SM_User_Request{
		Values: values,
	}

	data, _ := grpcproto.Marshal(frame)
	resp, err := muxRouter.Handler(kit.SM.Int32(), &corespb.Request{Command: command.SMUserStatus.Int32(), Payload: data, Header: header})
	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		resp := pb.SM_User_Reply{}
		if err := grpcproto.Unmarshal(payload.Content, &resp); err != nil {
			return nil, err
		}
		return resp.Values, nil

	case *corespb.Response_Error:
		return nil, fmt.Errorf("<%d> %s", payload.Error.Code, payload.Error.Message)
	}
	return nil, errors.New("status failed")
}

func getUserServers(values ...uint64) (map[uint64]map[string]string, error) {
	if len(values) > 100 {
		return nil, errors.New("the value length cannot be greater then 100")
	}

	frame := &pb.SM_User_Servers_Request{
		Values: values,
	}

	data, _ := grpcproto.Marshal(frame)
	resp, err := muxRouter.Handler(kit.SM.Int32(), &corespb.Request{Command: command.SMUserServers.Int32(), Payload: data, Header: make(map[string]string)})
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		reply := pb.SM_User_Servers_Reply{}
		if err := grpcproto.Unmarshal(payload.Content, &reply); err != nil {
			return nil, err
		}
		newValues := make(map[uint64]map[string]string)
		for _, v := range reply.Values {
			newValues[v.UserId] = v.Servers
		}
		return newValues, nil

	case *corespb.Response_Error:
		return nil, fmt.Errorf("<%d> %s", payload.Error.Code, payload.Error.Message)
	}
	return nil, errors.New("sm failed")
}

func assignServers(userid uint64, values ...*pb.SM_User_Servers_Assign) (*pb.SM_User_Servers_AssignServerReply, error) {

	frame := &pb.SM_User_Servers_AssignServerRequest{
		UserId: userid,
		Values: values,
	}

	data, _ := grpcproto.Marshal(frame)
	resp, err := muxRouter.Handler(kit.SM.Int32(), &corespb.Request{Command: command.SMAssginServers.Int32(), Payload: data, Header: make(map[string]string)})
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		reply := pb.SM_User_Servers_AssignServerReply{}
		if err := grpcproto.Unmarshal(payload.Content, &reply); err != nil {
			return nil, err
		}
		return &reply, nil

	case *corespb.Response_Error:
		return nil, fmt.Errorf("<%d> %s", payload.Error.Code, payload.Error.Message)
	}
	return nil, errors.New("sm failed")
}
