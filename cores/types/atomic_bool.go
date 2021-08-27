package types

import "sync/atomic"

// AtomicBool 线程安全bool
type AtomicBool int32

// Set 存储
func (a *AtomicBool) Set(value bool) bool {
	if value {
		return atomic.SwapInt32((*int32)(a), 1) == 0
	}

	return atomic.SwapInt32((*int32)(a), 0) == 1
}

// Get 获取
func (a *AtomicBool) Get() bool {
	return atomic.LoadInt32((*int32)(a)) != 0
}
