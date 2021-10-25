package cache

import (
	"context"
	"errors"
	"sync"

	"github.com/doublemo/baa/cores/cache/ringcacher"
)

var (
	uidCacher sync.Map
	uidConfig CacherConfig

	// ErrNotfound 不存在
	ErrNotfound = errors.New("ErrNotfound")
)

// GetUID 从缓存中获取自增ID
func GetUID(ctx context.Context, k string, num int) ([]uint64, error) {
	var r *ringcacher.Uint64Cacher
	m, ok := uidCacher.Load(k)
	if !ok || m == nil {
		r = uidConfig.UIDNew(k)
		uidCacher.Store(k, r)
	} else {
		m0, ok := m.(*ringcacher.Uint64Cacher)
		if !ok {
			return nil, ErrNotfound
		}
		r = m0
	}

	data := make([]uint64, num)
	for i := 0; i < num; i++ {
		v, err := r.Pop(ctx)
		if err != nil {
			return nil, err
		}

		data[i] = v
	}

	return data, nil
}

// RemoveUID 删除自增ID缓存
func RemoveUID(k string) {
	uidCacher.Delete(k)
}
