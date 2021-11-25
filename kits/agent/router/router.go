package router

import (
	"errors"
	"sync"

	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/kits/agent/session"
)

var (
	// ErrNotFoundRouter 路由不存在
	ErrNotFoundRouter = errors.New("ErrNotFoundRouter")

	// ErrNotSupportCommand 不支持的命令
	ErrNotSupportCommand = errors.New("ErrNotSupportCommand")
)

type (
	// Handler 路由处理接口
	Handler interface {
		// Serve 处理请求
		Serve(session.Peer, coresproto.Request) (coresproto.Response, error)

		// Destroy 销毁和Peer相关的信息
		Destroy(session.Peer) error
	}

	// HandlerFunc 路由处理函数
	HandlerFunc func(session.Peer, coresproto.Request) (coresproto.Response, error)

	// Router 路由
	Router struct {
		m     map[coresproto.Command]Handler
		mutex sync.RWMutex
	}
)

// Serve calls f(w, r).
func (f HandlerFunc) Serve(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	return f(peer, req)
}

// Destroy 销毁和Peer相关的信息
func (f HandlerFunc) Destroy(peer session.Peer) error {
	return nil
}

// Handle 注册接口方式
func (r *Router) Handle(pattern coresproto.Command, handler Handler) {
	r.mutex.Lock()
	r.m[pattern] = handler
	r.mutex.Unlock()
}

// HandleFunc 注册函数方式
func (r *Router) HandleFunc(pattern coresproto.Command, handler func(session.Peer, coresproto.Request) (coresproto.Response, error)) {
	r.Handle(pattern, HandlerFunc(handler))
}

// Handler 处理
func (r *Router) Handler(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	r.mutex.RLock()
	route, ok := r.m[req.Command()]
	r.mutex.RUnlock()

	if !ok {
		return nil, ErrNotFoundRouter
	}

	return route.Serve(peer, req)
}

// Destroy 销毁和Peer相关的信息
func (r *Router) Destroy(peer session.Peer) {
	r.mutex.RLock()
	for _, route := range r.m {
		r.mutex.RUnlock()
		route.Destroy(peer)
		r.mutex.RLock()
	}
	r.mutex.RUnlock()
}

// New 创建路由
func New() *Router {
	return &Router{
		m: make(map[coresproto.Command]Handler),
	}
}
