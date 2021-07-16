package agent

import (
	"context"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpcx"
	"github.com/doublemo/baa/kits/agent/adapter/router"
	"github.com/doublemo/baa/kits/agent/errcode"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/session"
	sfuCommand "github.com/doublemo/baa/kits/sfu/proto"
	sfupb "github.com/doublemo/baa/kits/sfu/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	etcdClient "github.com/rpcxio/rpcx-etcd/client"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
)

// sfuRouter sfu 服务
func sfuRouter(config *RouterConfig) router.Callback {
	if config.Etcd == nil || config.Sfu == nil {
		panic("Invalid config in sfuRouter")
	}

	etcdDiscovery, err := etcdClient.NewEtcdV3Discovery(config.Etcd.BasePath, "sfu", config.Etcd.Addr, true, nil)
	if err != nil {
		panic(err)
	}

	// return
	return func(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
		if req.SubCommand() == sfuCommand.JoinCommand {
			return sfuSubscribe(peer, req, etcdDiscovery, *config.Sfu)
		}

		xclient := client.NewXClient("sfu", client.Failtry, client.RoundRobin, etcdDiscovery, rpcx.Option(*config.Sfu))
		defer func() {
			xclient.Close()
		}()

		// middleware
		rpcx.Use(xclient, *config.Sfu)
		args := &corespb.Request{
			Header:  map[string]string{"PeerId": peer.ID()},
			Command: req.SubCommand().Int32(),
			Payload: req.Body(),
		}

		reply := &corespb.Response{}
		err := xclient.Call(context.Background(), "Call", args, reply)
		if err != nil {
			reply.Command = req.SubCommand().Int32()
			return proto.NewResponseBytes(req.Command(), errcode.Bad(reply, errcode.ErrInternalServer, err.Error())), nil
		}

		if reply.Command < 1 {
			return nil, nil
		}

		return proto.NewResponseBytes(req.Command(), reply), nil
	}
}

func sfuSubscribe(peer session.Peer, req coresproto.Request, d client.ServiceDiscovery, option conf.RPCXClient) (coresproto.Response, error) {
	if m, ok := peer.Params("sfuXClient"); ok && m != nil {
		return nil, nil
	}

	ch := make(chan *protocol.Message)
	xclient := client.NewBidirectionalXClient("sfu", client.Failtry, client.RoundRobin, d, rpcx.Option(option), ch)
	rpcx.Use(xclient, option)

	var args sfupb.SFU_Subscribe_Request
	{
		if err := grpcproto.Unmarshal(req.Body(), &args); err != nil {
			return nil, err
		}
	}

	args.PeerId = peer.ID()
	reply := sfupb.SFU_Subscribe_Reply{}
	resp := corespb.Response{Command: req.SubCommand().Int32()}
	err := xclient.Call(context.Background(), "Subscribe", &args, &reply)
	if err != nil || !reply.Ok {
		xclient.Close()
		return nil, nil
	}

	go func(peerId string, messageChan chan *protocol.Message) {
		defer func() {
			close(messageChan)
		}()

		for frame := range messageChan {
			mStautsType := frame.Header.MessageStatusType()
			if mStautsType == protocol.Error || len(frame.Payload) < 1 {
				return
			}

			var data corespb.Response
			{
				if err := grpcproto.Unmarshal(frame.Payload, &data); err != nil {
					return
				}
			}

			w := proto.NewResponseBytes(proto.SFUCommand, &data)
			if sess, ok := session.GetPeer(peerId); ok && sess != nil {
				bytes, err := w.Marshal()
				if err != nil {
					continue
				}
				sess.Send(session.PeerMessagePayload{Type: websocket.BinaryMessage, Data: bytes})
			}

		}
	}(peer.ID(), ch)
	peer.SetParams("sfuXClient", xclient)

	bytes, _ := grpcproto.Marshal(&reply)
	resp.Payload = &corespb.Response_Content{Content: bytes}
	return proto.NewResponseBytes(req.Command(), &resp), nil
}

func sfuUnsubscribe(peer session.Peer) {
	sfuXClient, ok := peer.Params("sfuXClient")
	if !ok || sfuXClient == nil {
		return
	}

	xclient, ok := sfuXClient.(client.XClient)
	if !ok {
		return
	}
	xclient.Close()
	peer.SetParams("sfuXClient", nil)
}
