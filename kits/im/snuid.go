package im

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	grpcpool "github.com/doublemo/baa/cores/pool/grpc"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/snid"
	grpcproto "github.com/golang/protobuf/proto"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

var (
	ErrRematchServiceID = errors.New("ErrRematchServiceID")
)

type (
	snuidRouter struct {
		c     conf.RPCClient
		pool  map[string]*grpcpool.Pool
		mutex sync.RWMutex
	}

	uidCallValue struct {
		id        uint64
		serviceID string
	}
)

func (r *snuidRouter) Serve(req *corespb.Request) (*corespb.Response, error) {
	var err error
	serviceAddr := ""
	if req.Header != nil {
		if m, ok := req.Header["service-addr"]; ok {
			serviceAddr = m
		}
	}

	if serviceAddr == "" {
		return nil, errors.New("notsupported")
	}

	r.mutex.RLock()
	p := r.pool[serviceAddr]
	r.mutex.RUnlock()

	if p == nil {
		p, err = r.createPool(serviceAddr)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := p.Get(ctx)
	if err != nil {
		return nil, err
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
	defer func() {
		conn.Close()
		cancel2()
	}()

	client := corespb.NewServiceClient(conn.ClientConn)
	resp, err := client.Call(ctx2, req)
	return resp, err
}

func (r *snuidRouter) ServeHTTP(req *corespb.Request) (*corespb.Response, error) {
	return nil, errors.New("notsupported")
}

func (r *snuidRouter) createPool(serviceAddr string) (*grpcpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	fn := func(ctx context.Context) (*grpc.ClientConn, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		opts := []grpc.DialOption{grpc.WithBlock()}
		if len(r.c.Key) > 0 && len(r.c.Salt) > 0 {
			creds, err := credentials.NewClientTLSFromFile(r.c.Salt, r.c.Key)
			if err != nil {
				return nil, err
			}

			opts = append(opts, grpc.WithTransportCredentials(creds))
			opts = append(opts, grpc.WithPerRPCCredentials(oauth.NewOauthAccess(
				&oauth2.Token{AccessToken: r.c.ServiceSecurityKey},
			)))
		} else {
			opts = append(opts, grpc.WithInsecure())
		}

		conn, err := grpc.Dial(serviceAddr, opts...)
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
	r.pool[serviceAddr] = p
	r.mutex.Unlock()
	return p, nil
}

func newSnuidRouter(c conf.RPCClient) *snuidRouter {
	return &snuidRouter{c: c, pool: make(map[string]*grpcpool.Pool)}
}

func getUID(status map[uint64]map[string]string, values ...uint64) (map[uint64]uint64, []uint64, error) {
	needRematch := make(map[uint64]bool)
	siddata := make([]string, 0)
	siddatamap := make(map[string]map[uint64]bool)
	for _, id := range values {
		if m, ok := status[id]; ok {
			if addr, ok := m[snid.ServiceName]; ok {
				if _, ok := siddatamap[addr]; !ok {
					siddatamap[addr] = make(map[uint64]bool)
				}

				siddata = append(siddata, addr)
				siddatamap[addr][id] = true
			} else {
				needRematch[id] = true
				continue
			}
		} else {
			needRematch[id] = true
			continue
		}
	}

	if len(siddata) < 1 {
		rematch := make([]uint64, len(needRematch))
		i := 0
		for v := range needRematch {
			rematch[i] = v
		}
		return nil, rematch, ErrRematchServiceID
	}

	addrs, err := sd.GetEndpointsByID(siddata...)
	if err != nil {
		return nil, nil, err
	}

	newValues := make(map[string][]uint64)
	newValuesMap := make(map[string]map[uint64]bool)
	for _, v := range siddata {
		if m, ok := addrs[v]; ok {
			if _, ok := newValues[m.Addr()]; !ok {
				newValues[m.Addr()] = make([]uint64, 0)
				newValuesMap[m.Addr()] = make(map[uint64]bool)
			}

			if uids, ok := siddatamap[v]; ok {
				for uid := range uids {
					if newValuesMap[m.Addr()][uid] {
						continue
					}

					newValues[m.Addr()] = append(newValues[m.Addr()], uid)
					newValuesMap[m.Addr()][uid] = true
				}
			}
		} else {
			if uids, ok := siddatamap[v]; ok {
				for uid := range uids {
					needRematch[uid] = true
				}
			}
		}
	}

	if len(needRematch) > 0 {
		rematch := make([]uint64, len(needRematch))
		i := 0
		for v := range needRematch {
			rematch[i] = v
		}
		return nil, rematch, ErrRematchServiceID
	}

	retValues := make(map[uint64]uint64)
	for kk, vv := range newValues {
		frame := pb.SNID_MoreRequest{
			Request: make([]*pb.SNID_Request, 0),
		}

		for _, id := range vv {
			frame.Request = append(frame.Request, &pb.SNID_Request{K: namerUserTimeline(id), N: 1})
		}

		b, _ := grpcproto.Marshal(&frame)
		resp, err := muxRouter.Handler(snid.ServiceName, &corespb.Request{Command: command.SNIDMoreAutoincrement.Int32(), Payload: b, Header: map[string]string{"service-addr": kk}})
		if err != nil {
			return nil, nil, err
		}

		switch payload := resp.Payload.(type) {
		case *corespb.Response_Content:
			resp := pb.SNID_MoreReply{}
			if err := grpcproto.Unmarshal(payload.Content, &resp); err != nil {
				return nil, nil, err
			}

			for _, id := range vv {
				if m, ok := resp.Values[namerUserTimeline(id)]; ok && len(m.Values) > 0 {
					retValues[id] = m.Values[0]
				}
			}

		case *corespb.Response_Error:
			return nil, nil, errors.New(payload.Error.Message)
		}
	}

	return retValues, nil, nil
}

func namerUserTimeline(id uint64) string {
	return strconv.FormatUint(id, 10) + ":timelineid"
}
