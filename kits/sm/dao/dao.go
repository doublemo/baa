package dao

import (
	"errors"
	"strings"
	"time"

	"github.com/doublemo/baa/cores/cache/memcacher"
	"github.com/doublemo/baa/internal/conf"
	"github.com/go-redis/redis/v8"
)

const (
	// 设置缓存过期时间
	defaultCacheExpiration = 5

	// 设置缓存自动回收时间
	defaultCacheCleanupInterval = 10
)

var (
	rdb       redis.UniversalClient
	rdbPrefix string
	cacher    = memcacher.New(defaultCacheExpiration*time.Minute, defaultCacheCleanupInterval*time.Minute)
)

var (
	ErrRecordIsFound  = errors.New("RecordIsFound")
	ErrRecordNotFound = errors.New("RecordNotFound")
)

// CacherConfig 缓存配置
type CacherConfig struct {
	// 设置缓存过期时间
	Expiration int `alias:"expiration"`

	// 设置缓存自动回收时间
	CleanupInterval int `alias:"cleanupInterval"`
}

// Open 打开数据库
func Open(rc conf.Redis, cc CacherConfig) error {
	var err error
	// 连接redis
	rdbPrefix = rc.Prefix
	rdb, err = rc.Connect()

	if cc.Expiration != 0 || cc.CleanupInterval != 0 {
		cacher = memcacher.New(time.Duration(cc.Expiration)*time.Minute, time.Duration(cc.CleanupInterval)*time.Minute)
	}

	return err
}

// RDB 获取redis数据库
func RDB() redis.UniversalClient {
	return rdb
}

// RDBNamer 创建redis key
func RDBNamer(name ...string) string {
	prefix := rdbPrefix
	if prefix != "" {
		prefix += ":"
	}
	return prefix + strings.Join(name, ":")
}
