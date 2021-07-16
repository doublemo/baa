package sfu

import (
	"encoding/json"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/sfu/adapter/router"
	"github.com/doublemo/baa/kits/sfu/proto"
	"github.com/doublemo/baa/kits/sfu/proto/pb"
	"github.com/doublemo/baa/kits/sfu/session"
	grpcproto "github.com/golang/protobuf/proto"
	ionsfu "github.com/pion/ion-sfu/pkg/sfu"
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

	resp := &corespb.Response{Command: req.Command}
	peerId, ok := req.Header["PeerId"]
	if !ok {
		return nil, nil
	}

	p, ok := session.GetPeer(peerId)
	if !ok {
		return nil, nil
	}

	peer, ok := p.(*ionsfu.PeerLocal)
	if !ok {
		return nil, nil
	}

	switch payload := frame.Payload.(type) {
	case *pb.SFU_Signal_Request_Description:
		var sdp webrtc.SessionDescription
		{
			if err := json.Unmarshal(payload.Description, &sdp); err != nil {
				return nil, err
			}
		}

		if sdp.Type == webrtc.SDPTypeOffer {
			answer, err := peer.Answer(sdp)
			if err != nil {
				return nil, err
			}

			bytes, _ := json.Marshal(answer)
			data := pb.SFU_Signal_Reply{
				SessionId: "test",
				PeerId:    peerId,
				Payload: &pb.SFU_Signal_Reply_Description{
					Description: bytes,
				},
			}

			bytes2, _ := grpcproto.Marshal(&data)
			resp.Payload = &corespb.Response_Content{
				Content: bytes2,
			}

			return resp, nil
		} else if sdp.Type == webrtc.SDPTypeAnswer {
			var sdp webrtc.SessionDescription
			{
				if err := json.Unmarshal(payload.Description, &sdp); err != nil {
					return nil, err
				}
			}
			peer.SetRemoteDescription(sdp)
		}
	case *pb.SFU_Signal_Request_Trickle:
		var candidate webrtc.ICECandidateInit
		json.Unmarshal([]byte(payload.Trickle.Candidate), &candidate)
		peer.Trickle(candidate, int(payload.Trickle.Target))

	}
	return nil, nil
}
