package command

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/kit"
)

const (
	// AuthLogin 登录
	AuthLogin coresproto.Command = kit.Auth + iota + 1

	// AuthRegister 注册
	AuthRegister

	// AuthLogout 退出登录
	AuthLogout

	// AuthOffline 玩家离线
	AuthOffline

	// AuthAccountInfo 账户信息
	AuthAccountInfo

	// AuthorizedToken 验证token
	AuthorizedToken
)
