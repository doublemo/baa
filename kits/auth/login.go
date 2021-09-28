package auth

import (
	"fmt"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	snpb "github.com/doublemo/baa/kits/snid/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

func login(req *corespb.Request) (*corespb.Response, error) {
	frame := snpb.SNID_Request{
		N: 99,
	}

	b, _ := grpcproto.Marshal(&frame)
	resp, err := ir.Handler(&corespb.Request{Command: internalSnidRouter, Payload: b})
	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		fmt.Println(payload)
	}
	return resp, nil
}
