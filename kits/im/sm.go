package im

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/im/cache"
	"github.com/doublemo/baa/kits/sm"
	grpcproto "github.com/golang/protobuf/proto"
)

func publishUserState(frame *pb.SM_Event) error {
	nc := nats.Conn()
	if nc == nil {
		return errors.New("nats is nil")
	}

	req := coresproto.RequestBytes{
		Cmd:    kit.SM,
		SubCmd: command.SMEvent,
	}

	req.Content, _ = grpcproto.Marshal(frame)
	data, _ := req.Marshal()
	if err := nc.Publish(sm.ChannelStateEventReceiver, data); err != nil {
		return err
	}

	return nc.FlushTimeout(time.Second)
}

func getUserServers(nocache bool, values ...uint64) (map[uint64]map[string]string, error) {
	if len(values) > 100 {
		return nil, errors.New("the value length cannot be greater then 100")
	}

	newValues := make(map[uint64]map[string]string)
	needRequestValues := make([]uint64, 0)
	if !nocache {
		for _, id := range values {
			if m, ok := cache.Get(namerCacheUserServers(id)); ok && m != nil {
				if m0, ok := m.(map[string]string); ok && m0 != nil {
					newValues[id] = m0
					continue
				}
			}

			needRequestValues = append(needRequestValues, id)
		}
	} else {
		needRequestValues = values
	}

	if len(needRequestValues) < 1 {
		return newValues, nil
	}

	frame := &pb.SM_User_Servers_Request{
		Values: needRequestValues,
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
		for _, v := range reply.Values {
			newValues[v.UserId] = v.Servers
			cache.Set(namerCacheUserServers(v.UserId), v.Servers, 0)
		}
		return newValues, nil

	case *corespb.Response_Error:
		return nil, fmt.Errorf("<%d> %s", payload.Error.Code, payload.Error.Message)
	}
	return nil, errors.New("usrt failed")
}

func namerCacheUserServers(id uint64) string {
	return "userservers_" + strconv.FormatUint(id, 10)
}

func findServerEndpoint(id string) (coressd.Endpoint, error) {
	eds, err := sd.Endpoints()
	if err != nil {
		return nil, err
	}

	for _, ed := range eds {
		if ed.ID() == id {
			return ed, nil
		}
	}

	return nil, errors.New("Node does not exist ")
}
