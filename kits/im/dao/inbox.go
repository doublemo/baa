package dao

import (
	"context"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const (
	defaultInboxKey        = "inbox"
	defaultInboxMessageKey = "messages"
)

// WriteInboxC 写到个人信息箱
func WriteInboxC(ctx context.Context, msg *Messages) error {
	tinboxNamer := RDBNamer(defaultInboxKey, strconv.FormatUint(msg.To, 10))
	finboxNamer := RDBNamer(defaultInboxKey, strconv.FormatUint(msg.From, 10))
	messageNamer := RDBNamer(defaultInboxMessageKey, strconv.FormatUint(msg.ID, 10))

	message := make(map[string]interface{})
	message["ID"] = msg.ID
	message["SeqId"] = msg.SeqId
	message["To"] = msg.To
	message["From"] = msg.From
	message["Content"] = msg.Content
	message["Group"] = msg.Group
	message["ContentType"] = msg.ContentType
	message["CreatedAt"] = msg.CreatedAt
	message["TSeqId"] = msg.TSeqId
	message["FSeqId"] = msg.FSeqId
	message["Status"] = msg.Status
	message["Topic"] = msg.Topic

	var (
		retInbox   []*redis.IntCmd
		retMessage *redis.BoolCmd
	)

	tvalue := strconv.FormatUint(msg.TSeqId, 10) + ":" + strconv.FormatUint(msg.ID, 10)
	fvalue := strconv.FormatUint(msg.FSeqId, 10) + ":" + strconv.FormatUint(msg.ID, 10)
	_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		retInbox = append(retInbox, pipe.LPush(ctx, tinboxNamer, tvalue))
		retInbox = append(retInbox, pipe.LPush(ctx, finboxNamer, fvalue))
		retMessage = pipe.HMSet(ctx, messageNamer, message)
		pipe.LTrim(ctx, tinboxNamer, 0, 10000)
		pipe.LTrim(ctx, finboxNamer, 0, 10000)
		return nil
	})

	if err != nil {
		return err
	}

	for _, ret := range retInbox {
		err = ret.Err()
		if err != nil {
			return err
		}
	}
	return retMessage.Err()
}

// WriteInboxG 写到群信息箱
func WriteInboxG(ctx context.Context, msg *Messages) error {
	tinboxNamer := RDBNamer(defaultInboxKey, strconv.FormatUint(msg.To, 10))
	messageNamer := RDBNamer(defaultInboxMessageKey, strconv.FormatUint(msg.ID, 10))

	message := make(map[string]interface{})
	message["ID"] = msg.ID
	message["SeqId"] = msg.SeqId
	message["To"] = msg.To
	message["From"] = msg.From
	message["Content"] = msg.Content
	message["Group"] = msg.Group
	message["ContentType"] = msg.ContentType
	message["CreatedAt"] = msg.CreatedAt
	message["TSeqId"] = msg.TSeqId
	message["FSeqId"] = msg.FSeqId
	message["Status"] = msg.Status
	message["Topic"] = msg.Topic

	var (
		retInbox   []*redis.IntCmd
		retMessage *redis.BoolCmd
	)

	tvalue := strconv.FormatUint(msg.TSeqId, 10) + ":" + strconv.FormatUint(msg.ID, 10)
	_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		retInbox = append(retInbox, pipe.LPush(ctx, tinboxNamer, tvalue))
		retMessage = pipe.HMSet(ctx, messageNamer, message)
		pipe.LTrim(ctx, tinboxNamer, 0, 10000)
		return nil
	})

	if err != nil {
		return err
	}

	for _, ret := range retInbox {
		err = ret.Err()
		if err != nil {
			return err
		}
	}
	return retMessage.Err()
}

// GetInboxMesssage 获取邮件
func GetInboxMesssage(id uint64) (*Messages, error) {
	return nil, nil
}

// ChangeInboxMessageStatus 修改信息状态
func ChangeInboxMessageStatus(ctx context.Context, id uint64, status int32) (bool, error) {
	messageNamer := RDBNamer(defaultInboxMessageKey, strconv.FormatUint(id, 10))
	retMessage := rdb.HMSet(ctx, messageNamer, map[string]interface{}{"Status": status})
	return retMessage.Val(), retMessage.Err()
}
