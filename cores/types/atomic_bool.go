package types

import "sync/atomic"

// AtomicBool 线程安全bool
type AtomicBool int32

// Set 存储
func (a *AtomicBool) Set(value bool) {
	var i int32
	if value {
		i = 1
	}

	atomic.StoreInt32((*int32)(a), i)
}

// Get 获取
func (a *AtomicBool) Get() bool {
	return atomic.LoadInt32((*int32)(a)) != 0
}
