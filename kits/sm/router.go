package sm

import (
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/router"
)

var (
	r               = router.New()
	nrRouter        = router.NewMux()
	nInternalRouter = router.New()
)

// RouterConfig 路由配置
type RouterConfig struct {
}

// InitRouter init
func InitRouter() {
	// Register grpc load balance

	// 注册处理请求
	r.HandleFunc(command.SMUserStatus, getUsersStatus)
	r.HandleFunc(command.SMBroadcastMessagesToAgent, broadcastMessagesToAgent)
	r.HandleFunc(command.SMUserServers, getUserServers)
	r.HandleFunc(command.SMAssginServers, userAssignServer)

	// 订阅处理
	nrRouter.Register(kit.SM.Int32(), router.New()).HandleFunc(command.SMEvent, eventHandler)
	nInternalRouter.HandleFunc(command.SMEvent, internalEventHandler)
}
