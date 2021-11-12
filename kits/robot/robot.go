package robot

import (
	"errors"

	"github.com/doublemo/baa/cores/crypto/id"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/robot/dao"
	"github.com/doublemo/baa/kits/robot/errcode"
	grpcproto "github.com/golang/protobuf/proto"
)

// RobotConfig 机人配置
type RobotConfig struct {
	// IDSecret 用户ID 加密key 16位
	IDSecret string `alias:"idSecret" default:"7581BDD8E8DA3839"`
}

func createRobot(req *corespb.Request, c RobotConfig) (*corespb.Response, error) {
	var frame pb.Robot_Create_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	switch payload := frame.Payload.(type) {
	case *pb.Robot_Create_Request_Account:
		return createRobotByAccount(req, payload, c)

	case *pb.Robot_Create_Request_Register:
		return createRobotByRegister(req, payload, c)
	}

	return nil, errors.New("notsupported")
}

func createRobotByAccount(req *corespb.Request, frame *pb.Robot_Create_Request_Account, c RobotConfig) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	account, err := internalAccountLogin(frame.Account.Name, frame.Account.Secret)
	if err != nil {
		return errcode.Bad(w, errcode.ErrUsernameOrPasswordIncorrect, err.Error()), nil
	}

	aid, err := id.Decrypt(account.ID, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	unid, err := id.Decrypt(account.UnionID, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	uid, err := id.Decrypt(account.UserID, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	userinfo, err := internalUserinfo(account.ID)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	robot := dao.Robots{
		AccountID: aid,
		UnionID:   unid,
		UserID:    uid,
		Schema:    "Account",
		Name:      frame.Account.Name,
		Secret:    frame.Account.Secret,
		IndexNo:   userinfo.IndexNo,
		Nickname:  userinfo.Nickname,
		Headimg:   userinfo.Headimg,
		Age:       int8(userinfo.Age),
		Sex:       int8(userinfo.Sex),
		Idcard:    userinfo.Idcard,
		Phone:     userinfo.Phone,
	}

	// 获取账户信息
	if err := dao.CreateRobot(&robot); err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	respFrame := pb.Robot_Create_Reply{
		OK: true,
	}

	bytes, _ := grpcproto.Marshal(&respFrame)
	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
}

func createRobotByRegister(req *corespb.Request, frame *pb.Robot_Create_Request_Register, c RobotConfig) (*corespb.Response, error) {
	return nil, nil
}
