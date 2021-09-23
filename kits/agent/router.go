package agent

import (
	"crypto/rc4"
	"fmt"
	"math/big"

	"github.com/doublemo/baa/cores/crypto/dh"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/agent/errcode"
	midPeer "github.com/doublemo/baa/kits/agent/middlewares/peer"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/proto/pb"
	mrouter "github.com/doublemo/baa/kits/agent/router"
	"github.com/doublemo/baa/kits/agent/session"
	grpcproto "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/resolver"
)

var (
	sRouter = mrouter.New()
	dRouter = mrouter.New()
)

// RouterConfig 路由配置
type RouterConfig struct {
	ServiceSFU *conf.RPCClient `alias:"sfu"`
}

// InitRouter init
func InitRouter(config *RouterConfig) {
	// Register grpc load balance
	resolver.Register(coressd.NewResolverBuilder(config.ServiceSFU.Name, config.ServiceSFU.Group, sd.Endpointer()))

	// 注册处理socket/websocket来的请求
	sRouter.HandleFunc(proto.Agent, agentRouter)
	sRouter.Handle(proto.SFU, newSFURouter(config.ServiceSFU))

	// 注册处理datachannel来的请求
}

func onMessage(peer session.Peer, msg session.PeerMessagePayload) error {
	var (
		err  error
		resp coresproto.Response
	)

	switch peerLocal := peer.(type) {
	case *session.PeerSocket:
		if msg.Channel == session.PeerMessageChannelWebrtc {
			resp, err = handleFromDataChannelBinaryMessage(peer, msg.Data)
		} else {
			resp, err = handleBinaryMessage(peer, msg.Data)
		}

	case *session.PeerWebsocket:
		if msg.Channel == session.PeerMessageChannelWebrtc {
			resp, err = handleFromDataChannelBinaryMessage(peer, msg.Data)
		} else {
			if peerLocal.MessageType() == websocket.TextMessage {
				resp, err = handleTextMessage(peer, msg.Data)
			} else {
				resp, err = handleBinaryMessage(peer, msg.Data)
			}
		}
	}

	if resp != nil {
		bytes, err := resp.Marshal()
		if err != nil {
			return err
		}
		err = peer.Send(session.PeerMessagePayload{Channel: msg.Channel, Data: bytes})
	}

	return err
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

	resp, err := sRouter.Handler(peer, req)
	if resp != nil {
		resp.SeqID(req.SID())
	}

	return resp, err
}

func handleFromDataChannelBinaryMessage(peer session.Peer, frame []byte) (coresproto.Response, error) {
	req := &coresproto.RequestBytes{}
	if err := req.Unmarshal(frame); err != nil {
		return nil, errcode.ErrorInvalidProtoVersion.ToError()
	}

	resp, err := dRouter.Handler(peer, req)
	if resp != nil {
		resp.SeqID(req.SID())
	}

	return resp, err
}

func agentRouter(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	switch req.SubCommand() {
	case proto.HandshakeCommand:
		return handshake(peer, req)

	case proto.DatachannelCommand:
		return datachannel(peer, req)

	case proto.HeartbeaterCommand:
	}

	return nil, nil
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
