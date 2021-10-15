package snid

import (
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/kits/snid/proto"
)

var (
	r = router.New()
)

// RouterConfig 路由配置
type RouterConfig struct {
	Snowflake snidConfig `alias:"snowflake"`
}

// InitRouter init
func InitRouter(config RouterConfig) {

	// 注册处理请求
	r.Handle(proto.SnowflakeCommand, newSnid(config.Snowflake))
	r.HandleFunc(proto.AutoincrementCommand, autoincrementID)
}
