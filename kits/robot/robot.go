package robot

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/doublemo/baa/cores/crypto/aes"
	"github.com/doublemo/baa/cores/crypto/id"
	log "github.com/doublemo/baa/cores/log/level"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/internal/worker"
	"github.com/doublemo/baa/kits/robot/dao"
	"github.com/doublemo/baa/kits/robot/errcode"
	"github.com/doublemo/baa/kits/robot/session"
	grpcproto "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// RobotConfig 机人配置
type RobotConfig struct {
	// IDSecret 用户ID 加密key 16位
	IDSecret string `alias:"idSecret" default:"7581BDD8E8DA3839"`

	// PasswordSecret 密码 加密key 16位
	PasswordSecret string `alias:"passwordSecret" default:"7531BDD8E5DA38397531BDD8E5DA3839"`

	// PasswordMinLen 密码最少字符
	PasswordMinLen int `alias:"passwordMinLen" default:"8"`

	// PasswordMaxLen 密码最大字符
	PasswordMaxLen int `alias:"passwordMaxLen" default:"16"`

	// NicknameMaxLength 昵称最大长度
	NicknameMaxLength int `alias:"nicknameMaxLength" default:"34"`

	// StartIntervalTime 机器人启动时间间隔
	StartIntervalTime int `alias:"startIntervalTime" default:"34"`

	// ReadBufferSize 读取缓存大小 32767
	ReadBufferSize int `alias:"readbuffersize" default:"32767"`

	// WriteBufferSize 写入缓存大小 32767
	WriteBufferSize int `alias:"writebuffersize" default:"32767"`

	// ReadDeadline 读取超时
	ReadDeadline int `alias:"readdeadline" default:"310"`

	// WriteDeadline 写入超时
	WriteDeadline int `alias:"writedeadline"`

	// NetProtocol 机器人连接网络协议 scoket websocket
	NetProtocol string `alias:"netProtocol" default:"socket"`

	Datachannel webrtc.Configuration `alias:"-"`
}

