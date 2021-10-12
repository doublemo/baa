package router

import (
	"sync"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
)

type (
	// HandlerCoresPB 路由处理接口
	HandlerCoresPB interface {
		// Serve 处理请求
		Serve(*corespb.Request) (*corespb.Response, error)
	}

	// HandlerFuncCoresPB 路由处理函数
	HandlerFuncCoresPB func(*corespb.Request) (*corespb.Response, error)

	// RouterCoresPB 路由
	RouterCoresPB struct {
		m     map[coresproto.Command]HandlerCoresPB
		mutex sync.RWMutex
	}
)

// Serve calls f(w, r).
func (f HandlerFuncCoresPB) Serve(req *corespb.Request) (*corespb.Response, error) {
	return f(req)
}

// Handle 注册接口方式
func (r *RouterCoresPB) Handle(pattern coresproto.Command, handler HandlerCoresPB) {
	r.mutex.Lock()
	r.m[pattern] = handler
	r.mutex.Unlock()
}

// HandleFunc 注册函数方式
func (r *RouterCoresPB) HandleFunc(pattern coresproto.Command, handler func(*corespb.Request) (*corespb.Response, error)) {
	r.Handle(pattern, HandlerFuncCoresPB(handler))
}

// Handler 处理
func (r *RouterCoresPB) Handler(req *corespb.Request) (*corespb.Response, error) {
	r.mutex.RLock()
	route, ok := r.m[coresproto.Command(req.Command)]
	r.mutex.RUnlock()

	if !ok {
		return nil, ErrNotFoundRouter
	}
	return route.Serve(req)
}

// NewCoresPB 创建路由
func NewCoresPB() *RouterCoresPB {
	return &RouterCoresPB{
		m: make(map[coresproto.Command]HandlerCoresPB),
	}
}
