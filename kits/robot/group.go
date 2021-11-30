package robot

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/robot/session"
	"github.com/golang/protobuf/jsonpb"
)

// doCreateAndJoinGroup 创建并加入群聊
func doCreateAndJoinGroup(peer session.Peer, c RobotConfig) error {
	userid, ok := peer.Params("UserID")
	if !ok {
		return errors.New("invalid UserID")
	}

	uid := userid.(string)
	agent, ok := peer.Params("AgentHttp")
	if !ok {
		return errors.New("invalid agent addr")
	}

	tk, ok := peer.Params("Token")
	if !ok {
		return errors.New("invalid token")
	}

	frame := &pb.User_Group_Create_Request{
		UserId:  uid,
		Members: []string{"NJlPF3UZcr0"},
	}

	pm := jsonpb.Marshaler{}
	data, err := pm.MarshalToString(frame)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	body, errcode := RequestPostWithContext(ctx, command.UserCreateGroup, agent.(string)+"/v1/user", []byte(data), []byte(c.CommandSecret), tk.(string), c.CSRFSecret)
	if errcode != nil {
		return errcode.ToError()
	}

	var frameW pb.User_Group_Create_Reply
	{
		if err := jsonpb.UnmarshalString(string(body), &frameW); err != nil {
			log.Error(Logger()).Log("action", "doCreateAndJoinGroup", "error", err.Error(), "body", string(body))
			return err
		}
	}

	fmt.Println(frameW.Info)
	return nil
}
