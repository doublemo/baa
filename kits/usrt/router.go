package usrt

import (
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/kits/usrt/proto"
)

const (
	internalSnidRouter = 10001
)

var (
	r  = router.New()
	nr = router.New()
)

// RouterConfig 路由配置
type RouterConfig struct {
}

// InitRouter init
func InitRouter() {
	// Register grpc load balance
	//resolver.Register(coressd.NewResolverBuilder(config.ServiceSNID.Name, config.ServiceSNID.Group, sd.Endpointer()))

	// 注册处理请求
	r.HandleFunc(proto.UpdateUserStatusCommand, updateUserStatus)
	r.HandleFunc(proto.DeleteUserStatusCommand, deleteUserStatus)
	r.HandleFunc(proto.GetUserStatusCommand, getUserStatus)

	// 订阅处理
	nr.HandleFunc(proto.DeleteUserStatusCommand, deleteUserStatus)
}
