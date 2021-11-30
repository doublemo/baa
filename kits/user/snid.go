package user

import (
	"errors"
	"fmt"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

func getSNID(num int32) ([]uint64, error) {
	if num > 10 {
		return nil, errors.New("the number cannot be greater then 10")
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
		return nil, fmt.Errorf("<%d> %s", payload.Error.Code, payload.Error.Message)
	}

	return nil, errors.New("snid failed")
}
