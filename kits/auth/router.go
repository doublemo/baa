package auth

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/auth/proto"
	"github.com/doublemo/baa/kits/auth/router"
	"google.golang.org/grpc/resolver"
)

const (
	internalSnidRouter = 10001
)

var (
	r  = router.New()
	ir = router.New()
)

// RouterConfig 路由配置
type RouterConfig struct {
	ServiceSNID conf.RPCClient `alias:"snid"`
	SMS         SMSConfig      `alias:"sms"`
}

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	resolver.Register(coressd.NewResolverBuilder(config.ServiceSNID.Name, config.ServiceSNID.Group, sd.Endpointer()))

	// 注册处理请求
	r.HandleFunc(proto.LoginCommand, func(r *corespb.Request) (*corespb.Response, error) {
		return login(r, config.SMS)
	})

	// 注册内部使用路由
	ir.Handle(internalSnidRouter, newSnidRouter(config.ServiceSNID))
}
