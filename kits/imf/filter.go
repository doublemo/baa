package imf

import (
	"errors"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/im/mime"
	"github.com/doublemo/baa/kits/imf/errcode"
	"github.com/doublemo/baa/kits/imf/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

// FilterConfig 过滤
type FilterConfig struct {

	// TextReplaceWord 如果遇到脏话将替换为指定字符
	TextReplaceWord string `alias:"textReplaceWord"`

	// DictionaryPath 字典路径
	DictionaryPath string `alias:"dictionaryPath" default:"dictionary/dictionary.txt"`

	// DirtyPath 脏词字典路径
	DirtyPath string `alias:"dirtyPath"  default:"dictionary/dirty.txt"`
}

func checkFromNats(req *corespb.Request, c FilterConfig) (*corespb.Response, error) {
	var frame pb.IMF_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	ret := &pb.IMF_Reply{
		Values: make([]*pb.IMF_Response, 0),
	}

	for _, value := range frame.Values {
		resp, err := check(value, c)
		if err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		if resp.Ok {
			ret.Values = append(ret.Values, resp)
		}
	}

	if len(ret.Values) < 1 {
		return nil, nil
	}

	b, _ := grpcproto.Marshal(ret)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func checkFromRPC(req *corespb.Request, c FilterConfig) (*corespb.Response, error) {
	var frame pb.IMF_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	ret := &pb.IMF_Reply{
		Values: make([]*pb.IMF_Response, 0),
	}

	for _, value := range frame.Values {
		resp, err := check(value, c)
		if err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		if resp.Ok {
			ret.Values = append(ret.Values, resp)
		}
	}

	b, _ := grpcproto.Marshal(ret)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func check(frame *pb.IMF_Content, c FilterConfig) (*pb.IMF_Response, error) {
	resp := pb.IMF_Response{}
	switch frame.ContentType {
	case mime.Text:
		text, ok := filterText(frame.Content, c)
		frame.Content = text
		resp.Ok = ok
		resp.Content = frame
	default:
		return nil, errors.New("notsupported")
	}
	return &resp, nil
}
