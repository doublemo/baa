package router

import (
	"errors"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/sfu/session"
)

var (
	// ErrNotFound 路由不存在
	ErrNotFound = errors.New("ErrNotFound")
)

// Callback 配置器调用函数类型
type Callback func(session.Peer, *corespb.Request) (*corespb.Response, error)

var routes = make(map[coresproto.Command]Callback)

// On 注册路由
func On(code coresproto.Command, fn Callback) {
	routes[code] = fn
}

// Fn 获取路由
func Fn(code coresproto.Command) (Callback, error) {
	if fn, ok := routes[code]; ok {
		return fn, nil
	}
	return nil, ErrNotFound
}
