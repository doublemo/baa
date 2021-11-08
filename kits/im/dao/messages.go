package dao

import (
	"time"
)

const defaultChatMessageKey = "inbox"

// Messages聊天信息
type Messages struct {
	ID          uint64 `gorm:"<-:create;primaryKey"`
	SeqId       uint64
	TSeqId      uint64
	FSeqId      uint64
	To          uint64
	From        uint64
	Content     string
	Group       int32
	ContentType string
	Topic       uint64
	Status      int32
	CreatedAt   time.Time
}
