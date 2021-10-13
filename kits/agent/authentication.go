package agent

import (
	"context"
	"strconv"
	"sync"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	grpcpool "github.com/doublemo/baa/cores/pool/grpc"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/kits/agent/errcode"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/session"
	authproto "github.com/doublemo/baa/kits/auth/proto"
	authpb "github.com/doublemo/baa/kits/auth/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

type authenticationRouter struct {
	c     conf.RPCClient
	pool  *grpcpool.Pool
	mutex sync.RWMutex
}

func (r *authenticationRouter) Serve(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	resp, err := r.call(&corespb.Request{
		Header:  map[string]string{"PeerId": peer.ID(), "seqno": strconv.FormatUint(uint64(req.SID()), 10)},
		Command: req.SubCommand().Int32(),
		Payload: req.Body(),
	})

	if err != nil {
		return proto.NewResponseBytes(req.Command(), errcode.Bad(&corespb.Response{Command: req.SubCommand().Int32()}, errcode.ErrInternalServer, grpc.ErrorDesc(err))), nil
	}

	// 处理登录后缓存用户信息
	r.onLogin(peer, resp)
	w := proto.NewResponseBytes(req.Command(), resp)
	return w, nil
}

func (r *authenticationRouter) call(req *corespb.Request) (*corespb.Response, error) {
	p, err := r.createPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := p.Get(ctx)
	if err != nil {
		return nil, err
	}

	client := corespb.NewServiceClient(conn.ClientConn)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		conn.Close()
		cancel2()
	}()

	return client.Call(ctx2, req)
}

func (r *authenticationRouter) createPool() (*grpcpool.Pool, error) {
	r.mutex.RLock()
	if r.pool != nil {
		r.mutex.RUnlock()
		return r.pool, nil
	}
	r.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	fn := func(ctx context.Context) (*grpc.ClientConn, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		conn, err := rpc.NewConnect(r.c)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	p, err := grpcpool.NewWithContext(ctx, fn, r.c.Pool.Init, r.c.Pool.Capacity, time.Duration(r.c.Pool.IdleTimeout)*time.Minute, time.Duration(r.c.Pool.MaxLife)*time.Minute)
	if err != nil {
		return nil, err
	}

	r.mutex.Lock()
	if r.pool == nil {
		r.pool = p
	} else {
		p = r.pool
	}
	r.mutex.Unlock()
	return p, nil
}

// Destroy 清理
func (r *authenticationRouter) Destroy(peer session.Peer) {
	accountID, ok := peer.Params("AccountID")
	if !ok {
		return
	}

	// 处理玩家离线
	frame := authpb.Authentication_Form_Logout{
		Payload: &authpb.Authentication_Form_Logout_PeerID{PeerID: peer.ID()},
	}

	body, _ := grpcproto.Marshal(&frame)
	_, err := r.call(&corespb.Request{
		Header:  map[string]string{"PeerId": peer.ID(), "AccountID": accountID.(string)},
		Command: authproto.OfflineCommand.Int32(),
		Payload: body,
	})

	if err != nil {
		log.Error(Logger()).Log("action", "Destroy", "error", err)
	}
}

func (r *authenticationRouter) onLogin(peer session.Peer, w *corespb.Response) {
	if w.Command == authproto.LoginCommand.Int32() {
		return
	}

	var content []byte
	switch payload := w.Payload.(type) {
	case *corespb.Response_Content:
		content = payload.Content
	default:
		return
	}

	var frame authpb.Authentication_Form_LoginReply
	{
		if err := grpcproto.Unmarshal(content, &frame); err != nil {
			log.Error(Logger()).Log("action", "onLogin", "error", err)
			return
		}
	}

	switch payload := frame.Payload.(type) {
	case *authpb.Authentication_Form_LoginReply_Account:
		peer.SetParams("AccountID", payload.Account.ID)
		peer.SetParams("AccountUnionID", payload.Account.UnionID)
	default:
		return
	}
}

func newAuthenticationRouter(config conf.RPCClient) *authenticationRouter {
	return &authenticationRouter{
		c: config,
	}
}
