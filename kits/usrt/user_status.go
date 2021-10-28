package usrt

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/usrt/dao"
	grpcproto "github.com/golang/protobuf/proto"
)

func updateUserStatus(r *corespb.Request) (*corespb.Response, error) {
	var frame pb.USRT_Status_Update
	{
		if err := grpcproto.Unmarshal(r.Payload, &frame); err != nil {
			return nil, err
		}
	}

	changeDataMap := make(map[uint64]bool)
	changeData := make([]uint64, 0)
	for _, v := range frame.Values {
		changeDataMap[v.ID] = true
	}

	w := &corespb.Response{Command: r.Command, Header: r.Header}
	if len(frame.Values) < 1 {
		reply := &pb.USRT_Status_Reply{Values: make([]*pb.USRT_User, 0)}
		b, _ := grpcproto.Marshal(reply)
		w.Payload = &corespb.Response_Content{Content: b}
		return w, nil
	}

	noCompleted, err := dao.UpdateStatusByUser(frame.Values...)
	if err != nil {
		return nil, err
	}

	if len(noCompleted) < 1 {
		for k := range changeDataMap {
			changeData = append(changeData, k)
		}

		if err := pushUserStatusChangeMessage(command.USRTUpdateUserStatus.Int32(), changeData...); err != nil {
			return nil, err
		}

		reply := &pb.USRT_Status_Reply{Values: make([]*pb.USRT_User, 0)}
		b, _ := grpcproto.Marshal(reply)
		w.Payload = &corespb.Response_Content{Content: b}
		return w, nil
	}

	noCompletedMap := make(map[uint64]bool)
	for _, id := range noCompleted {
		noCompletedMap[id] = true
	}

	for k := range changeDataMap {
		if noCompletedMap[k] {
			continue
		}
		changeData = append(changeData, k)
	}

	if err := pushUserStatusChangeMessage(command.USRTUpdateUserStatus.Int32(), changeData...); err != nil {
		return nil, err
	}

	reply := &pb.USRT_Status_Reply{Values: make([]*pb.USRT_User, len(noCompletedMap))}
	for i, v := range frame.Values {
		if !noCompletedMap[v.ID] {
			continue
		}

		reply.Values[i] = v
	}

	b, _ := grpcproto.Marshal(reply)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func deleteUserStatus(r *corespb.Request) (*corespb.Response, error) {
	var frame pb.USRT_Status_Update
	{
		if err := grpcproto.Unmarshal(r.Payload, &frame); err != nil {
			return nil, err
		}
	}

	reply := &pb.USRT_Status_Reply{Values: make([]*pb.USRT_User, 0)}
	w := &corespb.Response{Command: r.Command, Header: r.Header}
	b, _ := grpcproto.Marshal(reply)
	w.Payload = &corespb.Response_Content{Content: b}
	if len(frame.Values) < 1 {
		return w, nil
	}

	if err := dao.RemoveStatusByUser(frame.Values...); err != nil {
		return nil, err
	}

	changeDataMap := make(map[uint64]bool)
	changeData := make([]uint64, 0)
	for _, v := range frame.Values {
		changeDataMap[v.ID] = true
	}

	for k := range changeDataMap {
		changeData = append(changeData, k)
	}

	if err := pushUserStatusChangeMessage(command.USRTDeleteUserStatus.Int32(), changeData...); err != nil {
		return nil, err
	}

	return w, nil
}

func getUserStatus(r *corespb.Request) (*corespb.Response, error) {
	var frame pb.USRT_Status_Request
	{
		if err := grpcproto.Unmarshal(r.Payload, &frame); err != nil {
			return nil, err
		}
	}

	noCache := false
	if r.Header != nil {
		if _, ok := r.Header["no-cache"]; ok {
			noCache = true
		}
	}

	w := &corespb.Response{Command: r.Command, Header: r.Header}
	if len(frame.Values) < 1 {
		reply := &pb.USRT_Status_Reply{Values: make([]*pb.USRT_User, 0)}
		b, _ := grpcproto.Marshal(reply)
		w.Payload = &corespb.Response_Content{Content: b}
		return w, nil
	}

	var (
		data []*pb.USRT_User
		err  error
	)

	if noCache {
		data, err = dao.GetStatusByUser(frame.Values...)
	} else {
		data, err = dao.GetStatueCacheByUser(frame.Values...)
	}

	if err != nil {
		return nil, err
	}

	reply := &pb.USRT_Status_Reply{Values: data}
	b, _ := grpcproto.Marshal(reply)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func pushUserStatusChangeMessage(command int32, id ...uint64) error {
	frame := pb.USRT_Status_Request{Values: id}
	req := corespb.Request{
		Command: command,
		Header:  map[string]string{"service": ServiceName, "addr": sd.Endpoint().Addr(), "id": sd.Endpoint().ID()},
	}

	req.Payload, _ = grpcproto.Marshal(&frame)
	bytes, _ := grpcproto.Marshal(&req)
	nc := nats.Conn()
	if err := nc.Publish(NatsUserStatusWatchSubject, bytes); err != nil {
		return err
	}

	return nil
}
