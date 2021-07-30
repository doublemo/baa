package sfu

import (
	"encoding/json"
	"fmt"

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

func join(client corespb.Service_BidirectionalStreamingServer, peerId string, req *corespb.Request) (*corespb.Response, error) {
	var args pb.SFU_Subscribe_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &args); err != nil {
			return nil, err
		}
	}

	args.PeerId = peerId

	var offer webrtc.SessionDescription
	{
		if err := json.Unmarshal(args.Description, &offer); err != nil {
			return nil, err
		}
	}

	peer := ionsfu.NewPeer(ionsfuServer)
	peer.OnOffer = func(offer *webrtc.SessionDescription) {
		bytes, err := json.Marshal(offer)
		if err != nil {
			return
		}

		reply := pb.SFU_Signal_Reply{
			SessionId: args.SessionId,
			PeerId:    args.PeerId,
		}

		reply.Payload = &pb.SFU_Signal_Reply_Description{
			Description: bytes,
		}

		b, _ := grpcproto.Marshal(&reply)
		resp := corespb.Response{Command: proto.NegotiateCommand.Int32()}
		resp.Payload = &corespb.Response_Content{
			Content: b,
		}

		client.Send(&resp)
	}

	peer.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
		bytes, err := json.Marshal(candidate)
		if err != nil {
			return
		}

		reply := pb.SFU_Signal_Reply{
			SessionId: args.SessionId,
			PeerId:    args.PeerId,
		}

		reply.Payload = &pb.SFU_Signal_Reply_Trickle{
			Trickle: &pb.SFU_Trickle{
				Target:    pb.SFU_Target(target),
				Candidate: string(bytes),
			},
		}
		b, _ := grpcproto.Marshal(&reply)
		resp := corespb.Response{Command: proto.NegotiateCommand.Int32()}
		resp.Payload = &corespb.Response_Content{
			Content: b,
		}

		client.Send(&resp)
	}

	err := peer.Join(args.SessionId, args.PeerId, ionsfu.JoinConfig{
		NoPublish:   false,
		NoSubscribe: false,
	})

	if err != nil {
		return nil, err
	}

	answer, err := peer.Answer(offer)
	if err != nil {
		return nil, err
	}

	answerBytes, err := json.Marshal(answer)
	if err != nil {
		return nil, err
	}

	session.AddPeer(peer)
	reply := pb.SFU_Subscribe_Reply{}
	reply.Ok = true
	reply.Description = answerBytes
	b, _ := grpcproto.Marshal(&reply)
	resp := corespb.Response{Command: proto.JoinCommand.Int32()}
	resp.Header = req.Header
	resp.Payload = &corespb.Response_Content{
		Content: b,
	}

	fmt.Println("headeer", resp)
	return &resp, nil
}

func negotiate(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SFU_Signal_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	resp := &corespb.Response{Command: req.Command}
	resp.Header = req.Header
	peerId, ok := req.Header["PeerId"]
	fmt.Println("PeerID", peerId, ok)
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
