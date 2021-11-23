package sm

import (
	"context"
	"errors"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/worker"
	"github.com/doublemo/baa/kits/sm/dao"
	grpcproto "github.com/golang/protobuf/proto"
)

func broadcastMessagesToAgent(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SM_Broadcast_Messages
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	resp := &pb.SM_Broadcast_Ack{Successed: make([]uint64, 0), Failed: make([]uint64, 0)}
	if frame.Sync {
		ack := make(chan *pb.SM_Broadcast_Ack, len(frame.Values))
		fn := func(m *pb.SM_Broadcast_Message, c chan *pb.SM_Broadcast_Ack) func() {
			return func() {
				data, _ := sendMessages(m)
				c <- data
			}
		}

		for _, value := range frame.Values {
			worker.Submit(fn(value, ack))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for i := 0; i < len(frame.Values); i++ {
			select {
			case data, ok := <-ack:
				if !ok {
					continue
				}

				if len(data.Successed) > 0 {
					resp.Successed = append(resp.Successed, data.Successed...)
				}

				if len(data.Failed) > 0 {
					resp.Successed = append(resp.Failed, data.Failed...)
				}

			case <-ctx.Done():
				return nil, errors.New("task run timeout")
			}
		}
	} else {
		for _, value := range frame.Values {
			worker.Submit(func() {
				sendMessages(value)
			})
		}
	}

	w := &corespb.Response{Command: req.Command}
	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func sendMessages(message *pb.SM_Broadcast_Message) (*pb.SM_Broadcast_Ack, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	users, err := dao.GetCachedMultiUsers(ctx, message.Receiver...)
	if err != nil {
		log.Error(Logger()).Log("action", "sendMessages", "error", err)
		return &pb.SM_Broadcast_Ack{Successed: make([]uint64, 0), Failed: message.Receiver}, err
	}

	agents := make(map[string][]uint64)
	successed := make([]uint64, 0)
	failed := make([]uint64, 0)
	for _, v := range message.Receiver {
		if _, ok := users[v]; !ok {
			failed = append(failed, v)
		}
	}

	for id, values := range users {
		if len(values) < 1 {
			failed = append(failed, id)
			continue
		}

		if _, ok := agents[values[0].AgentServer]; !ok {
			agents[values[0].AgentServer] = make([]uint64, 0)
		}
		agents[values[0].AgentServer] = append(agents[values[0].AgentServer], id)
	}

	for addr, values := range agents {
		eds := endpoints.Endpoints(kit.AgentServiceName)
		if len(eds) < 1 {
			failed = append(failed, values...)
			continue
		}

		bool := false
		for _, v := range eds {
			if v == addr {
				bool = true
				break
			}
		}

		if !bool {
			failed = append(failed, values...)
			continue
		}

		m := &pb.Agent_BroadcastMessage{
			Receiver:   values,
			Command:    message.Command,
			SubCommand: message.SubCommand,
			Payload:    message.Payload,
		}

		if err := pushMessages(addr, m); err != nil {
			log.Error(Logger()).Log("action", "pushMessages", "error", err)
			failed = append(failed, values...)
			continue
		}

		successed = append(successed, values...)
	}

	return &pb.SM_Broadcast_Ack{Successed: successed, Failed: failed}, nil
}

func pushMessages(addr string, msg ...*pb.Agent_BroadcastMessage) error {
	nc := nats.Conn()
	if nc == nil {
		return errors.New("conn is nil")
	}

	frame := &pb.Agent_Broadcast{Messages: msg}
	req := coresproto.RequestBytes{
		Cmd:    kit.Agent,
		SubCmd: command.AgentBroadcast,
		SeqID:  1,
	}

	req.Content, _ = grpcproto.Marshal(frame)
	bytes, err := req.Marshal()
	if err != nil {
		return err
	}

	if err := nc.Publish(addr, bytes); err != nil {
		return err
	}

	return nc.FlushTimeout(time.Second * 10)
}
