package im

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
	"github.com/doublemo/baa/kits/im/cache"
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
	ServiceSNID conf.RPCClient `alias:"snid"`
	ServiceUSRT conf.RPCClient `alias:"usrt"`
	Chat        ChatConfig     `alias:"chat"`
}

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	resolver.Register(coressd.NewResolverBuilder(config.ServiceSNID.Name, config.ServiceSNID.Group, sd.Endpointer()))
	resolver.Register(coressd.NewResolverBuilder(config.ServiceUSRT.Name, config.ServiceUSRT.Group, sd.Endpointer()))

	// 注册处理请求
	r.HandleFunc(command.IMSend, func(req *corespb.Request) (*corespb.Response, error) { return send(req, config.Chat) })

	// 订阅处理
	nrRouter.Register(kit.IMF.Int32(), router.New()).HandleFunc(command.IMFCheck, handleMsgInspectionReport)
	nrRouter.Register(kit.USRT.Int32(), router.New()).
		HandleFunc(command.USRTDeleteUserStatus, resetUserStatusCache).
		HandleFunc(command.USRTUpdateUserStatus, resetUserStatusCache)

	// 注册内部使用路由
	snserv := newSnidRouter(config.ServiceSNID)
	muxRouter.Register(kit.SNID.Int32(), router.New()).
		Handle(command.SNIDSnowflake, snserv).
		Handle(command.SNIDAutoincrement, snserv).
		Handle(command.SNIDMoreAutoincrement, newSnuidRouter(config.ServiceSNID))

	// cache
	cache.SnowflakeCacherOnFill(func(i int) ([]uint64, error) { return getSNID(int32(i)) })

	usrtserv := newUSRTRouter(config.ServiceUSRT)
	muxRouter.Register(kit.USRT.Int32(), router.New()).
		Handle(command.USRTGetUserStatus, usrtserv).
		Handle(command.USRTDeleteUserStatus, usrtserv).
		Handle(command.USRTUpdateUserStatus, usrtserv)

	time.AfterFunc(time.Second*10, testSend)
}

func testSend() {
	fmt.Println("testSend  start ........")
	req := &corespb.Request{
		Command: command.IMSend.Int32(),
		Header:  make(map[string]string),
	}

	frame := &pb.IM_Send{
		Messages: &pb.IM_Msg_List{
			Values: []*pb.IM_Msg_Content{
				&pb.IM_Msg_Content{
					SeqID:   1,
					To:      "NS07bbD2yLM",
					Payload: &pb.IM_Msg_Content_Text{Text: &pb.IM_Msg_ContentType_Text{Content: "你是不是个SB, 狗日的"}},
					From:    "2FRx9KAc-Jw",
					Group:   pb.IM_Msg_ToC,
				},
			},
		},
	}

	req.Payload, _ = grpcproto.Marshal(frame)
	//fmt.Println(r.Handler(req))

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
			r.Handler(req)
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
