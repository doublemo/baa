package snid

import (
	"github.com/doublemo/baa/cores/uid"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/kits/snid/proto"
)

var (
	r        = router.New()
	nrRouter = router.NewMux()
)

// RouterConfig 路由配置
type RouterConfig struct {
	Snowflake uid.SnowflakeConfig `alias:"snowflake"`
}

// InitRouter init
func InitRouter(config RouterConfig) {

	// 注册处理请求
	r.Handle(proto.SnowflakeCommand, newSnHandler(config.Snowflake))
	r.HandleFunc(proto.AutoincrementCommand, autoincrementID)
	r.HandleFunc(proto.MoreAutoincrementCommand, moreAutoincrementID)

	// 订阅请求
	nrRouter.Register(ServiceName, router.New()).HandleFunc(proto.ClearAutoincrementCommand, clearAutoincrementID)
}
