package auth

import (
	"errors"
	"time"

	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

const (
	// channelStateEventBroadcast 状态变更时通知通道
	channelStateEventBroadcast = "sm.broadcast"

	// channelStateEventReceiver 接收状态事件通道
	channelStateEventReceiver = "sm.receiver"
)

func publishUserState(frame *pb.SM_Event) error {
	nc := nats.Conn()
	if nc == nil {
		return errors.New("nats is nil")
	}

	req := coresproto.RequestBytes{
		Cmd:    kit.SM,
		SubCmd: command.SMEvent,
	}

	req.Content, _ = grpcproto.Marshal(frame)
	data, _ := req.Marshal()
	if err := nc.Publish(channelStateEventReceiver, data); err != nil {
		return err
	}

	return nc.FlushTimeout(time.Second)
}
