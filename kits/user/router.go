package user

import (
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/internal/sd"
	"google.golang.org/grpc/resolver"
)

var (
	r         = router.New()
	nrRouter  = router.NewMux()
	muxRouter = router.NewMux()
)

// RouterConfig 路由配置
type RouterConfig struct {
	ServiceAuth conf.RPCClient
}

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	resolver.Register(coressd.NewResolverBuilder(config.ServiceAuth.Name, config.ServiceAuth.Group, sd.Endpointer()))

	// 注册处理请求
	// r.HandleFunc(command.USRTUpdateUserStatus, updateUserStatus)
	// r.HandleFunc(command.USRTDeleteUserStatus, deleteUserStatus)
	// r.HandleFunc(command.USRTGetUserStatus, getUserStatus)

	// 内部调用
	muxRouter.Register(kit.Auth.Int32(), router.New()).Handle(command.AuthAccountInfo, router.NewCall(config.ServiceAuth))

	// 订阅处理
	// nrRouter.Register(kit.USRT.Int32(), router.New()).HandleFunc(command.USRTDeleteUserStatus, deleteUserStatus)
}
