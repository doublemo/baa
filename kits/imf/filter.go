package imf

import (
	"errors"

	corespb "github.com/doublemo/baa/cores/proto/pb"
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

	resp, ok, err := check(req, &frame, c)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	if frame.RequiredReply || ok {
		return resp, nil
	}

	return nil, nil
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

	resp, _, err := check(req, &frame, c)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	return resp, nil
}

func check(req *corespb.Request, frame *pb.IMF_Request, c FilterConfig) (*corespb.Response, bool, error) {
	switch payload := frame.Payload.(type) {
	case *pb.IMF_Request_Text:
		return filterText(req, frame, payload, c)

	case *pb.IMF_Request_Image:
	case *pb.IMF_Request_Video:
	case *pb.IMF_Request_Voice:
	case *pb.IMF_Request_File:
	}

	return nil, false, errors.New("notsupported")
}
