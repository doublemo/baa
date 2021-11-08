package im

import (
	"errors"
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
		return nil, errors.New(payload.Error.Message)
	}
	return nil, errors.New("usrt failed")
}

func getCacheUsersStatus(noCache bool, values ...uint64) ([]*pb.SM_User_Status, error) {
	data := make([]*pb.SM_User_Status, 0)
	noCacheData := make([]uint64, 0)

	if !noCache {
		for _, value := range values {
			if m, ok := cache.Get(namerCacheUserStatus(value)); ok && m != nil {
				if m0, ok := m.(*pb.SM_User_Status); ok {
					data = append(data, m0)
					continue
				}
			}
			noCacheData = append(noCacheData, value)
		}
	} else {
		noCacheData = values
	}

	if len(noCacheData) < 1 {
		return data, nil
	}

	retValues, err := getUsersStatus(noCache, noCacheData...)
	if err != nil {
		return nil, err
	}

	dataLen := len(data)
	retValuesLen := len(retValues)
	newData := make([]*pb.SM_User_Status, dataLen+retValuesLen)
	if dataLen > 0 {
		copy(newData[0:], data[0:])
		copy(newData[:dataLen], retValues[0:])
	} else {
		newData = retValues
	}

	for _, value := range retValues {
		cache.Set(namerCacheUserStatus(value.UserId), value, 0)
	}

	return newData, nil
}

func namerCacheUserStatus(id uint64) string {
	return "userstatus_" + strconv.FormatUint(id, 10)
}

func findServersID(id uint64, server string, data []*pb.SM_User_Status) ([]string, bool) {
	servers := make([]string, 0)
	for _, value := range data {
		if value.UserId != id {
			continue
		}

		if len(value.Values) < 1 {
			return nil, false
		}

		for _, v := range value.Values {
			switch server {
			case kit.SNIDServiceName:
				servers = append(servers, v.IDServer)
			case kit.IMServiceName:
				servers = append(servers, v.IMServer)
			case kit.AgentServiceName:
				servers = append(servers, v.AgentServer)
			}
		}

		break
	}

	if len(servers) < 1 {
		return nil, false
	}
	return servers, true
}

func findServersAddr(id uint64, server string, data []*pb.SM_User_Status) ([]string, bool) {
	servers := make([]string, 0)
	for _, value := range data {
		if value.UserId != id {
			continue
		}

		if len(value.Values) < 1 {
			return nil, false
		}

		for _, v := range value.Values {
			switch server {
			case kit.SNIDServiceName:
				if ed, err := findServerEndpoint(v.IDServer); err == nil {
					servers = append(servers, ed.Addr())
				}

			case kit.IMServiceName:
				if ed, err := findServerEndpoint(v.IMServer); err == nil {
					servers = append(servers, ed.Addr())
				}

			case kit.AgentServiceName:
				if ed, err := findServerEndpoint(v.AgentServer); err == nil {
					servers = append(servers, ed.Addr())
				}
			}
		}

		break
	}

	if len(servers) < 1 {
		return nil, false
	}
	return servers, true
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
