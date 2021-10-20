package cache

import (
	"github.com/doublemo/baa/cores/queue"
)

var snowflakeCacher *queue.OrderedUint64

func init() {
	snowflakeCacher = queue.NewOrderedUint64()
}

// GetSnowflakeID 获取ID
func GetSnowflakeID() uint64 {
	return snowflakeCacher.Pop()
}

// GetSnowflakeLen 缓存队列长度
func GetSnowflakeLen() int {
	return snowflakeCacher.Len()
}

// ResetSnowflakeID 重置
func ResetSnowflakeID(values ...uint64) {
	snowflakeCacher.Push(values...)
}
