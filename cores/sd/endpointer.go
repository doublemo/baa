package sd

import (
	"sync"
	"time"
)

type (
	// Endpointer 节点
	Endpointer interface {
		Endpoints() ([]Endpoint, error)
		Register(chan<- struct{})
		Deregister(chan<- struct{})
	}

	// EndpointerLocal 本地节点
	EndpointerLocal struct {
		cache     *endpointCache
		instancer Instancer
		ch        chan Event
		registry  map[chan<- struct{}]bool
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

func (e *EndpointerLocal) Register(ch chan<- struct{}) {
	e.mutx.Lock()
	e.registry[ch] = true
	e.mutx.Unlock()
}
func (e *EndpointerLocal) Deregister(ch chan<- struct{}) {
	e.mutx.Lock()
	delete(e.registry, ch)
	e.mutx.Unlock()
}

func (e *EndpointerLocal) receive() {
	for event := range e.ch {
		// todo update cache
		e.cache.Update(event)
		if event.Err != nil {
			continue
		}

		e.mutx.RLock()
		for ch := range e.registry {
			e.mutx.RUnlock()
			select {
			case ch <- struct{}{}:
			default:
			}
			e.mutx.RLock()
		}
		e.mutx.RUnlock()
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
		registry:  make(map[chan<- struct{}]bool),
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
