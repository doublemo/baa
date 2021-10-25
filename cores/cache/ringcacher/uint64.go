package ringcacher

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/doublemo/baa/cores/pool/worker"
	"github.com/doublemo/baa/cores/queue"
)

// Uint64Cacher 64位数字缓存 RingCacher
// 环形缓存当同步时才能保存所有缓存数组队列中的ID是为有序的,否则只能保证单个队列为有序
type Uint64Cacher struct {
	reserveCacher []*queue.OrderedUint64
	cacher        *queue.OrderedUint64
	maxCacheQueue int
	queueSize     int
	fillfn        atomic.Value
	async         bool
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

	s.mutex.Lock()
	reserveCacherLen = len(s.reserveCacher)
	maxCacheQueue := s.maxCacheQueue
	s.mutex.Unlock()

	var id uint64
	if cacheLen >= 1 {
		id = s.cacher.Pop()
	} else {
		if reserveCacherLen > 0 {
			s.mutex.Lock()
			s.cacher = s.reserveCacher[0]
			s.reserveCacher = s.reserveCacher[1:]
			s.mutex.Unlock()
		} else if s.async {
			id = s.readSync()
		}
	}

	if reserveCacherLen < maxCacheQueue {
		m := maxCacheQueue - reserveCacherLen
		s.task(m, s.async)
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

func (s *Uint64Cacher) task(num int, async bool) {
	if async {
		for i := 0; i < num; i++ {
			s.workers.Submit(fill(s))
		}

		return
	}

	s.workers.SubmitWait(func() {
		for i := 0; i < num; i++ {
			fill(s)()
		}
	})

	// sort
	s.mutex.Lock()
	sort.Slice(s.reserveCacher, func(i, j int) bool {
		a := s.reserveCacher[i].Pop()
		b := s.reserveCacher[j].Pop()

		s.reserveCacher[i].Push(a)
		s.reserveCacher[j].Push(b)
		return a < b
	})

	if s.cacher.Len() < 1 {
		s.cacher = s.reserveCacher[0]
		s.reserveCacher = s.reserveCacher[1:]
	}
	s.mutex.Unlock()
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
		s.task(m, s.async)
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

func NewUint64Cacher(cacheQueueSize, maxCacheQueue, maxWorkers, maxBuffer int, async bool) *Uint64Cacher {
	sr := &Uint64Cacher{
		reserveCacher: make([]*queue.OrderedUint64, 0),
		cacher:        queue.NewOrderedUint64(),
		maxCacheQueue: maxCacheQueue,
		queueSize:     cacheQueueSize,
		async:         async,
		requestChan:   make(chan struct{}, maxBuffer),
		replyChan:     make(chan uint64, maxBuffer),
		closeChan:     make(chan struct{}),
		workers:       *worker.New(maxWorkers),
	}
	return sr
}
