package usrt

import (
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/router"
)

const (
	internalSnidRouter = 10001
)

var (
	r        = router.New()
	nrRouter = router.NewMux()
)

// RouterConfig 路由配置
type RouterConfig struct {
}

// InitRouter init
func InitRouter() {
	// Register grpc load balance

	// 注册处理请求
	r.HandleFunc(command.USRTUpdateUserStatus, updateUserStatus)
	r.HandleFunc(command.USRTDeleteUserStatus, deleteUserStatus)
	r.HandleFunc(command.USRTGetUserStatus, getUserStatus)

	// 订阅处理
	nrRouter.Register(ServiceName, router.New()).HandleFunc(command.USRTDeleteUserStatus, deleteUserStatus)
}
