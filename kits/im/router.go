package im

import (
	"fmt"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/im/proto"
	"github.com/doublemo/baa/kits/im/proto/pb"
	"github.com/doublemo/baa/kits/imf"
	imfproto "github.com/doublemo/baa/kits/imf/proto"
	"github.com/doublemo/baa/kits/snid"
	snproto "github.com/doublemo/baa/kits/snid/proto"
	"github.com/doublemo/baa/kits/usrt"
	usrtproto "github.com/doublemo/baa/kits/usrt/proto"
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
	r.HandleFunc(proto.SendCommand, func(req *corespb.Request) (*corespb.Response, error) { return send(req, config.Chat) })

	// 订阅处理
	nrRouter.Register(imf.ServiceName, router.New()).HandleFunc(imfproto.CheckCommand, handleMsgInspectionReport)

	// 注册内部使用路由
	snserv := newSnidRouter(config.ServiceSNID)
	muxRouter.Register(snid.ServiceName, router.New()).
		Handle(snproto.SnowflakeCommand, snserv).
		Handle(snproto.AutoincrementCommand, snserv)

	usrtserv := newUSRTRouter(config.ServiceUSRT)
	muxRouter.Register(usrt.ServiceName, router.New()).
		Handle(usrtproto.GetUserStatusCommand, usrtserv).
		Handle(usrtproto.DeleteUserStatusCommand, usrtserv).
		Handle(usrtproto.UpdateUserStatusCommand, usrtserv)
}

func testSend() {
	fmt.Println("testSend  start ........")
	req := &corespb.Request{
		Command: proto.SendCommand.Int32(),
		Header:  make(map[string]string),
	}

	frame := &pb.IM_Msg_Body{
		SeqID:       1,
		To:          "xxxx",
		Payload:     &pb.IM_Msg_Body_Text{Text: &pb.IM_Msg_Content_Text{Content: "你是不是个SB, 狗日的"}},
		From:        "test",
		FName:       "test",
		FHeadImgurl: "1233",
		FSex:        "1",
	}

	req.Payload, _ = grpcproto.Marshal(frame)
	for i := 0; i < 1; i++ {
		r.Handler(req)
	}
}
