package imf

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/imf/proto/pb"
	"github.com/doublemo/baa/kits/imf/segmenter"
	grpcproto "github.com/golang/protobuf/proto"
)

func filterText(req *corespb.Request, frame *pb.IMF_Request, payload *pb.IMF_Request_Text, c FilterConfig) (*corespb.Response, bool, error) {
	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	text, ok := segmenter.ReplaceDirtyWords(payload.Text.Content, c.TextReplaceWord)

	ret := pb.IMF_Reply{
		OK:      ok,
		Payload: &pb.IMF_Reply_Text{Text: &pb.IMF_Content_Text{Content: text}},
		MsgId:   frame.MsgId,
		Topic:   frame.Topic,
		ToType:  frame.ToType,
		SeqId:   frame.SeqId,
	}

	bytes, _ := grpcproto.Marshal(&ret)
	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, ok, nil
}
