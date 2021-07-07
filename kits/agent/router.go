package agent

import (
	"context"
	"crypto/rc4"
	"fmt"
	"math/big"

	"github.com/doublemo/baa/cores/crypto/dh"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpcx"
	"github.com/doublemo/baa/kits/agent/adapter/router"
	"github.com/doublemo/baa/kits/agent/errcode"
	midPeer "github.com/doublemo/baa/kits/agent/middlewares/peer"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/proto/pb"
	"github.com/doublemo/baa/kits/agent/session"
	grpcproto "github.com/golang/protobuf/proto"
	etcdClient "github.com/rpcxio/rpcx-etcd/client"
	"github.com/smallnest/rpcx/client"
)

// RouterConfig 路由配置
type RouterConfig struct {
	Etcd *conf.Etcd       `alias:"etcd"`
	Sfu  *conf.RPCXClient `alias:"sfu"`
}

// InitRouter init
func InitRouter(config *RouterConfig) {
	router.On(proto.HandshakeCommand, handshake)
	router.On(proto.SFUCommand, sfuRouter(config))
}

func handleTextMessage(peer session.Peer, frame []byte) (coresproto.Response, error) {
	return nil, errcode.ErrorInvalidProtoVersion.ToError()
}

func handleBinaryMessage(peer session.Peer, frame []byte) (coresproto.Response, error) {
	req := &coresproto.RequestBytes{}
	if err := req.Unmarshal(frame); err != nil {
		return nil, errcode.ErrorInvalidProtoVersion.ToError()
	}

	if req.SID() != peer.LoadOrResetSeqNo() {
		return proto.NewResponseBytes(req.Cmd, errcode.Bad(&corespb.Response{Command: req.Command().Int32()}, errcode.ErrorInvalidSEQID)), nil
	}

	fn, err := router.Fn(req.Cmd)
	if err != nil {
		return proto.NewResponseBytes(req.Cmd, errcode.Bad(&corespb.Response{Command: req.Command().Int32()}, errcode.ErrCommandInvalid)), nil
	}
	return fn(peer, req)
}

// handshake rc4加密握手
func handshake(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	var frame pb.Agent_Handshake
	{
		if err := grpcproto.Unmarshal(req.Body(), &frame); err != nil {
			return nil, err
		}
	}

	x1, e1 := dh.DHExchange()
	x2, e2 := dh.DHExchange()
	key1 := dh.DHKey(x1, big.NewInt(frame.GetE1()))
	key2 := dh.DHKey(x2, big.NewInt(frame.GetE2()))

	frameResp := &pb.Agent_Handshake{
		E1: e1.Int64(),
		E2: e2.Int64(),
	}

	bytes, err := grpcproto.Marshal(frameResp)
	if err != nil {
		return nil, err
	}

	resp := &coresproto.ResponseBytes{
		Ver:     req.V(),
		Cmd:     req.Command(),
		SubCmd:  req.SubCommand(),
		SID:     req.SID(),
		Content: bytes,
	}

	encoder, err := rc4.NewCipher([]byte(fmt.Sprintf("%v%v", "DH", key2)))
	if err != nil {
		return nil, err
	}

	decoder, err := rc4.NewCipher([]byte(fmt.Sprintf("%v%v", "DH", key1)))
	if err != nil {
		return nil, err
	}

	peer.Use(midPeer.NewRC4(encoder, decoder))
	return resp, err
}

// sfuRouter sfu 服务
func sfuRouter(config *RouterConfig) router.Callback {
	if config.Etcd == nil || config.Sfu == nil {
		panic("Invalid config in sfuRouter")
	}

	etcdDiscovery, err := etcdClient.NewEtcdV3Discovery(config.Etcd.BasePath, "sfu", config.Etcd.Addr, true, nil)
	if err != nil {
		panic(err)
	}
	return func(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
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

		return proto.NewResponseBytes(req.Command(), reply), nil
	}
}
