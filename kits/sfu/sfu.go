package sfu

import (
	"encoding/json"
	"fmt"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/sfu/errcode"
	"github.com/doublemo/baa/kits/sfu/proto"
	"github.com/doublemo/baa/kits/sfu/proto/pb"
	"github.com/doublemo/baa/kits/sfu/session"
	grpcproto "github.com/golang/protobuf/proto"
	ionsfu "github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
)

func NewSFUServer() {}

func join(peer session.Peer, r *corespb.Request) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: r.Command,
	}

	if r.Header != nil {
		w.Header = r.Header
	}

	var args pb.SFU_Subscribe_Request
	{
		if err := grpcproto.Unmarshal(r.Payload, &args); err != nil {
			return nil, err
		}
	}

	var offer webrtc.SessionDescription
	{
		if err := json.Unmarshal(args.Description, &offer); err != nil {
			return nil, err
		}
	}

	sfuPeer := peer.Peer().(*ionsfu.PeerLocal)

	// Notify user of new ice candidate
	sfuPeer.OnIceCandidate = makeOnIceCandidate(peer, &args)

	// Notify user of new offer
	sfuPeer.OnOffer = makeOnOffer(peer, &args)

	// join
	if err := sfuPeer.Join(args.SessionId, peer.ID(), ionsfu.JoinConfig{NoPublish: false, NoSubscribe: false}); err != nil {
		errcode.Bad(w, errcode.ErrInternalServer, err.Error())
		return w, nil
	}

	answer, err := sfuPeer.Answer(offer)
	if err != nil {
		errcode.Bad(w, errcode.ErrInternalServer, err.Error())
		return w, nil
	}

	answerBytes, _ := json.Marshal(answer)
	peer.Peer(sfuPeer)
	reply := pb.SFU_Subscribe_Reply{
		Ok:          true,
		Description: answerBytes,
	}

	b, _ := grpcproto.Marshal(&reply)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func negotiate(peer session.Peer, r *corespb.Request) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: r.Command,
	}

	if r.Header != nil {
		w.Header = r.Header
	}

	var frame pb.SFU_Signal_Request
	{
		if err := grpcproto.Unmarshal(r.Payload, &frame); err != nil {
			errcode.Bad(w, errcode.ErrInternalServer, err.Error())
			return w, nil
		}
	}

	switch payload := frame.Payload.(type) {
	case *pb.SFU_Signal_Request_Description:
		return negotiateBySDP(peer, w, payload.Description, &frame)

	case *pb.SFU_Signal_Request_Trickle:
		return negotiateByTrickle(peer, w, payload.Trickle, &frame)
	}

	return nil, nil
}

func negotiateBySDP(peer session.Peer, w *corespb.Response, payload []byte, frame *pb.SFU_Signal_Request) (*corespb.Response, error) {
	var sdp webrtc.SessionDescription
	{
		if err := json.Unmarshal(payload, &sdp); err != nil {
			errcode.Bad(w, errcode.ErrInternalServer, err.Error())
			return w, nil
		}
	}

	sfuPeerOld := peer.Peer()
	if sfuPeerOld == nil {
		errcode.Bad(w, errcode.ErrInternalServer, "sfu peer is nil")
		return w, nil
	}

	sfuPeer := sfuPeerOld.(*ionsfu.PeerLocal)

	switch sdp.Type {
	case webrtc.SDPTypeOffer:
		answer, err := sfuPeer.Answer(sdp)
		if err != nil {
			errcode.Bad(w, errcode.ErrInternalServer, err.Error())
			return w, nil
		}

		bytes, _ := json.Marshal(answer)
		bytes2, _ := grpcproto.Marshal(&pb.SFU_Signal_Reply{
			SessionId: frame.SessionId,
			PeerId:    peer.ID(),
			Payload: &pb.SFU_Signal_Reply_Description{
				Description: bytes,
			},
		})
		w.Payload = &corespb.Response_Content{Content: bytes2}
		return w, nil

	case webrtc.SDPTypeAnswer:
		if err := sfuPeer.SetRemoteDescription(sdp); err != nil {
			errcode.Bad(w, errcode.ErrInternalServer, err.Error())
			return w, nil
		}
	}

	return nil, nil
}

func negotiateByTrickle(peer session.Peer, w *corespb.Response, payload *pb.SFU_Trickle, frame *pb.SFU_Signal_Request) (*corespb.Response, error) {
	fmt.Println(payload.Target, payload.Candidate)
	var candidate webrtc.ICECandidateInit
	{
		if err := json.Unmarshal([]byte(payload.Candidate), &candidate); err != nil {
			errcode.Bad(w, errcode.ErrInternalServer, err.Error())
			return w, nil
		}
	}

	sfuPeerOld := peer.Peer()
	if sfuPeerOld == nil {
		errcode.Bad(w, errcode.ErrInternalServer, "sfu peer is nil")
		return w, nil
	}

	sfuPeer := sfuPeerOld.(*ionsfu.PeerLocal)
	sfuPeer.Trickle(candidate, int(payload.Target))
	return nil, nil
}

func makeOnOffer(peer session.Peer, r *pb.SFU_Subscribe_Request) func(*webrtc.SessionDescription) {
	return func(offer *webrtc.SessionDescription) {
		bytes, _ := json.Marshal(offer)
		w := pb.SFU_Signal_Reply{
			SessionId: r.SessionId,
			PeerId:    peer.ID(),
		}

		w.Payload = &pb.SFU_Signal_Reply_Description{Description: bytes}
		b, _ := grpcproto.Marshal(&w)
		resp := corespb.Response{Command: proto.NegotiateCommand.Int32()}
		resp.Payload = &corespb.Response_Content{Content: b}
		peer.Send(&resp)
	}
}

func makeOnIceCandidate(peer session.Peer, r *pb.SFU_Subscribe_Request) func(*webrtc.ICECandidateInit, int) {
	return func(candidate *webrtc.ICECandidateInit, target int) {
		bytes, _ := json.Marshal(candidate)
		w := pb.SFU_Signal_Reply{
			SessionId: r.SessionId,
			PeerId:    peer.ID(),
		}

		w.Payload = &pb.SFU_Signal_Reply_Trickle{Trickle: &pb.SFU_Trickle{
			Target:    pb.SFU_Target(target),
			Candidate: string(bytes),
		}}

		b, _ := grpcproto.Marshal(&w)
		resp := corespb.Response{Command: proto.NegotiateCommand.Int32()}
		resp.Payload = &corespb.Response_Content{Content: b}
		peer.Send(&resp)
	}
}
