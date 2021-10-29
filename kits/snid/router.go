package snid

import (
	"github.com/doublemo/baa/cores/uid"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/router"
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
	r.Handle(command.SNIDSnowflake, newSnHandler(config.Snowflake))
	r.HandleFunc(command.SNIDAutoincrement, autoincrementID)
	r.HandleFunc(command.SNIDMoreAutoincrement, moreAutoincrementID)

	// 订阅请求
	nrRouter.Register(kit.SNID.Int32(), router.New()).HandleFunc(command.SNIDClearAutoincrement, clearAutoincrementID)
}
