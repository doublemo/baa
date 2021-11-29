package robot

import (
	"crypto/rc4"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/doublemo/baa/cores/crypto/dh"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	ir "github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/internal/sd"
	midPeer "github.com/doublemo/baa/kits/robot/middlewares/peer"
	"github.com/doublemo/baa/kits/robot/router"
	"github.com/doublemo/baa/kits/robot/session"
	grpcproto "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/resolver"
)

var (
	r            = ir.New()
	nrRouter     = ir.NewMux()
	muxRouter    = ir.NewMux()
	socketRouter = router.New()
	dcRouter     = router.New()
)

// RouterConfig 路由配置
type RouterConfig struct {
	ServiceAuth conf.RPCClient `alias:"auth"`
	ServiceUser conf.RPCClient `alias:"user"`
	Robot       RobotConfig    `alias:"robotsettings"`
}

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	resolver.Register(coressd.NewResolverBuilder(config.ServiceAuth.Name, config.ServiceAuth.Group, sd.Endpointer()))
	resolver.Register(coressd.NewResolverBuilder(config.ServiceUser.Name, config.ServiceUser.Group, sd.Endpointer()))

	// 注册处理请求
	r.HandleFunc(command.RobotCreate, func(req *corespb.Request) (*corespb.Response, error) { return createRobot(req, config.Robot) })
	r.HandleFunc(command.RobotStart, func(req *corespb.Request) (*corespb.Response, error) { return startRobots(req, config.Robot) })

	// 内部调用
	auth := ir.NewCall(config.ServiceAuth)
	muxRouter.Register(kit.Auth.Int32(), ir.New()).
		Handle(command.AuthLogin, auth).
		Handle(command.AuthRegister, auth)

	user := ir.NewCall(config.ServiceUser)
	muxRouter.Register(kit.User.Int32(), ir.New()).
		Handle(command.UserInfo, user).
		Handle(command.UserRegister, user)

	// 订阅处理
	// nrRouter.Register(kit.USRT.Int32(), router.New()).HandleFunc(command.USRTDeleteUserStatus, deleteUserStatus)

	// socket 请求
	socketRouter.HandleFunc(kit.Agent, command.AgentHandshake, func(peer session.Peer, w coresproto.Response) error { return handshake(peer, w, config.Robot) })
	socketRouter.HandleFunc(kit.Agent, command.AgentDatachannel, func(peer session.Peer, w coresproto.Response) error { return datachannel(peer, w) })
	socketRouter.HandleFunc(kit.Agent, command.AgentHeartbeater, heartbeater)
	socketRouter.HandleFunc(kit.Auth, command.AuthLogin, func(peer session.Peer, w coresproto.Response) error { return login(peer, w, config.Robot) })

	dcRouter.HandleFunc(kit.IM, command.IMSend, func(peer session.Peer, w coresproto.Response) error {
		return handleIMNotify(peer, w, config.Robot)
	})
	dcRouter.HandleFunc(kit.IM, command.IMPush, func(peer session.Peer, w coresproto.Response) error { return handleIMNotify(peer, w, config.Robot) })

	time.AfterFunc(time.Second*10, testSend)
}

func testSend() {
	fmt.Println("test start ........")
	req := &corespb.Request{
		Command: command.RobotStart.Int32(),
		Header:  map[string]string{"UserId": "XDzgT9EkkQc", "AccountID": "D5ChBlQ1da4"}, // "Content-Type": "json"
	}

	// frame0 := &pb.Robot_Create_Request{
	// 	Payload: &pb.Robot_Create_Request_Account{
	// 		Account: &pb.Robot_Create_Account{
	// 			Name:   "4588ut",
	// 			Secret: "Aa123456",
	// 		},
	// 	},
	// }

	// frame0 := &pb.Robot_Create_Request{
	// 	Payload: &pb.Robot_Create_Request_Register{
	// 		Register: &pb.Robot_Create_Register{
	// 			Schema:   "password",
	// 			Name:     "robot05",
	// 			Secret:   "Aa123456",
	// 			Nickname: "Robot05",
	// 			Headimg:  "xxx",
	// 			Age:      28,
	// 			Sex:      2,
	// 			Idcard:   "12555555",
	// 			Phone:    "13896968989",
	// 		},
	// 	},
	// }

	frame2 := &pb.Robot_Start_Request{
		Values: []*pb.Robot_Start_Robot{
			&pb.Robot_Start_Robot{ID: 1, TaskGroup: 0},
			// &pb.Robot_Start_Robot{ID: 2, TaskGroup: 0},
			&pb.Robot_Start_Robot{ID: 3, TaskGroup: 0},
		},
		Async: true,
	}

	req.Payload, _ = grpcproto.Marshal(frame2)
	// jsonpbM := &jsonpb.Marshaler{}
	// json, _ := jsonpbM.MarshalToString(frame2)
	// req.Payload = []byte(json)
	fmt.Println(r.Handler(req))

	// for i := 0; i < 100; i++ {
	// 	go func(idx int) {
	// 		frame := &pb.Robot_Create_Request{
	// 			Payload: &pb.Robot_Create_Request_Register{
	// 				Register: &pb.Robot_Create_Register{
	// 					Schema:   "password",
	// 					Name:     "robot10" + strconv.FormatInt(int64(idx), 10),
	// 					Secret:   "Aa123456",
	// 					Nickname: "Robot10" + strconv.FormatInt(int64(idx), 10),
	// 					Headimg:  "xxx",
	// 					Age:      28,
	// 					Sex:      2,
	// 					Idcard:   "12555555",
	// 					Phone:    "13896968989",
	// 				},
	// 			},
	// 		}

	// 		req := &corespb.Request{
	// 			Command: command.RobotCreate.Int32(),
	// 			Header:  map[string]string{}, // "Content-Type": "json"
	// 		}

	// 		req.Payload, _ = grpcproto.Marshal(frame)
	// 		fmt.Println(r.Handler(req))
	// 	}(i)

	// 	time.Sleep(time.Millisecond * 1)
	// }
}

