package sm

import (
	"context"
	"errors"
	"time"

	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/sm/dao"
	grpcproto "github.com/golang/protobuf/proto"
)

func online(evt *pb.SM_Event) error {
	var frame pb.SM_User_Action_Online
	{
		if err := grpcproto.Unmarshal(evt.Data, &frame); err != nil {
			return err
		}
	}

	users := dao.Users{
		ID:          frame.UserId,
		AgentServer: frame.Agent,
		Platform:    frame.Platform,
		Token:       frame.Token,
		OnlineAt:    time.Now().Unix(),
	}

	servers, err := dao.GetUserServers(context.Background(), users.ID)
	if err != nil {
		return err
	}

	if m, ok := servers[kit.IMServiceName]; ok {
		users.IMServer = m
	}

	if m, ok := servers[kit.SNIDServiceName]; ok {
		users.IDServer = m
	}

	if users.IMServer == "" {
		if m, err := assignServer(users.ID, kit.IMServiceName, 1); err == nil {
			users.IMServer = m
		}
	}

	if users.IDServer == "" {
		if m, err := assignServer(users.ID, kit.SNIDServiceName, 1); err == nil {
			users.IDServer = m
		}
	}

	if err := dao.Online(context.Background(), &users); err != nil {
		return err
	}

	return dao.UpdateUsersServer(context.Background(), frame.UserId, kit.AgentServiceName, users.AgentServer)
}

func offline(evt *pb.SM_Event) error {
	var frame pb.SM_User_Action_Offline
	{
		if err := grpcproto.Unmarshal(evt.Data, &frame); err != nil {
			return err
		}
	}
	return dao.Offline(context.Background(), frame.UserId, frame.Platform)
}

func updateUserStatus(evt *pb.SM_Event) error {
	var frame pb.SM_User_Action_Update
	{
		if err := grpcproto.Unmarshal(evt.Data, &frame); err != nil {
			return err
		}
	}
	return nil
}

func assignServer(userid uint64, server string, lb int) (string, error) {
	var (
		addr string
		ok   bool
	)

	switch lb {
	case 1:
		addr, ok = endpoints.RoundRobin(server)
	default:
		addr, ok = endpoints.Random(server)
	}

	if !ok {
		return "", errors.New("No specified server was found")
	}

	err := dao.UpdateUsersServer(context.Background(), userid, server, addr)
	return addr, err
}
