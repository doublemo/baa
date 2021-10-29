package cache

import (
	"time"

	"github.com/doublemo/baa/cores/cache/memcacher"
)

var (
	// 设置缓存过期时间
	defaultCacheExpiration = 5

	// 设置缓存自动回收时间
	defaultCacheCleanupInterval = 10

	cacher *memcacher.Cache
)

// CacherConfig 缓存配置
type CacherConfig struct {
	// MemCacheExpiration 一般数据缓存 缓存过期时间
	MemCacheExpiration int `alias:"memCacheExpiration" default:"2"`

	// MemCacheCleanupInterval 一般数据缓存 缓存自动回收时间
	MemCacheCleanupInterval int `alias:"memCacheCleanupInterval" default:"2"`
}

// Init 初始化
func Init(c CacherConfig) {
	if c.MemCacheExpiration < 1 {
		c.MemCacheExpiration = defaultCacheExpiration
	}

	if c.MemCacheCleanupInterval < 1 {
		c.MemCacheCleanupInterval = defaultCacheCleanupInterval
	}
	cacher = memcacher.New(time.Duration(c.MemCacheExpiration)*time.Minute, time.Duration(c.MemCacheCleanupInterval)*time.Minute)
}

// Get 从缓存中获取数据
func Get(k string) (interface{}, bool) {
	return cacher.Get(k)
}

// Set 设置缓存
func Set(k string, value interface{}, d time.Duration) {
	cacher.Set(k, value, d)
}

// Remove 从缓存中删除数据
func Remove(k string) {
	cacher.Delete(k)
	cacher.Flush()
}

// Add 添加缓存
func Add(k string, value interface{}, d time.Duration) error {
	return cacher.Add(k, value, d)
}

// Increment 增加
func Increment(k string, n int64) error {
	return cacher.Increment(k, n)
}

// Decrement 减少
func Decrement(k string, n int64) error {
	return cacher.Decrement(k, n)
}
