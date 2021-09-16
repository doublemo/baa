package agent

import (
	"encoding/json"

	log "github.com/doublemo/baa/cores/log/level"
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/proto/pb"
	"github.com/doublemo/baa/kits/agent/session"
	awebrtc "github.com/doublemo/baa/kits/agent/webrtc"
	grpcproto "github.com/golang/protobuf/proto"
	"github.com/pion/webrtc/v3"
)

// useDataChannel 绑定datachannel
func useDataChannel(peer session.Peer) error {
	dc, err := peer.CreateDataChannel(awebrtc.Transport())
	if err != nil {
		return err
	}

	dc.OnICEConnectionStateChange(iceConnectionStateChange(peer))
	return nil
}

// datachannel 数据通道
func datachannel(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	var frame pb.Agent_Webrtc_Signal
	{
		if err := grpcproto.Unmarshal(req.Body(), &frame); err != nil {
			return nil, err
		}
	}

	switch payload := frame.Payload.(type) {
	case *pb.Agent_Webrtc_Signal_Description:
		return webrtcNewPeerConnection(peer, req, payload.Description)

	case *pb.Agent_Webrtc_Signal_Trickle:
		return webrtcTrickle(peer, req, payload)
	}

	return nil, nil
}

func webrtcNewPeerConnection(peer session.Peer, req coresproto.Request, payload []byte) (coresproto.Response, error) {
	resp := &coresproto.ResponseBytes{
		Ver:    req.V(),
		Cmd:    req.Command(),
		SubCmd: req.SubCommand(),
		SID:    req.SID(),
	}

	var sdp webrtc.SessionDescription
	{
		if err := json.Unmarshal(payload, &sdp); err != nil {
			return nil, err
		}
	}

	if sdp.Type == webrtc.SDPTypeAnswer {
		return nil, nil
	}

	answer, err := peer.DataChannel().Answer(sdp)
	if err != nil {
		log.Error(Logger()).Log("error", err, "peer_id", peer.ID())
		return nil, err
	}

	s := pb.Agent_Webrtc_Signal_Description{}
	s.Description, _ = json.Marshal(answer)
	resp.Content, _ = grpcproto.Marshal(&pb.Agent_Webrtc_Signal{Payload: &s})
	return resp, nil
}

func webrtcTrickle(peer session.Peer, req coresproto.Request, payload *pb.Agent_Webrtc_Signal_Trickle) (coresproto.Response, error) {
	var candidate webrtc.ICECandidateInit
	{
		if err := json.Unmarshal([]byte(payload.Trickle.Candidate), &candidate); err != nil {
			return nil, err
		}
	}

	if err := peer.DataChannel().AddICECandidate(candidate); err != nil {
		log.Error(Logger()).Log("AddICECandidate err:", err, "peer_id", peer.ID())
		return nil, err

	}
	return nil, nil
}

func iceConnectionStateChange(peer session.Peer) func(connectionState webrtc.ICEConnectionState) {
	return func(connectionState webrtc.ICEConnectionState) {
		resp := &coresproto.ResponseBytes{
			Ver:    1,
			Cmd:    proto.Agent,
			SubCmd: proto.DatachannelCommand,
			SID:    1,
		}

		w := pb.Agent_Webrtc_Signal{
			Payload: &pb.Agent_Webrtc_Signal_IceConnectionState{IceConnectionState: connectionState.String()},
		}

		resp.Content, _ = grpcproto.Marshal(&w)
		bytes, _ := resp.Marshal()
		peer.Send(session.PeerMessagePayload{Data: bytes})
	}
}
