package im

import (
	"errors"
	"strconv"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

func getSNID(num int32) ([]uint64, error) {
	if num > 1000 {
		return nil, errors.New("the number cannot be greater then 100")
	}

	frame := pb.SNID_Request{N: num}
	b, _ := grpcproto.Marshal(&frame)
	resp, err := muxRouter.Handler(kit.SNID.Int32(), &corespb.Request{Command: command.SNIDSnowflake.Int32(), Payload: b})
	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		resp := pb.SNID_Reply{}
		if err := grpcproto.Unmarshal(payload.Content, &resp); err != nil {
			return nil, err
		}

		if len(resp.Values) != int(num) {
			return nil, errors.New("errorSNIDLen")
		}

		return resp.Values, nil

	case *corespb.Response_Error:
		return nil, errors.New(payload.Error.Message)
	}
	return nil, errors.New("snid failed")
}

func namerUserTimeline(id uint64) string {
	return strconv.FormatUint(id, 10) + ":timelineid"
}

func getTimelines(nocache bool, values ...uint64) (map[uint64]uint64, error) {
	servers, err := getUserServers(nocache, values...)
	if err != nil {
		return nil, err
	}

	// 分组获取
	serversMap := make(map[string][]uint64)
	for id, addrs := range servers {
		addr, ok := addrs[kit.SNIDServiceName]
		if !ok {
			continue
		}

		enpoint, err := findServerEndpoint(addr)
		if err != nil {
			return nil, err
		}

		enpointAddr := enpoint.Addr()
		if _, ok := serversMap[enpointAddr]; !ok {
			serversMap[enpointAddr] = make([]uint64, 0)
		}
		serversMap[enpointAddr] = append(serversMap[enpointAddr], id)
	}

	retValues := make(map[uint64]uint64)
	for addr, users := range serversMap {
		frame := &pb.SNID_MoreRequest{
			Request: make([]*pb.SNID_Request, len(users)),
		}

		for i, value := range users {
			frame.Request[i] = &pb.SNID_Request{K: namerUserTimeline(value), N: 1}
		}

		b, _ := grpcproto.Marshal(frame)
		resp, err := muxRouter.Handler(kit.SNID.Int32(), &corespb.Request{Command: command.SNIDMoreAutoincrement.Int32(), Payload: b, Header: map[string]string{"Host": addr}})
		if err != nil {
			return nil, err
		}

		switch payload := resp.Payload.(type) {
		case *corespb.Response_Content:
			resp := pb.SNID_MoreReply{}
			if err := grpcproto.Unmarshal(payload.Content, &resp); err != nil {
				return nil, err
			}

			for _, id := range values {
				if m, ok := resp.Values[namerUserTimeline(id)]; ok && len(m.Values) > 0 {
					retValues[id] = m.Values[0]
				}
			}

		case *corespb.Response_Error:
			return nil, errors.New(payload.Error.Message)
		}
	}

	return retValues, nil
}
