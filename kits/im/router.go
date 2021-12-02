package im

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/im/cache"
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
	ServiceSM   conf.RPCClient `alias:"sm"`
	ServiceUser conf.RPCClient `alias:"user"`
	Chat        ChatConfig     `alias:"chat"`
}

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	resolver.Register(coressd.NewResolverBuilder(config.ServiceSNID.Name, config.ServiceSNID.Group, sd.Endpointer()))
	resolver.Register(coressd.NewResolverBuilder(config.ServiceSM.Name, config.ServiceSM.Group, sd.Endpointer()))
	resolver.Register(coressd.NewResolverBuilder(config.ServiceUser.Name, config.ServiceUser.Group, sd.Endpointer()))

	// 注册处理请求
	r.HandleFunc(command.IMSend, func(req *corespb.Request) (*corespb.Response, error) { return send(req, config.Chat) })

	// 订阅处理
	nrRouter.Register(kit.IMF.Int32(), router.New()).HandleFunc(command.IMFCheck, handleMsgInspectionReport)
	// nrRouter.Register(kit.USRT.Int32(), router.New()).
	// 	HandleFunc(command.USRTDeleteUserStatus, resetUserStatusCache).
	// 	HandleFunc(command.USRTUpdateUserStatus, resetUserStatusCache)

	// 注册内部使用路由
	snserv := router.NewCall(config.ServiceSNID)
	muxRouter.Register(kit.SNID.Int32(), router.New()).
		Handle(command.SNIDSnowflake, snserv).
		Handle(command.SNIDAutoincrement, snserv).
		Handle(command.SNIDMoreAutoincrement, snserv)

	// cache
	cache.SnowflakeCacherOnFill(func(i int) ([]uint64, error) { return getSNID(int32(i)) })

	sm := router.NewCall(config.ServiceSM)
	muxRouter.Register(kit.SM.Int32(), router.New()).Handle(command.SMUserServers, sm)

	user := router.NewCall(config.ServiceUser)
	muxRouter.Register(kit.User.Int32(), router.New()).Handle(command.UserCheckIsMyFriend, user).
		Handle(command.UserCheckInGroup, user).
		Handle(command.UserGroupMembers, user).
		Handle(command.UserGroupMembersValidID, user)
}
