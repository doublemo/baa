package sd

import (
	"sync"
	"time"
)

type (
	// Endpointer 节点
	Endpointer interface {
		Endpoints() ([]Endpoint, error)
		Register(string, chan<- struct{})
		Deregister(string)
	}

	// EndpointerLocal 本地节点
	EndpointerLocal struct {
		cache     *endpointCache
		instancer Instancer
		ch        chan Event
		registry  map[string]chan<- struct{}
		mutx      sync.RWMutex
	}

	// EndpointerOption allows control of endpointCache behavior.
	EndpointerOption func(*endpointerOptions)

	endpointerOptions struct {
		invalidateOnError bool
		invalidateTimeout time.Duration
	}
)

// Endpoints 获取节点
func (e *EndpointerLocal) Endpoints() ([]Endpoint, error) {
	return e.cache.Endpoints()
}

func (e *EndpointerLocal) Register(name string, ch chan<- struct{}) {
	e.mutx.Lock()
	e.registry[name] = ch
	e.mutx.Unlock()
}
func (e *EndpointerLocal) Deregister(name string) {
	e.mutx.Lock()
	delete(e.registry, name)
	e.mutx.Unlock()
}

func (e *EndpointerLocal) receive() {
	for event := range e.ch {

		// todo update cache
		e.cache.Update(event)

		if event.Err != nil {
			continue
		}

		// 如果没有错误那么广播
		for _, instance := range event.Instances {
			endpoint := &EndpointLocal{}
			if err := endpoint.Unmarshal(instance); err == nil {
				e.mutx.RLock()
				m, ok := e.registry[endpoint.Name()]
				e.mutx.RUnlock()
				if ok {
					select {
					case m <- struct{}{}:
					default:
					}
				}
			}
		}
	}
}

// Close deregisters DefaultEndpointer from the Instancer and stops the internal go-routine.
func (e *EndpointerLocal) Close() {
	e.instancer.Deregister(e.ch)
	close(e.ch)
}

// NewEndpointer 创建节点
func NewEndpointer(src Instancer, options ...EndpointerOption) *EndpointerLocal {
	opts := endpointerOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	el := &EndpointerLocal{
		cache:     newEndpointCache(opts),
		instancer: src,
		ch:        make(chan Event),
		registry:  make(map[string]chan<- struct{}),
	}

	go el.receive()
	src.Register(el.ch)
	return el
}

// InvalidateOnError returns EndpointerOption that controls how the Endpointer
// behaves when then Instancer publishes an Event containing an error.
// Without this option the Endpointer continues returning the last known
// endpoints. With this option, the Endpointer continues returning the last
// known endpoints until the timeout elapses, then closes all active endpoints
// and starts returning an error. Once the Instancer sends a new update with
// valid resource instances, the normal operation is resumed.
func InvalidateOnError(timeout time.Duration) EndpointerOption {
	return func(opts *endpointerOptions) {
		opts.invalidateOnError = true
		opts.invalidateTimeout = timeout
	}
}
