package dao

import (
	"time"
)

const defaultChatMessageKey = "inbox"

// Messages聊天信息
type Messages struct {
	ID          uint64 `gorm:"<-:create;primaryKey"`
	SeqId       uint64 `gorm:"<-:create;index()"`
	To          uint64
	From        uint64
	Content     string
	Group       int32
	ContentType string
	Topic       string
	CreatedAt   time.Time
}
