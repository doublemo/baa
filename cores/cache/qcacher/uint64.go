package qcacher

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/doublemo/baa/cores/pool/worker"
	"github.com/doublemo/baa/cores/queue"
)

// Uint64Cacher 64位数字缓存
type Uint64Cacher struct {
	reserveCacher []*queue.OrderedUint64
	cacher        *queue.OrderedUint64
	maxCacheQueue int
	queueSize     int
	fillfn        atomic.Value
	requestChan   chan struct{}
	replyChan     chan uint64
	closeChan     chan struct{}
	workers       worker.WorkerPool
	closeOnce     sync.Once
	mutex         sync.Mutex
}

func (s *Uint64Cacher) serve() {
	for {
		select {
		case <-s.requestChan:
			s.replyChan <- s.read()

		case <-s.closeChan:
			return
		}
	}
}

func (s *Uint64Cacher) read() uint64 {
	s.mutex.Lock()
	reserveCacherLen := len(s.reserveCacher)
	s.mutex.Unlock()

	if s.cacher.Len() < 1 && reserveCacherLen < 1 {
		return s.readSync()
	}

	cacheLen := s.cacher.Len()
	rate := int(float32(s.queueSize) * 0.3)
	if rate < 1 {
		rate = 1
	}

	if cacheLen > rate {
		return s.cacher.Pop()
	}

	var id uint64
	if cacheLen >= 1 {
		id = s.cacher.Pop()
	} else {
		s.mutex.Lock()
		s.cacher = s.reserveCacher[0]
		s.reserveCacher = s.reserveCacher[1:]
		s.mutex.Unlock()
	}

	s.mutex.Lock()
	reserveCacherLen = len(s.reserveCacher)
	maxCacheQueue := s.maxCacheQueue
	s.mutex.Unlock()

	if reserveCacherLen < maxCacheQueue {
		m := maxCacheQueue - reserveCacherLen
		for i := 0; i < m; i++ {
			s.workers.Submit(fill(s))
		}
	}

	if id > 0 {
		return id
	}

	id = s.cacher.Pop()
	if id < 1 {
		return s.readSync()
	}
	return id
}

func (s *Uint64Cacher) readSync() uint64 {
	handler, ok := s.fillfn.Load().(func(int) ([]uint64, error))
	if !ok || handler == nil {
		return 0
	}

	data, err := handler(s.queueSize)
	if err != nil {
		return 0
	}

	s.cacher = queue.NewOrderedUint64(data...)
	return s.cacher.Pop()
}

func (s *Uint64Cacher) OnFill(fn func(int) ([]uint64, error)) {
	s.fillfn.Store(fn)

	s.mutex.Lock()
	reserveCacherLen := len(s.reserveCacher)
	maxCacheQueue := s.maxCacheQueue
	s.mutex.Unlock()

	if reserveCacherLen < maxCacheQueue {
		m := maxCacheQueue - reserveCacherLen
		for i := 0; i < m; i++ {
			s.workers.Submit(fill(s))
		}
	}
}

func (s *Uint64Cacher) Pop(ctx context.Context) (uint64, error) {
	select {
	case s.requestChan <- struct{}{}:
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	select {
	case id, ok := <-s.replyChan:
		if !ok {
			return 0, errors.New("chanclosed")
		}

		return id, nil

	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (s *Uint64Cacher) Start() {
	go s.serve()
}

func (s *Uint64Cacher) Close() {
	s.closeOnce.Do(func() {
		close(s.closeChan)
	})
}

func fill(sr *Uint64Cacher) func() {
	return func() {
		handler, ok := sr.fillfn.Load().(func(int) ([]uint64, error))
		if !ok || handler == nil {
			return
		}

		data, err := handler(sr.queueSize)
		if err != nil {
			return
		}

		q := queue.NewOrderedUint64(data...)
		sr.mutex.Lock()
		sr.reserveCacher = append(sr.reserveCacher, q)
		sr.mutex.Unlock()
	}
}

func NewUint64Cacher(cacheQueueSize, maxCacheQueue, maxWorkers, maxBuffer int) *Uint64Cacher {
	sr := &Uint64Cacher{
		reserveCacher: make([]*queue.OrderedUint64, 0),
		cacher:        queue.NewOrderedUint64(),
		maxCacheQueue: maxCacheQueue,
		queueSize:     cacheQueueSize,
		requestChan:   make(chan struct{}, maxBuffer),
		replyChan:     make(chan uint64, maxBuffer),
		closeChan:     make(chan struct{}),
		workers:       *worker.New(maxWorkers),
	}
	return sr
}