func startRobots(req *corespb.Request, c RobotConfig) (*corespb.Response, error) {
	var frame pb.Robot_Start_Request
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
	if valuesLen > 1000 || valuesLen < 1 {
		return errcode.Bad(w, errcode.ErrRobotsTooMany), nil
	}

	robotsIDValues := make([]uint64, valuesLen)
	robotsMap := make(map[uint64]*pb.Robot_Start_Robot)
	for k, v := range frame.Values {
		robotsIDValues[k] = v.ID
		robotsMap[v.ID] = v
	}

	robots, err := dao.FindRobotsInID(robotsIDValues...)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	successed := make([]uint64, 0)
	successedmap := make(map[uint64]bool)
	failed := make([]uint64, 0)
	failedmap := make(map[uint64]bool)
	exitchan, err := session.GetCtlChannel()
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	for _, robot := range robots {
		successed = append(successed, robot.ID)
		successedmap[robot.ID] = true
		task, ok := robotsMap[robot.ID]
		if !ok {
			continue
		}

		if frame.Async {
			worker.Submit(func(rob *dao.Robots, t *pb.Robot_Start_Robot, config RobotConfig) func() {
				return func() {
					if err := runRobot(rob, t, c, exitchan); err != nil {
						log.Error(Logger()).Log("action", "runRobot", "error", err)
					}
				}
			}(robot, task, c))
		} else {
			if err := runRobot(robot, task, c, exitchan); err != nil {
				failed = append(failed, robot.ID)
				failedmap[robot.ID] = true
			}
		}

		if c.StartIntervalTime > 0 {
			time.Sleep(time.Duration(c.StartIntervalTime) * time.Millisecond)
		}
	}

	for _, v := range frame.Values {
		if !successedmap[v.ID] && !failedmap[v.ID] {
			failed = append(failed, v.ID)
			failedmap[v.ID] = true
		}
	}

	bytes, err := grpcproto.Marshal(&pb.Robot_Start_Reply{Succeeded: successed, Failed: failed})
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
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

	if rob, err := dao.FindRobotsByAccountID(aid, "id"); err == nil && rob.ID > 0 {
		return errcode.Bad(w, errcode.ErrRobotsIsExists), nil
	}

	unid, err := id.Decrypt(account.UnionID, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	uid, err := id.Decrypt(account.UserID, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	userinfo, err := internalGetUserinfo(account.ID)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	password, err := encryptPassword(frame.Account.Secret, []byte(c.PasswordSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	robot := dao.Robots{
		AccountID:  aid,
		UnionID:    unid,
		UserID:     uid,
		SchemaName: "password",
		Name:       frame.Account.Name,
		Secret:     password,
		IndexNo:    userinfo.IndexNo,
		Nickname:   userinfo.Nickname,
		Headimg:    userinfo.Headimg,
		Age:        int8(userinfo.Age),
		Sex:        int8(userinfo.Sex),
		Idcard:     userinfo.Idcard,
		Phone:      userinfo.Phone,
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
	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	// ^[\u4e00-\u9fa5]
	reg := regexp.MustCompile(`^[a-z0-9A-Z\p{Han}]+([\.|_|@][a-z0-9A-Z\p{Han}]+)*$`)
	if !reg.MatchString(frame.Register.Name) {
		return errcode.Bad(w, errcode.ErrAccountNameLettersInvalid), nil
	}

	if !isValidPassword(frame.Register.Secret, c.PasswordMinLen, c.PasswordMaxLen) {
		return errcode.Bad(w, errcode.ErrPasswordLettersInvalid, fmt.Sprintf(errcode.ErrPasswordLettersInvalid.Error(), c.PasswordMinLen, c.PasswordMaxLen)), nil
	}

	if len(frame.Register.Nickname) > c.NicknameMaxLength {
		return errcode.Bad(w, errcode.ErrAccountNameLettersInvalid), nil
	}

	if rob, err := dao.FindRobotsBySchemaName("password", frame.Register.Name, "id"); err == nil && rob.ID > 0 {
		return errcode.Bad(w, errcode.ErrRobotsIsExists), nil
	}

	frame.Register.Schema = "password"
	request := &corespb.Request{
		Header:  map[string]string{"From": "Robot"},
		Command: int32(command.AuthRegister),
	}

	request.Payload, _ = grpcproto.Marshal(&pb.Authentication_Form_Register{
		Scheme: "password",
		Payload: &pb.Authentication_Form_Register_Robot{
			Robot: &pb.Authentication_Form_RegisterRobot{
				Username: frame.Register.Name,
				Password: frame.Register.Secret,
			},
		},
	})

	resp, err := muxRouter.Handler(kit.Auth.Int32(), request)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	var accountInfo *pb.Authentication_Form_AccountInfo
	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		var data pb.Authentication_Form_RegisterReply
		if err := grpcproto.Unmarshal(payload.Content, &data); err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

		switch p := data.Payload.(type) {
		case *pb.Authentication_Form_RegisterReply_Account:
			accountInfo = p.Account
		default:
			return errcode.Bad(w, errcode.ErrInternalServer), nil
		}

	case *corespb.Response_Error:
		return errcode.Bad(w, errcode.ErrInternalServer, payload.Error.Message), nil
	default:
		return errcode.Bad(w, errcode.ErrInternalServer), nil
	}

	accountsId, err := id.Decrypt(accountInfo.ID, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	unionId, err := id.Decrypt(accountInfo.UnionID, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	userId, err := id.Decrypt(accountInfo.UserID, []byte(c.IDSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	password, err := encryptPassword(frame.Register.Secret, []byte(c.PasswordSecret))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	robots := dao.Robots{
		AccountID:  accountsId,
		UnionID:    unionId,
		UserID:     userId,
		SchemaName: "password",
		Name:       frame.Register.Name,
		Secret:     password,
		IndexNo:    "",
		Nickname:   frame.Register.Nickname,
		Headimg:    frame.Register.Headimg,
		Age:        int8(frame.Register.Age),
		Sex:        int8(frame.Register.Sex),
		Idcard:     frame.Register.Idcard,
		Phone:      frame.Register.Phone,
	}

	if err := dao.CreateRobot(&robots); err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	infoRequest := &corespb.Request{
		Header:  map[string]string{"UserID": accountInfo.UserID, "AccountID": accountInfo.ID},
		Command: int32(command.UserRegister),
	}

	infoRequest.Payload, _ = grpcproto.Marshal(&pb.User_Register_Request{
		AccountId: accountInfo.ID,
		Info: &pb.User_Info{
			UserId:   accountInfo.UserID,
			Nickname: robots.Nickname,
			Headimg:  robots.Headimg,
			Age:      int32(robots.Age),
			Sex:      int32(robots.Sex),
			Idcard:   robots.Idcard,
			Phone:    robots.Phone,
		},
	})

	resp, err = muxRouter.Handler(kit.User.Int32(), infoRequest)
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	var userReply pb.User_Register_Reply
	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		if err := grpcproto.Unmarshal(payload.Content, &userReply); err != nil {
			return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
		}

	case *corespb.Response_Error:
		return errcode.Bad(w, errcode.ErrInternalServer, payload.Error.Message), nil
	default:
		return errcode.Bad(w, errcode.ErrInternalServer), nil
	}

	if _, err := dao.UpdatesRobotsByID(robots.ID, "index_no", userReply.IndexNo); err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	respFrame := pb.Robot_Create_Reply{
		OK: true,
	}

	bytes, _ := grpcproto.Marshal(&respFrame)
	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
}

func isValidPassword(str string, minLen, maxLen int) bool {
	var (
		isUpper   = false
		isLower   = false
		isNumber  = false
		isSpecial = false
	)

	if len(str) < minLen || len(str) > maxLen {
		return false
	}

	for _, s := range str {
		switch {
		case unicode.IsUpper(s):
			isUpper = true
		case unicode.IsLower(s):
			isLower = true
		case unicode.IsNumber(s):
			isNumber = true
		case unicode.IsPunct(s) || unicode.IsSymbol(s):
			isSpecial = true
		default:
		}
	}
	return (isUpper && isLower) && (isNumber || isSpecial)
}

func encryptPassword(password string, key []byte) (string, error) {
	dst, err := aes.Encrypt([]byte(password), key)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(dst), nil
}

func decryptPassword(password string, key []byte) (string, error) {
	m, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(password)
	if err != nil {
		return "", err
	}

	w, err := aes.Decrypt(m, key)
	if err != nil {
		return "", err
	}

	return string(w), nil
}

func runRobot(robot *dao.Robots, task *pb.Robot_Start_Robot, c RobotConfig, exitChan chan struct{}) error {
	if m, ok := session.GetPeer(strconv.FormatUint(robot.ID, 10)); ok {
		session.RemovePeer(m)
		m.Close()
	}

	netProtocol := "socket"
	if c.NetProtocol == "websocket" || c.NetProtocol == "websockets" {
		netProtocol = c.NetProtocol
	}

	var agent string
	re := regexp.MustCompile(`^[\d]{1,3}\.[\d]{1,3}\.[\d]{1,3}\.[\d]{1,3}(:\d{1,5})?$`)
	if re.MatchString(robot.Agent) {
		agent = robot.Agent
	} else if robot.Agent != "" {
		if m, err := selectAgent(robot.Agent, netProtocol); err == nil {
			agent = m
		}
	} else {
		if m, err := randomAgent(netProtocol); err == nil {
			agent = m
		}
	}

	if agent == "" {
		return errors.New("No gateway server can be used")
	}

	var peer session.Peer
	if netProtocol == "websocket" || netProtocol == "websockets" {
		scheme := "ws"
		if netProtocol == "websockets" {
			scheme = "wss"
		}

		wsaddr := url.URL{Scheme: scheme, Host: agent, Path: "/websocket"}
		dialer := &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		conn, _, err := dialer.DialContext(ctx, wsaddr.String(), nil)
		if err != nil {
			cancel()
			return err
		}
		cancel()
		peer = session.NewPeerWebsocket(strconv.FormatUint(robot.ID, 10), conn, time.Duration(c.ReadDeadline)*time.Second, time.Duration(c.WriteDeadline)*time.Second, 1048576, exitChan)
	} else {
		conn, err := net.DialTimeout("tcp", agent, 30*time.Second)
		if err != nil {
			return err
		}
		peer = session.NewPeerSocket(strconv.FormatUint(robot.ID, 10), conn, time.Duration(c.ReadDeadline)*time.Second, time.Duration(c.WriteDeadline)*time.Second, exitChan)
	}

	peer.OnReceive(onMessage)
	peer.OnClose(func(p session.Peer) {
		session.RemovePeer(p)
		log.Debug(Logger()).Log("Robot", p.ID(), "action", "shutdown")
	})
	peer.OnTimeout(doRequestHeartbeater)
	peer.Go()
	session.AddPeer(peer)

	// 设置Peer数据
	peer.SetParams("Robots", robot)
	peer.SetParams("Task", task)

	log.Debug(Logger()).Log("Robot", peer.ID(), "action", "started")
	// 开始握手
	return doRequestHandshake(peer)
}

func randomAgent(netProtocol string) (string, error) {
	eds, err := sd.Endpoints()
	if err != nil {
		return "", err
	}

	agents := make([]string, 0)
	for _, ed := range eds {
		if ed.Name() == kit.AgentServiceName {
			agents = append(agents, ed.Get("ip")+ed.Get(netProtocol))
		}
	}

	if len(agents) < 1 {
		return "", errors.New("No gateway server can be used")
	}
	return agents[rand.Intn(len(agents))], nil
}

func selectAgent(group, netProtocol string) (string, error) {
	eds, err := sd.Endpoints()
	if err != nil {
		return "", err
	}

	agents := make([]string, 0)
	for _, ed := range eds {
		if ed.Name() == kit.AgentServiceName && ed.Get("group") == group {
			agents = append(agents, ed.Get("ip")+ed.Get(netProtocol))
		}
	}

	if len(agents) < 1 {
		return "", errors.New("No gateway server can be used")
	}

	return agents[rand.Intn(len(agents))], nil
}
