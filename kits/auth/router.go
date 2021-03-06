package auth

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
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
	muxRouter = router.NewMux()
)

// RouterConfig 路由配置
type RouterConfig struct {
	ServiceSNID conf.RPCClient `alias:"snid"`
	ServiceSM   conf.RPCClient `alias:"sm"`

	//LR  登录注册配置信息
	LR LRConfig `alias:"lr"`
}

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	resolver.Register(coressd.NewResolverBuilder(config.ServiceSNID.Name, config.ServiceSNID.Group, sd.Endpointer()))
	resolver.Register(coressd.NewResolverBuilder(config.ServiceSM.Name, config.ServiceSM.Group, sd.Endpointer()))

	// 注册处理请求
	r.HandleFunc(command.AuthRegister, func(r *corespb.Request) (*corespb.Response, error) {
		return register(r, config.LR)
	})

	r.HandleFunc(command.AuthLogin, func(r *corespb.Request) (*corespb.Response, error) {
		return login(r, config.LR)
	})

	r.HandleFunc(command.AuthOffline, func(r *corespb.Request) (*corespb.Response, error) { return offline(r, config.LR) })
	r.HandleFunc(command.AuthAccountInfo, func(r *corespb.Request) (*corespb.Response, error) { return accountInfo(r, config.LR) })
	r.HandleFunc(command.AuthorizedToken, func(r *corespb.Request) (*corespb.Response, error) { return authorizedToken(r, config.LR) })

	// 注册内部使用路由
	muxRouter.Register(kit.SNID.Int32(), router.New())
	muxRouter.Handle(kit.SNID.Int32(), command.SNIDSnowflake, newSnidRouter(config.ServiceSNID))

	sm := router.NewCall(config.ServiceSM)
	muxRouter.Register(kit.SM.Int32(), router.New()).Handle(command.SMUserStatus, sm).Handle(command.SMAssginServers, sm)
}
