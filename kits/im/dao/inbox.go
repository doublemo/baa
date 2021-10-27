package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	defaultInboxKey         = "inbox"
	defaultInboxTimelineKey = "timelines"
)

// WriteInboxC 写到个人信息箱
func WriteInboxC(ctx context.Context, tid, fid uint64, msg *Messages) error {
	tinboxNamer := RDBNamer(defaultInboxKey, strconv.FormatUint(msg.To, 10), strconv.FormatUint(tid, 10))
	ttimelinesNamer := RDBNamer(defaultInboxTimelineKey, strconv.FormatUint(msg.To, 10))
	finboxNamer := RDBNamer(defaultInboxKey, strconv.FormatUint(msg.From, 10), strconv.FormatUint(fid, 10))
	ftimelinesNamer := RDBNamer(defaultInboxTimelineKey, strconv.FormatUint(msg.From, 10))

	inboxMessage := make(map[string]interface{})
	inboxMessage["ID"] = msg.ID
	inboxMessage["SeqId"] = msg.SeqId
	inboxMessage["To"] = msg.To
	inboxMessage["From"] = msg.From
	inboxMessage["Content"] = msg.Content
	inboxMessage["Group"] = msg.Group
	inboxMessage["ContentType"] = msg.ContentType
	inboxMessage["CreatedAt"] = time.Now().Unix()

	var (
		retInbox    []*redis.IntCmd
		retTimeline []*redis.BoolCmd
	)
	_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		retInbox = append(retInbox, pipe.LPush(ctx, ttimelinesNamer, tid))
		retInbox = append(retInbox, pipe.LPush(ctx, ftimelinesNamer, fid))
		inboxMessage["SeqId"] = tid
		retTimeline = append(retTimeline, pipe.HMSet(ctx, tinboxNamer, inboxMessage))
		inboxMessage["SeqId"] = fid
		retTimeline = append(retTimeline, pipe.HMSet(ctx, finboxNamer, inboxMessage))
		pipe.LTrim(ctx, ttimelinesNamer, 0, 10000)
		pipe.LTrim(ctx, ftimelinesNamer, 0, 10000)
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

	for _, ret := range retTimeline {
		err = ret.Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetInboxMesssage 获取邮件
func GetInboxMesssage(id uint64) (*Messages, error) {
	return nil, nil
}
