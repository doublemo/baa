package dao

import (
	"context"
	"fmt"
	"time"
)

const defaultChatMessageKey = "inbox"

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

func SaveMsgtoRedis(topic string, values []byte) error {
	namer := RDBNamer(defaultChatMessageKey, topic)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ret := rdb.LPush(ctx, namer, values)
	if ret.Val() > 100 {
		rdb.LTrim(ctx, topic, 1, 100)
	}

	fmt.Println("ret.Val()->", ret.Val())
	return ret.Err()
}

func GetMsgFromRedis(topic string) ([]string, error) {
	namer := RDBNamer(defaultChatMessageKey, topic)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ret := rdb.LRange(ctx, namer, -10, 10)
	return ret.Val(), ret.Err()
}
