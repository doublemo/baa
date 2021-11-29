package sm

import (
	"context"
	"errors"
	"time"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/sm/dao"
	"github.com/doublemo/baa/kits/sm/errcode"
	grpcproto "github.com/golang/protobuf/proto"
)

func getUsersStatus(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SM_User_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	valuesLen := len(frame.Values)
	if valuesLen < 1 || valuesLen > 100 {
		return errcode.Bad(w, errcode.ErrInternalServer), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	users, err := dao.GetCachedMultiUsers(ctx, frame.Values...)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	ideds := endpoints.Endpoints(kit.SNIDServiceName)
	imeds := endpoints.Endpoints(kit.IMServiceName)
	idedsMap := make(map[string]bool)
	imedsMap := make(map[string]bool)
	for _, v := range ideds {
		idedsMap[v] = true
	}

	for _, v := range imeds {
		imedsMap[v] = true
	}

	resp := &pb.SM_User_Reply{Values: make([]*pb.SM_User_Status, 0)}
	for id, all := range users {
		info := &pb.SM_User_Status{
			UserId: id,
			Values: make([]*pb.SM_User_Info, len(all)),
		}

		for i, user := range all {
			if !idedsMap[user.IDServer] {
				if m, err := assignServer(user.ID, kit.SNIDServiceName, 1); err == nil {
					user.IDServer = m
				}
			}

			if !imedsMap[user.IMServer] {
				if m, err := assignServer(user.ID, kit.IMServiceName, 1); err == nil {
					user.IMServer = m
				}
			}

			info.Values[i] = &pb.SM_User_Info{
				UserId:      user.ID,
				AgentServer: user.AgentServer,
				Platform:    user.Platform,
				Token:       user.Token,
				OnlineAt:    user.OnlineAt,
				IMServer:    user.IMServer,
				IDServer:    user.IDServer,
			}
		}
		resp.Values = append(resp.Values, info)
	}

	bytes, err := grpcproto.Marshal(resp)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}
	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
}

func getUserServers(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SM_User_Servers_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	valuesLen := len(frame.Values)
	if valuesLen < 1 || valuesLen > 100 {
		return errcode.Bad(w, errcode.ErrInternalServer), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	values, err := dao.GetMultiUserServers(ctx, frame.Values...)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.SM_User_Servers_Reply{
		Values: make([]*pb.SM_User_Servers_Info, len(values)),
	}

	i := 0
	for k, v := range values {
		resp.Values[i] = &pb.SM_User_Servers_Info{
			UserId:  k,
			Servers: v,
		}
		i++
	}

	bytes, err := grpcproto.Marshal(resp)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}
	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
}

func userAssignServer(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SM_User_Servers_AssignServerRequest
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	values, err := dao.GetUserServers(ctx, frame.UserId)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	status, err := dao.GetCachedUsers(ctx, frame.UserId)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.SM_User_Servers_AssignServerReply{
		PeerId: make(map[string]string),
		Values: &pb.SM_User_Servers_Info{
			UserId:  frame.UserId,
			Servers: make(map[string]string),
		},
	}

	for _, user := range status {
		resp.PeerId[user.Platform] = user.PeerID
	}

	newAssignServers := make(map[string]int32, 0)
	for _, v := range frame.Values {
		if m, ok := values[v.KitName]; ok && m != "" {
			eds := endpoints.Endpoints(m)
			for _, n := range eds {
				if n == m {
					resp.Values.Servers[v.KitName] = n
					break
				}
			}
		}
		newAssignServers[v.KitName] = v.LB
	}

	if len(newAssignServers) > 0 {
		for k, lb := range newAssignServers {
			if addr, err := assignServer(frame.UserId, k, int(lb)); err == nil {
				resp.Values.Servers[k] = addr
			}
		}
	}

	bytes, err := grpcproto.Marshal(resp)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
}

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
		PeerID:      frame.PeerId,
	}

	servers, err := dao.GetUserServers(context.Background(), users.ID)
	if err != nil {
		return err
	}

	ideds := endpoints.Endpoints(kit.SNIDServiceName)
	imeds := endpoints.Endpoints(kit.IMServiceName)
	idedsMap := make(map[string]bool)
	imedsMap := make(map[string]bool)
	for _, v := range ideds {
		idedsMap[v] = true
	}

	for _, v := range imeds {
		imedsMap[v] = true
	}

	if m, ok := servers[kit.IMServiceName]; ok {
		users.IMServer = m
	}

	if m, ok := servers[kit.SNIDServiceName]; ok {
		users.IDServer = m
	}

	if users.IMServer == "" || !imedsMap[users.IMServer] {
		if m, err := assignServer(users.ID, kit.IMServiceName, 1); err == nil {
			users.IMServer = m
		}
	}

	if users.IDServer == "" || !idedsMap[users.IDServer] {
		if m, err := assignServer(users.ID, kit.SNIDServiceName, 1); err == nil {
			users.IDServer = m
		}
	}

	if err := dao.Online(context.Background(), &users); err != nil {
		return err
	}

	return nil
}

func offline(evt *pb.SM_Event) error {
	var frame pb.SM_User_Action_Offline
	{
		if err := grpcproto.Unmarshal(evt.Data, &frame); err != nil {
			return err
		}
	}

	if err := dao.Offline(context.Background(), frame.UserId, frame.Platform, frame.PeerId); err != nil {
		return err
	}

	data, err := grpcproto.Marshal(&pb.SM_User_Action_CleanCache{UserId: frame.UserId})
	if err != nil {
		return err
	}

	return internalBroadcastEvent(command.SMEvent, &pb.SM_Event{Action: pb.SM_ActionUserCleanCache, Data: data})
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

func cleanCache(evt *pb.SM_Event) error {
	var frame pb.SM_User_Action_CleanCache
	{
		if err := grpcproto.Unmarshal(evt.Data, &frame); err != nil {
			return err
		}
	}

	dao.ClearUsersCachedByUserID(frame.UserId)
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
