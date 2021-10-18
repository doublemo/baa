package router

import (
	"errors"
	"sync"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
)

var (
	// ErrNotFoundRouterName 路由名称不存在
	ErrNotFoundRouterName = errors.New("ErrNotFoundRouterName")
)

// Mux 多路由管理
type Mux struct {
	routes map[string]*Router
	mutex  sync.RWMutex
}

// Register 注册路由
func (m *Mux) Register(name string, r *Router) {
	m.mutex.Lock()
	m.routes[name] = r
	m.mutex.Unlock()
}

// Handle 向指定路由中注册处理
func (m *Mux) Handle(name string, pattern coresproto.Command, handler Handler) error {
	m.mutex.RLock()
	r, ok := m.routes[name]
	m.mutex.RUnlock()

	if !ok {
		return ErrNotFoundRouterName
	}

	r.Handle(pattern, handler)
	return nil
}

// HandleFunc 向指定路由中注册处理
func (m *Mux) HandleFunc(name string, pattern coresproto.Command, handler func(*corespb.Request) (*corespb.Response, error)) error {
	m.mutex.RLock()
	r, ok := m.routes[name]
	m.mutex.RUnlock()

	if !ok {
		return ErrNotFoundRouterName
	}

	r.HandleFunc(pattern, handler)
	return nil
}

// Handler 处理指定，路由中的信息
func (m *Mux) Handler(name string, req *corespb.Request) (*corespb.Response, error) {
	m.mutex.RLock()
	r, ok := m.routes[name]
	m.mutex.RUnlock()

	if !ok {
		return nil, ErrNotFoundRouterName
	}

	return r.Handler(req)
}

// NewMux 创建多路由
func NewMux() *Mux {
	return &Mux{routes: make(map[string]*Router)}
}
