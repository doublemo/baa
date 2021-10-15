package dao

import (
	"strings"

	"github.com/doublemo/baa/internal/conf"
	"github.com/go-redis/redis/v8"
)

var (
	rdb       redis.UniversalClient
	rdbPrefix string
)

// Open 打开数据库
func Open(rc conf.Redis) error {
	var err error
	// 连接redis
	rdbPrefix = rc.Prefix
	rdb, err = rc.Connect()
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
