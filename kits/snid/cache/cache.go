package cache

import (
	"github.com/doublemo/baa/cores/cache/ringcacher"
)

// CacherConfig 缓存配置
type CacherConfig struct {

	// AutoUIDQueueSize  自增IDID缓存列表大小
	AutoUIDQueueSize int `alias:"autoUIDQueueSize" default:"100"`

	// AutoUIDMaxQueueNumber 自增ID缓存列表最大数量
	AutoUIDMaxQueueNumber int `alias:"autoUIDMaxQueueNumber" default:"3"`

	// AutoUIDMaxWorkers  自增ID异步获取最大工作池
	AutoUIDMaxWorkers int `alias:"autoUIDMaxWorkers" default:"3"`

	// AutoUIDMaxBuffer 读取缓冲区大小
	AutoUIDMaxBuffer int `alias:"autoUIDMaxBuffer" default:"1"`

	// UIDNew 自增ID创建
	UIDNew func(section string) *ringcacher.Uint64Cacher `alias:"-"`
}

func Init(c CacherConfig) {
	uidConfig = c
}
