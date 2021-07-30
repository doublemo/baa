package sfu

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/sfu/adapter/router"
	"github.com/doublemo/baa/kits/sfu/proto"
	"github.com/doublemo/baa/kits/sfu/proto/pb"
	"github.com/doublemo/baa/kits/sfu/session"
	grpcproto "github.com/golang/protobuf/proto"
	ionsfu "github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
	"github.com/smallnest/rpcx/server"
)

// ServiceName 服务名称
const ServiceName = "sfu"

// Configuration k
type Configuration struct {
	Ballast   int64               `alias:"ballast"`
	WithStats bool                `alias:"withstats"`
	WebRTC    ionsfu.WebRTCConfig `alias:"webrtc"`
	Router    ionsfu.RouterConfig `alias:"router"`
	Turn      ionsfu.TurnConfig   `alias:"turn"`
}

type sfuservice struct {
	server *server.Server
	sfu    *ionsfu.SFU
}

func (s *sfuservice) Subscribe(ctx context.Context, args *pb.SFU_Subscribe_Request, reply *pb.SFU_Subscribe_Reply) error {
	conn := ctx.Value(server.RemoteConnContextKey).(net.Conn)
	var offer webrtc.SessionDescription
	{
		if err := json.Unmarshal(args.Description, &offer); err != nil {
			return err
		}
	}

	fmt.Println("okkkk->", conn.RemoteAddr().String())

	peer := ionsfu.NewPeer(s.sfu)
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

		b2, _ := grpcproto.Marshal(&resp)
		s.server.SendMessage(conn, "sfu", "Subscribe", make(map[string]string), b2)
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

		b2, _ := grpcproto.Marshal(&resp)
		s.server.SendMessage(conn, "sfu", "Subscribe", make(map[string]string), b2)
	}

	err := peer.Join(args.SessionId, args.PeerId, ionsfu.JoinConfig{
		NoPublish:   false,
		NoSubscribe: false,
	})

	if err != nil {
		return err
	}

	answer, err := peer.Answer(offer)
	if err != nil {
		return err
	}

	answerBytes, err := json.Marshal(answer)
	if err != nil {
		return err
	}

	session.AddPeer(peer)
	reply.Ok = true
	reply.Description = answerBytes
	return nil
}

func (s *sfuservice) Call(ctx context.Context, args *corespb.Request, reply *corespb.Response) error {
	fn, err := router.Fn(coresproto.Command(args.Command))
	if err != nil {
		return err
	}

	resp, err := fn(args)
	if err != nil {
		return err
	}

	if resp == nil {
		return nil
	}

	reply.Command = resp.Command
	reply.Payload = resp.Payload
	return nil
}
