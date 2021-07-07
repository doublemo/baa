package sfu

import (
	"context"
	"net"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/sfu/adapter/router"
	"github.com/doublemo/baa/kits/sfu/proto/pb"
	"github.com/doublemo/baa/kits/sfu/session"
	"github.com/smallnest/rpcx/server"
)

type sfuservice struct {
	server *server.Server
}

func (s *sfuservice) Subscribe(ctx context.Context, args *pb.SFU_Subscribe_Request, reply *pb.SFU_Subscribe_Reply) error {
	conn := ctx.Value(server.RemoteConnContextKey).(net.Conn)
	peer := session.NewPeerLocal(args.PeerId)
	peer.OnNotify(func(payload []byte) error {
		return s.server.SendMessage(conn, "sfu", "Subscribe", make(map[string]string), payload)
	})

	session.AddPeer(peer, args.SessionId)
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

	reply.Command = resp.Command
	reply.Payload = resp.Payload
	return nil
}
