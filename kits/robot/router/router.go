package router

import (
	"errors"
	"sync"

	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/kits/robot/session"
)

var (
	// ErrNotFoundRouter 路由不存在
	ErrNotFoundRouter = errors.New("ErrNotFoundRouter")
)

type (
	// Handler 路由处理接口
	Handler interface {
		// Serve 处理请求
		Serve(session.Peer, coresproto.Response) error
	}

	// HandlerFunc 路由处理函数
	HandlerFunc func(session.Peer, coresproto.Response) error

	// Router 路由
	Router struct {
		m     map[coresproto.Command]map[coresproto.Command]Handler
		mutex sync.RWMutex
	}
)

// Serve calls f(w, r).
func (f HandlerFunc) Serve(peer session.Peer, resp coresproto.Response) error {
	return f(peer, resp)
}

// Destroy 销毁和Peer相关的信息
func (f HandlerFunc) Destroy(peer session.Peer) {
	// clear
}

// Handle 注册接口方式
func (r *Router) Handle(pattern1, pattern2 coresproto.Command, handler Handler) {
	r.mutex.Lock()
	if _, ok := r.m[pattern1]; !ok {
		r.m[pattern1] = make(map[coresproto.Command]Handler)
	}
	r.m[pattern1][pattern2] = handler
	r.mutex.Unlock()
}

// HandleFunc 注册函数方式
func (r *Router) HandleFunc(pattern1, pattern2 coresproto.Command, handler func(session.Peer, coresproto.Response) error) {
	r.Handle(pattern1, pattern2, HandlerFunc(handler))
}

// Handler 处理
func (r *Router) Handler(peer session.Peer, resp coresproto.Response) error {
	var (
		handler Handler
		exist   bool
	)

	r.mutex.RLock()
	if m, ok := r.m[resp.Command()]; ok && m != nil {
		handler, exist = m[resp.SubCommand()]
	}
	r.mutex.RUnlock()

	if !exist || handler == nil {
		return ErrNotFoundRouter
	}

	return handler.Serve(peer, resp)
}

// New 创建路由
func New() *Router {
	return &Router{
		m: make(map[coresproto.Command]map[coresproto.Command]Handler),
	}
}
