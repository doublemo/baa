package agent

import (
	"crypto/rc4"
	"fmt"
	"math/big"

	"github.com/doublemo/baa/cores/crypto/dh"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/agent/adapter/router"
	"github.com/doublemo/baa/kits/agent/errcode"
	midPeer "github.com/doublemo/baa/kits/agent/middlewares/peer"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/proto/pb"
	"github.com/doublemo/baa/kits/agent/session"
	grpcproto "github.com/golang/protobuf/proto"
)

func init() {
	router.On(proto.HandshakeCommand, handshake)
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

func taiRouter() router.Callback {
	return func(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
		return nil, nil
	}
}
