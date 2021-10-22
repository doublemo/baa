package cache

import (
	"github.com/doublemo/baa/cores/cache/ringcacher"
)

// CacherConfig 缓存配置
type CacherConfig struct {

	// SnowflakeQueueSize  雪花ID缓存列表大小
	SnowflakeQueueSize int `alias:"snowflakeQueueSize" default:"1000"`

	// SnowflakeMaxQueueNumber 雪花缓存列表最大数量
	SnowflakeMaxQueueNumber int `alias:"snowflakeMaxQueueNumber" default:"2"`

	// SnowflakeMaxWorkers 雪花异步获取最大工作池
	SnowflakeMaxWorkers int `alias:"snowflakeMaxWorkers" default:"2"`

	// MaxBuffer 读取缓冲区大小
	MaxBuffer int `alias:"maxBuffer" default:"128"`
}

// Init 初始化
func Init(c CacherConfig) {
	if c.SnowflakeMaxQueueNumber < 2 {
		c.SnowflakeMaxQueueNumber = 2
	}

	snowflakeCacher = ringcacher.NewUint64Cacher(c.SnowflakeQueueSize, c.SnowflakeMaxQueueNumber, c.SnowflakeMaxWorkers, c.MaxBuffer)
	snowflakeCacher.Start()
}
