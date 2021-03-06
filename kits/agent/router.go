package agent

import (
	"crypto/rc4"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/doublemo/baa/cores/crypto/dh"
	log "github.com/doublemo/baa/cores/log/level"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	irouter "github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/agent/errcode"
	"github.com/doublemo/baa/kits/agent/middlewares/interceptor"
	midPeer "github.com/doublemo/baa/kits/agent/middlewares/peer"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/router"
	"github.com/doublemo/baa/kits/agent/session"
	grpcproto "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/resolver"
)

var (
	// sRouter socket/websocket
	sRouter = router.New()

	// dRouter webrtc datachannel
	dRouter = router.New()

	// nRouter nats Subscribe
	nrRouter = irouter.NewMux()

	// muxRouter 内部调用处理
	muxRouter = irouter.NewMux()
)

type (
	// RouterConfig 路由配置
	RouterConfig struct {
		Routes       []Router           `alias:"routes"`
		HttpConfigV1 HttpRouterV1Config `alias:"httpv1"`
	}

	Router struct {
		KitID            int32          `alias:"kitid"`
		Authorization    bool           `alias:"authorization"`
		Config           conf.RPCClient `alias:"config"`
		NetProtocol      string         `alias:"net"`
		ContentType      string         `alias:"contentType"`
		Commands         []int32        `alias:"commands"`
		SkipAuthCommands []int32        `alias:"skipAuthCommands"`
	}
)

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	// 注册处理socket/websocket来的请求
	// 注册处理datachannel来的请求
	kitConfigs := make(map[coresproto.Command]Router)
	for _, route := range config.Routes {
		kitConfigs[coresproto.Command(route.KitID)] = route
		if route.KitID == kit.Agent.Int32() {
			continue
		}

		if m := resolver.Get(route.Config.Name); m == nil {
			resolver.Register(coressd.NewResolverBuilder(route.Config.Name, route.Config.Group, sd.Endpointer()))
		}

		if len(route.Commands) < 1 {
			continue
		}

		net := strings.ToLower(route.NetProtocol)
		interceptors := router.WithRequestInterceptor(interceptor.AllowCommands(route.Commands...))

		if route.Authorization {
			interceptors = append(interceptors, interceptor.Authenticate(route.SkipAuthCommands...))
		}

		switch net {
		case "socket", "websocket", "tcp":
			if route.ContentType == "stream" {
				streamCall := router.NewStream(route.Config, Logger())
				streamCall.UseRequestInterceptor(interceptors...)
				streamCall.UseResponseInterceptor(interceptor.OnStreamReceive(coresproto.Command(route.KitID), false))
				sRouter.Handle(coresproto.Command(route.KitID), streamCall)
				continue
			}

			call := router.NewCall(route.Config, Logger())
			if route.Config.Name == kit.AuthServiceName {
				call.UseDestroyInterceptor(interceptor.OnOfflineRouterDestroy(muxRouter))
				call.UseResponseInterceptor(interceptor.OnLogin)
			} else if route.Config.Name == kit.IMFServiceName {
				interceptors = append(interceptors, interceptor.OnSelectIMServer)
			} else if route.Config.Name == kit.UserServiceName {
				interceptors = append(interceptors, interceptor.AddIMServerToHeader)
			}

			call.UseRequestInterceptor(interceptors...)
			sRouter.Handle(coresproto.Command(route.KitID), call)

		case "udp", "datachannel":
			if route.ContentType == "stream" {
				streamCall := router.NewStream(route.Config, Logger())
				streamCall.UseRequestInterceptor(interceptors...)
				streamCall.UseResponseInterceptor(interceptor.OnStreamReceive(coresproto.Command(route.KitID), true))
				sRouter.Handle(coresproto.Command(route.KitID), streamCall)
				continue
			}

			if route.Config.Name == kit.IMServiceName {
				interceptors = append(interceptors, interceptor.OnSelectIMServer)
			} else if route.Config.Name == kit.UserServiceName {
				interceptors = append(interceptors, interceptor.AddIMServerToHeader)
			}

			call := router.NewCall(route.Config, Logger())
			call.UseRequestInterceptor(interceptors...)
			dRouter.Handle(coresproto.Command(route.KitID), call)
		}
	}
	sRouter.HandleFunc(kit.Agent, agentRouter)

	// 内部请求处理
	if m, ok := kitConfigs[kit.Auth]; ok {
		authRouter := irouter.NewCall(m.Config)
		muxRouter.Register(kit.Auth.Int32(), irouter.New()).Handle(command.AuthorizedToken, authRouter).Handle(command.AuthOffline, authRouter)
	}

	// 注册处理nats订阅的请求
	nrRouter.Register(kit.Agent.Int32(), irouter.New()).HandleFunc(command.AgentKickedOut, kickedOut).HandleFunc(command.AgentBroadcast, broadcast)
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
		return proto.NewResponseBytes(req.Cmd, errcode.Bad(&corespb.Response{Command: req.SubCmd.Int32()}, errcode.ErrorInvalidSEQID)), nil
	}

	resp, err := sRouter.Handler(peer, req)
	if resp != nil {
		resp.SeqID(req.SID())
	}

	if err == router.ErrNotFoundRouter {
		return proto.NewResponseBytes(req.Cmd, errcode.Bad(&corespb.Response{Command: req.SubCmd.Int32()}, errcode.ErrCommandInvalid)), nil
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

	if err == router.ErrNotFoundRouter {
		return proto.NewResponseBytes(req.Cmd, errcode.Bad(&corespb.Response{Command: req.Command().Int32()}, errcode.ErrCommandInvalid)), nil
	}

	return resp, err
}

func agentRouter(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	switch req.SubCommand() {
	case command.AgentHandshake:
		return handshake(peer, req)

	case command.AgentDatachannel:
		return datachannel(peer, req)

	case command.AgentHeartbeater:
		return heartbeater(peer, req)
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

func kickedOut(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.Agent_KickedOut
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &coresproto.ResponseBytes{
		Ver:    1,
		Cmd:    kit.Agent,
		SubCmd: command.AgentKickedOut,
		SID:    1,
	}

	for _, id := range frame.PeerID {
		if m, ok := session.GetPeer(id); ok {

			frame := pb.Agent_KickedOut{
				PeerID: []string{id},
			}

			w.Content, _ = grpcproto.Marshal(&frame)
			resp, _ := w.Marshal()
			m.Send(session.PeerMessagePayload{Data: resp})
			m.Close()
		}
	}
	return nil, nil
}

func broadcast(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.Agent_Broadcast
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	for _, value := range frame.Messages {
		w := &corespb.Response{
			Command: value.SubCommand,
			Payload: &corespb.Response_Content{Content: value.Payload},
		}
		msg, _ := proto.NewResponseBytes(coresproto.Command(value.Command), w).Marshal()
		for _, recv := range value.Receiver {
			peers, ok := session.GetDict(strconv.FormatUint(recv, 10))
			if !ok {
				continue
			}

			for _, id := range peers {
				peer, ok := session.GetPeer(id)
				if !ok {
					continue
				}

				if err := peer.Send(session.PeerMessagePayload{Channel: session.PeerMessageChannelWebrtc, Data: msg}); err != nil {
					log.Error(Logger()).Log("action", "broadcast", "error", err)
				}
			}
		}
	}
	return nil, nil
}

func heartbeater(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	var frame pb.Agent_Heartbeater
	{
		if err := grpcproto.Unmarshal(req.Body(), &frame); err != nil {
			return nil, err
		}
	}

	frameResp := &pb.Agent_Heartbeater{
		R: time.Now().Unix(),
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
	return resp, err
}
