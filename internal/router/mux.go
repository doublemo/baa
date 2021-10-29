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
	routes map[int32]*Router
	mutex  sync.RWMutex
}

// Register 注册路由
func (m *Mux) Register(id int32, r *Router) *Router {
	m.mutex.Lock()
	m.routes[id] = r
	m.mutex.Unlock()
	return r
}

// Handle 向指定路由中注册处理
func (m *Mux) Handle(id int32, pattern coresproto.Command, handler Handler) error {
	m.mutex.RLock()
	r, ok := m.routes[id]
	m.mutex.RUnlock()

	if !ok {
		return ErrNotFoundRouterName
	}

	r.Handle(pattern, handler)
	return nil
}

// HandleFunc 向指定路由中注册处理
func (m *Mux) HandleFunc(id int32, pattern coresproto.Command, handler func(*corespb.Request) (*corespb.Response, error)) error {
	m.mutex.RLock()
	r, ok := m.routes[id]
	m.mutex.RUnlock()

	if !ok {
		return ErrNotFoundRouterName
	}

	r.HandleFunc(pattern, handler)
	return nil
}

// Handler 处理指定，路由中的信息
func (m *Mux) Handler(id int32, req *corespb.Request) (*corespb.Response, error) {
	m.mutex.RLock()
	r, ok := m.routes[id]
	m.mutex.RUnlock()

	if !ok {
		return nil, ErrNotFoundRouterName
	}

	return r.Handler(req)
}

// NewMux 创建多路由
func NewMux() *Mux {
	return &Mux{routes: make(map[int32]*Router)}
}
