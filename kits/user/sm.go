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
