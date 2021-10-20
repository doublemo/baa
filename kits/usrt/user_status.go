package usrt

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/usrt/dao"
	"github.com/doublemo/baa/kits/usrt/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

func updateUserStatus(r *corespb.Request) (*corespb.Response, error) {
	var frame pb.USRT_Status_Update
	{
		if err := grpcproto.Unmarshal(r.Payload, &frame); err != nil {
			return nil, err
		}
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
		reply := &pb.USRT_Status_Reply{Values: make([]*pb.USRT_User, 0)}
		b, _ := grpcproto.Marshal(reply)
		w.Payload = &corespb.Response_Content{Content: b}
		return w, nil
	}

	noCompletedMap := make(map[uint64]bool)
	for _, id := range noCompleted {
		noCompletedMap[id] = true
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