package router

import (
	"errors"
	"sync"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
)

var (
	// ErrNotFoundRouter 路由不存在
	ErrNotFoundRouter = errors.New("ErrNotFoundRouter")
)

type (
	// Handler 路由处理接口
	Handler interface {
		// Serve 处理请求
		Serve(*corespb.Request) (*corespb.Response, error)
	}

	// HandlerFunc 路由处理函数
	HandlerFunc func(*corespb.Request) (*corespb.Response, error)

	// Router 路由
	Router struct {
		m     map[coresproto.Command]Handler
		mutex sync.RWMutex
	}
)

// Serve calls f(w, r).
func (f HandlerFunc) Serve(req *corespb.Request) (*corespb.Response, error) {
	return f(req)
}

// Handle 注册接口方式
func (r *Router) Handle(pattern coresproto.Command, handler Handler) {
	r.mutex.Lock()
	r.m[pattern] = handler
	r.mutex.Unlock()
}

// HandleFunc 注册函数方式
func (r *Router) HandleFunc(pattern coresproto.Command, handler func(*corespb.Request) (*corespb.Response, error)) {
	r.Handle(pattern, HandlerFunc(handler))
}

// Handler 处理
func (r *Router) Handler(req *corespb.Request) (*corespb.Response, error) {
	r.mutex.RLock()
	route, ok := r.m[coresproto.Command(req.Command)]
	r.mutex.RUnlock()

	if !ok {
		return nil, ErrNotFoundRouter
	}
	return route.Serve(req)
}

// New 创建路由
func New() *Router {
	return &Router{
		m: make(map[coresproto.Command]Handler),
	}
}