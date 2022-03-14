package user

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
	nrRouter  = router.NewMux()
	muxRouter = router.NewMux()
)

// RouterConfig 路由配置
type RouterConfig struct {
	ServiceAuth conf.RPCClient `alias:"auth"`
	ServiceSNID conf.RPCClient `alias:"snid"`
	ServiceIM   conf.RPCClient `alias:"im"`
	ServiceSM   conf.RPCClient `alias:"sm"`
	User        UserConfig     `alias:"usersettings"`
	Group       GroupConfig    `alias:"groupsettings"`
}

// InitRouter init
func InitRouter(config RouterConfig) {
	// Register grpc load balance
	resolver.Register(coressd.NewResolverBuilder(config.ServiceAuth.Name, config.ServiceAuth.Group, sd.Endpointer()))
	resolver.Register(coressd.NewResolverBuilder(config.ServiceSNID.Name, config.ServiceSNID.Group, sd.Endpointer()))
	resolver.Register(coressd.NewResolverBuilder(config.ServiceIM.Name, config.ServiceIM.Group, sd.Endpointer()))
	resolver.Register(coressd.NewResolverBuilder(config.ServiceSM.Name, config.ServiceSM.Group, sd.Endpointer()))

	// 注册处理请求
	r.HandleFunc(command.UserContacts, func(req *corespb.Request) (*corespb.Response, error) { return contact(req, config.User) })
	r.HandleFunc(command.UserContactsList, func(req *corespb.Request) (*corespb.Response, error) { return contacts(req, config.User) })
	r.HandleFunc(command.UserContactsRequest, func(req *corespb.Request) (*corespb.Response, error) { return friendRequestList(req, config.User) })
	r.HandleFunc(command.UserRegister, func(req *corespb.Request) (*corespb.Response, error) { return register(req, config.User) })
	r.HandleFunc(command.UserCheckIsMyFriend, checkIsMyFriend)
	r.HandleFunc(command.UserCheckInGroup, checkInGroup)
	r.HandleFunc(command.UserGroupMembers, func(req *corespb.Request) (*corespb.Response, error) { return groupMembers(req, config.Group) })
	r.HandleFunc(command.UserGroupMembersValidID, func(req *corespb.Request) (*corespb.Response, error) { return groupMembersID(req, config.Group) })
	r.HandleFunc(command.UserInfo, func(req *corespb.Request) (*corespb.Response, error) { return getUserInfo(req, config.User) })
	r.HandleFunc(command.UserCreateGroup, func(req *corespb.Request) (*corespb.Response, error) { return groupCreate(req, config.Group) })

	// 内部调用
	muxRouter.Register(kit.Auth.Int32(), router.New()).Handle(command.AuthAccountInfo, router.NewCall(config.ServiceAuth))
	muxRouter.Register(kit.SNID.Int32(), router.New()).Handle(command.SNIDSnowflake, router.NewCall(config.ServiceSNID))
	muxRouter.Register(kit.IM.Int32(), router.New()).Handle(command.IMSend, router.NewCall(config.ServiceIM))
	muxRouter.Register(kit.SM.Int32(), router.New()).Handle(command.SMUserServers, router.NewCall(config.ServiceSM))

	// 订阅处理
	// nrRouter.Register(kit.USRT.Int32(), router.New()).HandleFunc(command.USRTDeleteUserStatus, deleteUserStatus)

}
