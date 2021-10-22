package cache

import (
	"context"

	"github.com/doublemo/baa/cores/cache/qcacher"
)

var (
	snowflakeCacher *qcacher.Uint64Cacher
)

// GetSnowflakeID 获取ID
func GetSnowflakeID(ctx context.Context) (uint64, error) {
	return snowflakeCacher.Pop(ctx)
}

func SnowflakeCacherOnFill(fn func(int) ([]uint64, error)) {
	snowflakeCacher.OnFill(fn)
}
