package snid

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/uid"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/kits/snid/dao"
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
	} else if num > 1000 {
		return errcode.Bad(w, errcode.ErrMaxIDNumber), nil
	}

	values, err := dao.AutoincrementID(frame.K, int64(num))
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

	values, err := dao.AutoincrementID(frame.K, int64(num))
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