func onMessage(peer session.Peer, msg session.PeerMessagePayload) error {
	var (
		err error
	)

	switch peerLocal := peer.(type) {
	case *session.PeerSocket:
		if msg.Channel == session.PeerMessageChannelWebrtc {
			err = handleFromDataChannelBinaryMessage(peer, msg.Data)
		} else {
			err = handleBinaryMessage(peer, msg.Data)
		}

	case *session.PeerWebsocket:
		if msg.Channel == session.PeerMessageChannelWebrtc {
			err = handleFromDataChannelBinaryMessage(peer, msg.Data)
		} else {
			if peerLocal.MessageType() == websocket.TextMessage {
				err = handleTextMessage(peer, msg.Data)
			} else {
				err = handleBinaryMessage(peer, msg.Data)
			}
		}
	}

	return err
}

func handleTextMessage(peer session.Peer, frame []byte) error {
	return nil
}

func handleBinaryMessage(peer session.Peer, frame []byte) error {
	w := &coresproto.ResponseBytes{}
	if err := w.Unmarshal(frame); err != nil {
		return err
	}

	if w.Code == 65544 || w.Code == 65552 || w.Code == 65560 || w.Code == 69536 {
		return fmt.Errorf("Code: %d, error:%s, command:%d, subcommand:%d", w.Code, string(w.Body()), w.Cmd, w.SubCmd)
	}

	return socketRouter.Handler(peer, w)
}

func handleFromDataChannelBinaryMessage(peer session.Peer, frame []byte) error {
	w := &coresproto.ResponseBytes{}
	if err := w.Unmarshal(frame); err != nil {
		return err
	}

	return dcRouter.Handler(peer, w)
}

func doRequestHandshake(peer session.Peer) error {
	x1, e1 := dh.DHExchange()
	x2, e2 := dh.DHExchange()
	peer.SetParams("x1", x1)
	peer.SetParams("x2", x2)
	peer.SetParams("e1", e1)
	peer.SetParams("e2", e2)

	frame := &pb.Agent_Handshake{E1: e1.Int64(), E2: e2.Int64()}
	bytes, err := grpcproto.Marshal(frame)
	if err != nil {
		return err
	}

	req := &coresproto.RequestBytes{
		Ver:     1,
		Cmd:     kit.Agent,
		SubCmd:  command.AgentHandshake,
		Content: bytes,
		SeqID:   2,
	}

	r, err := req.Marshal()
	if err != nil {
		return err
	}

	return peer.Send(session.PeerMessagePayload{Data: r})
}

// handshake rc4加密握手
func handshake(peer session.Peer, w coresproto.Response, c RobotConfig) error {
	var frame pb.Agent_Handshake
	{
		if err := grpcproto.Unmarshal(w.Body(), &frame); err != nil {
			return err
		}
	}

	v, ok := peer.Params("x1")
	if !ok {
		return errors.New("handle handshake failed")
	}

	x1 := v.(*big.Int)

	v, ok = peer.Params("x2")
	if !ok {
		return errors.New("handle handshake failed")
	}
	x2 := v.(*big.Int)
	k1 := dh.DHKey(x1, big.NewInt(frame.GetE1()))
	k2 := dh.DHKey(x2, big.NewInt(frame.GetE2()))
	encoder, err := rc4.NewCipher([]byte(fmt.Sprintf("%v%v", "DH", k1)))
	if err != nil {
		return err
	}

	decoder, err := rc4.NewCipher([]byte(fmt.Sprintf("%v%v", "DH", k2)))
	if err != nil {
		return err
	}

	peer.Use(midPeer.NewRC4(encoder, decoder))
	return doRequestLogin(peer, c)
}

func doRequestHeartbeater(peer session.Peer) error {
	frame := &pb.Agent_Heartbeater{
		R: 1,
	}

	bytes, err := grpcproto.Marshal(frame)
	if err != nil {
		return err
	}

	req := &coresproto.RequestBytes{
		Ver:     1,
		Cmd:     kit.Agent,
		SubCmd:  command.AgentHeartbeater,
		Content: bytes,
		SeqID:   1,
	}

	r, err := req.Marshal()
	if err != nil {
		return err
	}

	return peer.Send(session.PeerMessagePayload{Data: r})
}

func heartbeater(peer session.Peer, w coresproto.Response) error {
	var frame pb.Agent_Heartbeater
	{
		if err := grpcproto.Unmarshal(w.Body(), &frame); err != nil {
			return nil
		}
	}
	fmt.Println("heartbeater ->", frame.R)
	return nil
}
