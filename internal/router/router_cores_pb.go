package router

import (
	"errors"
	"strings"
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

		// ServeHTTP 处理网关转发来的HTTP请求
		ServeHTTP(*corespb.Request) (*corespb.Response, error)
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

func (f HandlerFunc) ServeHTTP(req *corespb.Request) (*corespb.Response, error) {
	return f(req)
}

// Handle 注册接口方式
func (r *Router) Handle(pattern coresproto.Command, handler Handler) *Router {
	r.mutex.Lock()
	r.m[pattern] = handler
	r.mutex.Unlock()
	return r
}

// HandleFunc 注册函数方式
func (r *Router) HandleFunc(pattern coresproto.Command, handler func(*corespb.Request) (*corespb.Response, error)) *Router {
	r.Handle(pattern, HandlerFunc(handler))
	return r
}

// Handler 处理
func (r *Router) Handler(req *corespb.Request) (*corespb.Response, error) {
	r.mutex.RLock()
	route, ok := r.m[coresproto.Command(req.Command)]
	r.mutex.RUnlock()

	if !ok {
		return nil, ErrNotFoundRouter
	}

	if IsHTTP(req) {
		return route.ServeHTTP(req)
	}

	return route.Serve(req)
}

// New 创建路由
func New() *Router {
	return &Router{
		m: make(map[coresproto.Command]Handler),
	}
}

// IsHTTP 检查是否为网关转的http请求
func IsHTTP(req *corespb.Request) bool {
	if m, ok := req.Header["Content-Type"]; ok && strings.ToLower(m) == "json" {
		return true
	}

	return false
}
