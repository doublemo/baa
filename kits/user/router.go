package user

import (
	"fmt"
	"time"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/internal/sd"
	grpcproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc/resolver"
)

var (
	r         = router.New()
	nrRouter  = router.NewMux()
	muxRouter = router.NewMux()
)

// RouterConfig 路由配置
type RouterConfig struct {
	ServiceAuth conf.RPCClient `alias:"auth"`
	User        UserConfig     `alias:"usersettings"`
	Group       GroupConfig    `alias:"groupsettings"`
}

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	resolver.Register(coressd.NewResolverBuilder(config.ServiceAuth.Name, config.ServiceAuth.Group, sd.Endpointer()))

	// 注册处理请求
	r.HandleFunc(command.UserContacts, func(req *corespb.Request) (*corespb.Response, error) { return contact(req, config.User) })
	r.HandleFunc(command.UserContactsRequest, func(req *corespb.Request) (*corespb.Response, error) { return friendRequestList(req, config.User) })
	r.HandleFunc(command.UserRegister, func(req *corespb.Request) (*corespb.Response, error) { return register(req, config.User) })
	r.HandleFunc(command.UserCheckIsMyFriend, checkIsMyFriend)
	r.HandleFunc(command.UserCheckInGroup, checkInGroup)
	r.HandleFunc(command.UserGroupMembers, func(req *corespb.Request) (*corespb.Response, error) { return groupMembers(req, config.Group) })
	r.HandleFunc(command.UserGroupMembersValidID, func(req *corespb.Request) (*corespb.Response, error) { return groupMembersID(req, config.Group) })
	r.HandleFunc(command.UserInfo, func(req *corespb.Request) (*corespb.Response, error) { return getUserInfo(req, config.User) })

	// 内部调用
	muxRouter.Register(kit.Auth.Int32(), router.New()).Handle(command.AuthAccountInfo, router.NewCall(config.ServiceAuth))

	// 订阅处理
	// nrRouter.Register(kit.USRT.Int32(), router.New()).HandleFunc(command.USRTDeleteUserStatus, deleteUserStatus)

	//time.AfterFunc(time.Second*10, testSend)
}

func testSend() {
	fmt.Println("test start ........")
	req := &corespb.Request{
		Command: command.UserContacts.Int32(),
		Header:  map[string]string{"UserId": "XDzgT9EkkQc", "AccountID": "D5ChBlQ1da4"}, // "Content-Type": "json"
	}

	// frame0 := &pb.User_Register_Request{
	// 	AccountId: "D5ChBlQ1da4",
	// 	Info: &pb.User_Info{
	// 		UserId:   "",
	// 		Nickname: "小毛球",
	// 		Phone:    "13896936556",
	// 	},
	// }

	// frame1 := &pb.User_Register_Request{
	// 	AccountId: "jE6IsWqGCn4",
	// 	Info: &pb.User_Info{
	// 		UserId:   "",
	// 		Nickname: "球球",
	// 		Phone:    "13896936557",
	// 	},
	// }

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

	frame2 := &pb.User_Contacts_Request{
		Payload: &pb.User_Contacts_Request_Accept{
			Accept: &pb.User_Contacts_Accept{
				FriendId: "gcAKnnyepiE",
				Remark:   "XX",
				UserId:   "XDzgT9EkkQc",
			},
		},
	}

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

	req.Payload, _ = grpcproto.Marshal(frame2)
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
