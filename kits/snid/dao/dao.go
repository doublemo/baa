package dao

import (
	"context"
	"strings"
	"time"

	"github.com/doublemo/baa/internal/conf"
	"github.com/go-redis/redis/v8"
)

const (
	defaultAutoincrementKey = "autoincrementid"
)

var (
	db       redis.UniversalClient
	dbPrefix string
)

// OpenDB redis
func OpenDB(c conf.Redis) (err error) {
	dbPrefix = c.Prefix
	db, err = c.Connect()
	return
}

// DB 获取leveldb
func DB() redis.UniversalClient {
	return db
}

// RDBNamer 创建redis key
func RDBNamer(name ...string) string {
	prefix := dbPrefix
	if prefix != "" {
		prefix += ":"
	}

	return prefix + strings.Join(name, ":")
}

func AutoincrementID(key string, num int64) ([]uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if num < 1 {
		num = 1
	}

	ret := db.IncrBy(ctx, RDBNamer(defaultAutoincrementKey, key), num)
	err := ret.Err()
	if err != nil {
		return nil, err
	}

	if num == 1 {
		return []uint64{uint64(ret.Val())}, nil
	}

	last := uint64(ret.Val())
	min := last - uint64(num)
	values := make([]uint64, num)
	j := 0
	for i := min; i < last; i++ {
		values[j] = i + 1
		j++
	}

	return values, nil
}
