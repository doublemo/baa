package sm

import (
	"fmt"
	"time"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/router"
	grpcproto "github.com/golang/protobuf/proto"
)

var (
	r               = router.New()
	nrRouter        = router.NewMux()
	nInternalRouter = router.New()
)

// RouterConfig 路由配置
type RouterConfig struct {
}

// InitRouter init
func InitRouter() {
	// Register grpc load balance

	// 注册处理请求
	r.HandleFunc(command.SMUserStatus, getUsersStatus)
	r.HandleFunc(command.SMBroadcastMessagesToAgent, broadcastMessagesToAgent)
	// r.HandleFunc(command.USRTDeleteUserStatus, deleteUserStatus)
	// r.HandleFunc(command.USRTGetUserStatus, getUserStatus)

	// 订阅处理
	nrRouter.Register(kit.SM.Int32(), router.New()).HandleFunc(command.SMEvent, eventHandler)
	nInternalRouter.HandleFunc(command.SMEvent, internalEventHandler)
	time.AfterFunc(time.Second*10, testSend)
}

func testSend() {
	fmt.Println("test start ........")
	req := &corespb.Request{
		Command: command.SMEvent.Int32(),
		Header:  map[string]string{"UserId": "XDzgT9EkkQc", "AccountID": "D5ChBlQ1da4"}, // "Content-Type": "json"
	}

	frame2 := &pb.SM_User_Action_Online{
		UserId:   344709394144956418,
		Platform: "phone",
		Agent:    "agent1.cn.sc.cd",
		Token:    "xxxxxx",
	}

	b, _ := grpcproto.Marshal(frame2)

	frame := &pb.SM_Event{
		Action: pb.SM_ActionUserOnline,
		Data:   b,
	}

	req.Payload, _ = grpcproto.Marshal(frame)
	fmt.Println(nrRouter.Handler(kit.SM.Int32(), req))
	//fmt.Println(id.Encrypt(344709394144956418, []byte("7581BDD8E8DA3839")))

}
