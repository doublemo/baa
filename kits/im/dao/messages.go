package dao

import "time"

// Messages聊天信息
type Messages struct {
	MsgId      uint64 `gorm:"<-:create;primaryKey"`
	MsgSeqId   uint64 `gorm:"<-:create;index()"`
	MsgTo      string
	MsgFrom    string
	MsgContent string
	MsgType    string
	MsgOrder   int64
	CreatedAt  time.Time
}
