package auth

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/auth/proto"
	"github.com/doublemo/baa/kits/snid"
	snproto "github.com/doublemo/baa/kits/snid/proto"
	"github.com/doublemo/baa/kits/usrt"
	usrtproto "github.com/doublemo/baa/kits/usrt/proto"
	"google.golang.org/grpc/resolver"
)

var (
	r         = router.New()
	muxRouter = router.NewMux()
)

// RouterConfig 路由配置
type RouterConfig struct {
	ServiceSNID conf.RPCClient `alias:"snid"`
	ServiceUSRT conf.RPCClient `alias:"usrt"`

	//LR  登录注册配置信息
	LR LRConfig `alias:"lr"`
}

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	resolver.Register(coressd.NewResolverBuilder(config.ServiceSNID.Name, config.ServiceSNID.Group, sd.Endpointer()))
	resolver.Register(coressd.NewResolverBuilder(config.ServiceUSRT.Name, config.ServiceUSRT.Group, sd.Endpointer()))

	// 注册处理请求
	r.HandleFunc(proto.RegisterCommand, func(r *corespb.Request) (*corespb.Response, error) {
		return register(r, config.LR)
	})

	r.HandleFunc(proto.LoginCommand, func(r *corespb.Request) (*corespb.Response, error) {
		return login(r, config.LR)
	})

	r.HandleFunc(proto.OfflineCommand, offline)

	// 注册内部使用路由
	muxRouter.Register(snid.ServiceName, router.New())
	muxRouter.Handle(snid.ServiceName, snproto.SnowflakeCommand, newSnidRouter(config.ServiceSNID))

	usrtr := newUSRTRouter(config.ServiceUSRT)
	muxRouter.Register(usrt.ServiceName, router.New())
	muxRouter.Handle(usrt.ServiceName, usrtproto.GetUserStatusCommand, usrtr)
	muxRouter.Handle(usrt.ServiceName, usrtproto.DeleteUserStatusCommand, usrtr)
	muxRouter.Handle(usrt.ServiceName, usrtproto.UpdateUserStatusCommand, usrtr)
}

func registerInternalCall(serviceName string, cmd coresproto.Command, handle router.Handler) {

}

func internalCall(serviceName string, cmd coresproto.Command, r *corespb.Request) (*corespb.Response, error) {
	return nil, nil
}
