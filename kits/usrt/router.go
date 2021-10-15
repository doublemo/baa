package usrt

import (
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/router"
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
}

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	//resolver.Register(coressd.NewResolverBuilder(config.ServiceSNID.Name, config.ServiceSNID.Group, sd.Endpointer()))

	// 注册处理请求

	// 注册内部使用路由
	//ir.Handle(internalSnidRouter, newSnidRouter(config.ServiceSNID))
}
