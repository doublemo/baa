package snid

import (
	"context"
	"time"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/uid"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/kits/snid/cache"
	"github.com/doublemo/baa/kits/snid/errcode"
	"github.com/doublemo/baa/kits/snid/proto/pb"
	"github.com/golang/protobuf/jsonpb"
	grpcproto "github.com/golang/protobuf/proto"
)

type snHandler struct {
	uidGenerator *uid.Snowflake
}

func (sn *snHandler) Serve(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SNID_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{Command: req.Command}
	if req.Header != nil {
		w.Header = req.Header
	}

	num := frame.N
	if num < 1 {
		num = 1
	} else if num > 1000 {
		return errcode.Bad(w, errcode.ErrMaxIDNumber), nil
	}

	resp := &pb.SNID_Reply{
		Values: make([]uint64, num),
	}

	for i := 0; i < int(num); i++ {
		resp.Values[i] = sn.uidGenerator.NextId()
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func (sn *snHandler) ServeHTTP(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SNID_Request
	{
		if err := jsonpb.UnmarshalString(string(req.Payload), &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{Command: req.Command}
	if req.Header != nil {
		w.Header = req.Header
	}

	num := frame.N
	if num < 1 {
		num = 1
	} else if num > 100 {
		return errcode.Bad(w, errcode.ErrMaxIDNumber), nil
	}

	resp := &pb.SNID_Reply{
		Values: make([]uint64, num),
	}

	for i := 0; i < int(num); i++ {
		resp.Values[i] = sn.uidGenerator.NextId()
	}

	jsonpbM := &jsonpb.Marshaler{}
	b, _ := jsonpbM.MarshalToString(resp)
	w.Payload = &corespb.Response_Content{Content: []byte(b)}
	return w, nil
}

func newSnHandler(c uid.SnowflakeConfig) *snHandler {
	return &snHandler{
		uidGenerator: uid.NewSnowflakeGenerator(c),
	}
}

func autoincrementID(req *corespb.Request) (*corespb.Response, error) {
	if router.IsHTTP(req) {
		return autoincrementIDToHTTP(req)
	}

	var frame pb.SNID_Request
	{

		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{Command: req.Command}
	if req.Header != nil {
		w.Header = req.Header
	}

	if frame.K == "" {
		return errcode.Bad(w, errcode.ErrKeyIsEmpty), nil
	}

	num := frame.N
	if num < 1 {
		num = 1
	} else if num > 100 {
		return errcode.Bad(w, errcode.ErrMaxIDNumber), nil
	}

	//values, err := dao.AutoincrementID(frame.K, int64(num))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	values, err := cache.GetUID(ctx, frame.K, int(frame.N))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.SNID_Reply{
		Values: values,
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func autoincrementIDToHTTP(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SNID_Request
	{

		if err := jsonpb.UnmarshalString(string(req.Payload), &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{Command: req.Command}
	if req.Header != nil {
		w.Header = req.Header
	}

	if frame.K == "" {
		return errcode.Bad(w, errcode.ErrKeyIsEmpty), nil
	}

	num := frame.N
	if num < 1 {
		num = 1
	} else if num > 100 {
		return errcode.Bad(w, errcode.ErrMaxIDNumber), nil
	}

	//values, err := dao.AutoincrementID(frame.K, int64(num))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	values, err := cache.GetUID(ctx, frame.K, int(frame.N))

	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.SNID_Reply{
		Values: values,
	}

	jsonpbM := &jsonpb.Marshaler{}
	b, _ := jsonpbM.MarshalToString(resp)
	w.Payload = &corespb.Response_Content{Content: []byte(b)}
	return w, nil
}

func clearAutoincrementID(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SNID_Clear_Request
	{

		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	for _, k := range frame.K {
		cache.RemoveUID(k)
	}

	w := &corespb.Response{Command: req.Command}
	if req.Header != nil {
		w.Header = req.Header
	}

	resp := &pb.SNID_Clear_Reply{
		OK: true,
	}
	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func moreAutoincrementID(req *corespb.Request) (*corespb.Response, error) {
	if router.IsHTTP(req) {
		return moreAutoincrementIDToHTTP(req)
	}

	var frame pb.SNID_MoreRequest
	{

		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{Command: req.Command}
	if req.Header != nil {
		w.Header = req.Header
	}

	if len(frame.Request) < 1 {
		return errcode.Bad(w, errcode.ErrKeyIsEmpty), nil
	}

	values := make(map[string]*pb.SNID_Reply)
	for _, req := range frame.Request {
		if req.N < 1 {
			req.N = 1
		} else if req.N > 100 {
			return errcode.Bad(w, errcode.ErrMaxIDNumber), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		vs, err := cache.GetUID(ctx, req.K, int(req.N))
		if err != nil {
			cancel()
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		cancel()
		values[req.K] = &pb.SNID_Reply{
			Values: vs,
		}
	}

	resp := &pb.SNID_MoreReply{
		Values: values,
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func moreAutoincrementIDToHTTP(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SNID_MoreRequest
	{

		if err := jsonpb.UnmarshalString(string(req.Payload), &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{Command: req.Command}
	if req.Header != nil {
		w.Header = req.Header
	}

	if len(frame.Request) < 1 {
		return errcode.Bad(w, errcode.ErrKeyIsEmpty), nil
	}

	values := make(map[string]*pb.SNID_Reply)
	for _, req := range frame.Request {
		if req.N < 1 {
			req.N = 1
		} else if req.N > 100 {
			return errcode.Bad(w, errcode.ErrMaxIDNumber), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		vs, err := cache.GetUID(ctx, req.K, int(req.N))
		if err != nil {
			cancel()
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		cancel()
		values[req.K] = &pb.SNID_Reply{
			Values: vs,
		}
	}

	resp := &pb.SNID_MoreReply{
		Values: values,
	}
	jsonpbM := &jsonpb.Marshaler{}
	b, _ := jsonpbM.MarshalToString(resp)
	w.Payload = &corespb.Response_Content{Content: []byte(b)}
	return w, nil
}
