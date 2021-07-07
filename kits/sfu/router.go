package sfu

import (
	"encoding/json"
	"fmt"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/sfu/adapter/router"
	"github.com/doublemo/baa/kits/sfu/proto"
	"github.com/doublemo/baa/kits/sfu/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
	"github.com/pion/webrtc/v3"
)

// InitRouter init
func InitRouter() {
	router.On(proto.NegotiateCommand, negotiate)
}

func negotiate(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SFU_Signal_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	switch payload := frame.Payload.(type) {
	case *pb.SFU_Signal_Request_Description:
		var sdp webrtc.SessionDescription
		{
			if err := json.Unmarshal(payload.Description, &sdp); err != nil {
				return nil, err
			}
		}

		fmt.Println(sdp)
	case *pb.SFU_Signal_Request_Trickle:

	}
	return nil, nil
}
