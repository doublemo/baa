package robot

import (
	"fmt"
	"time"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	ir "github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/robot/router"
	"github.com/doublemo/baa/kits/robot/session"
	grpcproto "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/resolver"
)

var (
	r         = ir.New()
	nrRouter  = ir.NewMux()
	muxRouter = ir.NewMux()
	netRouter = router.New()
	dcRouter  = router.New()
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

	time.AfterFunc(time.Second*10, testSend)
}

func testSend() {
	fmt.Println("test start ........")
	req := &corespb.Request{
		Command: command.RobotCreate.Int32(),
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

	frame0 := &pb.Robot_Create_Request{
		Payload: &pb.Robot_Create_Request_Register{
			Register: &pb.Robot_Create_Register{
				Schema:   "password",
				Name:     "robot03",
				Secret:   "Aa123456",
				Nickname: "Robot03",
				Headimg:  "xxx",
				Age:      28,
				Sex:      2,
				Idcard:   "12555555",
				Phone:    "13896968989",
			},
		},
	}

	// frame2 := &pb.User_Contacts_Request{
	// 	Payload: &pb.User_Contacts_Request_Add{
	// 		Add: &pb.User_Contacts_Add{
	// 			FriendId: "XDzgT9EkkQc", // 344702845066416128 344702845066416130
	// 			Remark:   "小主人",
	// 			Message:  "小主加我为好友吧",
	// 			UserId:   "gcAKnnyepiE", // 344705556818169856 344705556818169858
	// 		},
	// 	},
	// }

	// frame2 := &pb.User_Contacts_Request{
	// 	Payload: &pb.User_Contacts_Request_Accept{
	// 		Accept: &pb.User_Contacts_Accept{
	// 			FriendId: "gcAKnnyepiE",
	// 			Remark:   "XX",
	// 			UserId:   "XDzgT9EkkQc",
	// 		},
	// 	},
	// }

	// frame2 := &pb.User_Contacts_Request{
	// 	Payload: &pb.User_Contacts_Request_Refuse{
	// 		Refuse: &pb.User_Contacts_Refuse{
	// 			FriendId: "gcAKnnyepiE",
	// 			Message:  "我不认识你，你是谁呀 OR 1=1--dddd'dddd OR 1=1",
	// 			UserId:   "XDzgT9EkkQc",
	// 		},
	// 	},
	// }

	// frame2 := &pb.User_Contacts_FriendRequestList{
	// 	UserId:  "XDzgT9EkkQc",
	// 	Page:    1,
	// 	Size:    10,
	// 	Version: 0,
	// }

	req.Payload, _ = grpcproto.Marshal(frame0)
	// jsonpbM := &jsonpb.Marshaler{}
	// json, _ := jsonpbM.MarshalToString(frame2)
	// req.Payload = []byte(json)
	fmt.Println(r.Handler(req))

	//fmt.Println(id.Encrypt(344709394144956418, []byte("7581BDD8E8DA3839")))
	// snserv := newSnidRouter(conf.RPCClient{
	// 	Name:               "snid",
	// 	Weight:             10,
	// 	Group:              "prod",
	// 	Salt:               "certs/x509/ca_cert.pem",
	// 	Key:                "x.test.example.com",
	// 	ServiceSecurityKey: "baa",
	// 	Pool: conf.RPCPool{
	// 		Init:        1,
	// 		Capacity:    10,
	// 		IdleTimeout: 1,
	// 		MaxLife:     1,
	// 	},
	// })
	// muxRouter.Register(snid.ServiceName, router.New()).
	// 	Handle(snproto.SnowflakeCommand, snserv).
	// 	Handle(snproto.AutoincrementCommand, snserv)

	for i := 0; i < 1; i++ {
		go func(idx int) {
			//r.Handler(req)
			// for x := 0; x < 10; x++ {
			// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			// 	id, err := cache.GetSnowflakeID(ctx)
			// 	cancel()
			// 	if err != nil {
			// 		fmt.Println(err)
			// 	}

			// 	if id < 1 {
			// 		panic("id zero")
			// 	}

			// 	fmt.Println("id-->", id)
			// }
		}(i)

		time.Sleep(time.Millisecond * 1)
	}
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

	return netRouter.Handler(peer, w)
}

func handleFromDataChannelBinaryMessage(peer session.Peer, frame []byte) error {
	w := &coresproto.ResponseBytes{}
	if err := w.Unmarshal(frame); err != nil {
		return err
	}
	return dcRouter.Handler(peer, w)
}
