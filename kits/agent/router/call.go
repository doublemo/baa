package router

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/doublemo/baa/cores/crypto/id"
	coreslog "github.com/doublemo/baa/cores/log"
	log "github.com/doublemo/baa/cores/log/level"
	grpcpool "github.com/doublemo/baa/cores/pool/grpc"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/agent/errcode"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/session"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

const (
	// defaultMaxQueryLength 默认http query最大长度
	defaultMaxQueryLength int = 1024

	// defaultMaxBytesReader 限制最大body大小
	defaultMaxBytesReader int64 = 1 << 20
)

type (
	// Call 处理路由Call方法
	Call struct {
		c                    conf.RPCClient
		pool                 map[string]*grpcpool.Pool
		destroyInterceptors  atomic.Value
		requestInterceptors  atomic.Value
		responseInterceptors atomic.Value
		commandSecret        []byte
		maxQureyLength       int
		maxBytesReader       int64
		mutex                sync.RWMutex
		logger               coreslog.Logger
	}

	CallOptions func(c *Call)
)

// Serve 服务处理
func (r *Call) Serve(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	request := &corespb.Request{
		Header: map[string]string{
			"PeerId":       peer.ID(),
			"seqno":        strconv.FormatUint(uint64(req.SID()), 10),
			"Content-Type": "stream",
			"Agent":        sd.Endpoint().ID(),
		},
		Command: req.SubCommand().Int32(),
		Payload: req.Body(),
	}

	if handler, ok := r.requestInterceptors.Load().(RequestInterceptors); ok && handler != nil {
		m := handler.Process(RequestInterceptorFunc(func(args RequestInterceptorArgs) error {
			return nil
		}))

		if err := m.Process(RequestInterceptorArgs{peer, request, req}); err != nil {
			return nil, err
		}
	}

	resp, err := r.Call(request)
	if err != nil {
		return proto.NewResponseBytes(req.Command(), errcode.Bad(&corespb.Response{Command: req.SubCommand().Int32()}, errcode.ErrInternalServer, grpc.ErrorDesc(err))), nil
	}

	if handler, ok := r.responseInterceptors.Load().(ResponseInterceptors); ok && handler != nil {
		m := handler.Process(ResponseInterceptorFunc(func(args ResponseInterceptorArgs) error {
			return nil
		}))

		if err := m.Process(ResponseInterceptorArgs{peer, resp}); err != nil {
			return nil, err
		}
	}

	w := proto.NewResponseBytes(req.Command(), resp)
	return w, nil
}

func (r *Call) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	cmd := vars["command"]
	if cmd == "" {
		http.Error(rw, "feature not supported", http.StatusBadRequest)
		return
	}

	command, err := id.Decrypt(cmd, []byte(r.commandSecret))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	request, err := r.buildCoresPBRequest(rw, req)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	request.Command = int32(command)
	request.Header["Content-Type"] = "json"
	request.Header["X-Session-Token"] = req.Header.Get("X-Session-Token")
	request.Header["Agent"] = sd.Endpoint().ID()
	if handler, ok := r.requestInterceptors.Load().(RequestInterceptors); ok && handler != nil {
		m := handler.Process(RequestInterceptorFunc(func(args RequestInterceptorArgs) error {
			return nil
		}))

		if err := m.Process(RequestInterceptorArgs{nil, request, nil}); err != nil {
			errData, _ := json.Marshal(&corespb.Error{Code: errcode.ErrInternalServer.Code(), Message: err.Error()})
			http.Error(rw, string(errData), http.StatusBadRequest)
			return
		}
	}

	resp, err := r.Call(request)
	if err != nil {
		log.Error(r.logger).Log("action", "call", "error", err)
		errData, _ := json.Marshal(&corespb.Error{Code: errcode.ErrInternalServer.Code(), Message: err.Error()})
		http.Error(rw, string(errData), http.StatusBadRequest)
		return
	}

	if handler, ok := r.responseInterceptors.Load().(ResponseInterceptors); ok && handler != nil {
		m := handler.Process(ResponseInterceptorFunc(func(args ResponseInterceptorArgs) error {
			return nil
		}))

		if err := m.Process(ResponseInterceptorArgs{nil, resp}); err != nil {
			errData, _ := json.Marshal(&corespb.Error{Code: errcode.ErrInternalServer.Code(), Message: err.Error()})
			http.Error(rw, string(errData), http.StatusBadRequest)
			return
		}
	}

	rw.Header().Set("Content-Type", "application/json;charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	switch rwPaylod := resp.Payload.(type) {
	case *corespb.Response_Error:
		ret, _ := json.Marshal(rwPaylod.Error)
		io.WriteString(rw, string(ret))

	case *corespb.Response_Content:
		io.WriteString(rw, string(rwPaylod.Content))
	}
}

// Call 调用
func (r *Call) Call(req *corespb.Request) (*corespb.Response, error) {
	addr := "default"
	if m, ok := req.Header["Host"]; ok && m != "" {
		addr = m
	}

	p, err := r.createPool(addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	conn, err := p.Get(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		conn.Close()
	}()

	client := corespb.NewServiceClient(conn.ClientConn)
	return client.Call(ctx, req)
}

func (r *Call) createPool(addr string) (*grpcpool.Pool, error) {
	r.mutex.Lock()
	pl, ok := r.pool[addr]
	r.mutex.Unlock()
	if !ok && pl != nil {
		return pl, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	fn := r.selectFn(addr)
	p, err := grpcpool.NewWithContext(ctx, fn, r.c.Pool.Init, r.c.Pool.Capacity, time.Duration(r.c.Pool.IdleTimeout)*time.Minute, time.Duration(r.c.Pool.MaxLife)*time.Minute)
	if err != nil {
		return nil, err
	}

	r.mutex.Lock()
	r.pool[addr] = p
	r.mutex.Unlock()
	return p, nil
}

func (r *Call) selectFn(addr string) func(ctx context.Context) (*grpc.ClientConn, error) {
	if addr == "default" || addr == "" {
		return func(ctx context.Context) (*grpc.ClientConn, error) {
			conn, err := rpc.NewConnectContext(ctx, r.c)
			if err != nil {
				return nil, err
			}

			return conn, nil
		}
	}

	return func(ctx context.Context) (*grpc.ClientConn, error) {
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
		conn, err := grpc.DialContext(ctx, addr, opts...)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

// Destroy 清理
func (r *Call) Destroy(peer session.Peer) error {
	if handler, ok := r.destroyInterceptors.Load().(ResponseInterceptors); ok && handler != nil {
		m := handler.Process(ResponseInterceptorFunc(func(args ResponseInterceptorArgs) error {
			return nil
		}))

		if err := m.Process(ResponseInterceptorArgs{peer, nil}); err != nil {
			return err
		}
	}

	return nil
}

// UseDestroyInterceptor 存储当路由清理时响应函数
func (r *Call) UseDestroyInterceptor(f ...func(ResponseInterceptor) ResponseInterceptor) {
	if len(f) < 1 {
		return
	}

	newInterceptors := make(ResponseInterceptors, len(f))
	for i, p := range f {
		newInterceptors[i] = p
	}

	interceptors, ok := r.destroyInterceptors.Load().(ResponseInterceptors)
	if !ok || interceptors == nil || len(interceptors) < 1 {
		r.destroyInterceptors.Store(newInterceptors)
		return
	}

	interceptorsLen := len(interceptors)
	newInterceptorsLen := len(newInterceptors)
	data := make(ResponseInterceptors, interceptorsLen+newInterceptorsLen)
	copy(data[0:interceptorsLen], interceptors[0:])
	copy(data[interceptorsLen:], newInterceptors[0:])
	r.destroyInterceptors.Store(data)
}

// UseRequestInterceptor Hook
func (r *Call) UseRequestInterceptor(f ...func(RequestInterceptor) RequestInterceptor) {
	if len(f) < 1 {
		return
	}

	newInterceptors := make(RequestInterceptors, len(f))
	for i, p := range f {
		newInterceptors[i] = p
	}

	interceptors, ok := r.requestInterceptors.Load().(RequestInterceptors)
	if !ok || interceptors == nil || len(interceptors) < 1 {
		r.requestInterceptors.Store(newInterceptors)
		return
	}

	interceptorsLen := len(interceptors)
	newInterceptorsLen := len(newInterceptors)
	data := make(RequestInterceptors, interceptorsLen+newInterceptorsLen)
	copy(data[0:interceptorsLen], interceptors[0:])
	copy(data[interceptorsLen:], newInterceptors[0:])
	r.requestInterceptors.Store(data)
}

// OnAfterCall Hook
func (r *Call) UseResponseInterceptor(f ...func(ResponseInterceptor) ResponseInterceptor) {
	if len(f) < 1 {
		return
	}

	newInterceptors := make(ResponseInterceptors, len(f))
	for i, p := range f {
		newInterceptors[i] = p
	}

	interceptors, ok := r.responseInterceptors.Load().(ResponseInterceptors)
	if !ok || interceptors == nil || len(interceptors) < 1 {
		r.responseInterceptors.Store(newInterceptors)
		return
	}

	interceptorsLen := len(interceptors)
	newInterceptorsLen := len(newInterceptors)
	data := make(ResponseInterceptors, interceptorsLen+newInterceptorsLen)
	copy(data[0:interceptorsLen], interceptors[0:])
	copy(data[interceptorsLen:], newInterceptors[0:])
	r.responseInterceptors.Store(data)
}

func (r *Call) resolveContentType(req *http.Request) string {
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		return "text/html"
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}

func (r *Call) buildCoresPBRequest(rw http.ResponseWriter, req *http.Request) (*corespb.Request, error) {
	switch strings.ToUpper(req.Method) {
	case "GET":
		return r.buildCoresPBMethodGet(req)

	case "POST":
		return r.buildCoresPBMethodPost(rw, req)
	}

	return nil, errors.New("invalid method")
}

func (r *Call) buildCoresPBMethodGet(req *http.Request) (*corespb.Request, error) {
	urlValues := req.URL.Query()
	request := &corespb.Request{
		Header:  make(map[string]string),
		Payload: make([]byte, 0),
	}

	if len(req.URL.RawQuery) > r.maxQureyLength {
		return nil, errors.New("URL query too long")
	}

	values := make(map[string]string)
	for k, v := range urlValues {
		values[k] = strings.Join(v, ";")
	}

	request.Header = values
	request.Payload = make([]byte, 0)
	return request, nil
}

func (r *Call) buildCoresPBMethodPost(rw http.ResponseWriter, req *http.Request) (*corespb.Request, error) {
	urlValues := req.URL.Query()
	request := &corespb.Request{
		Header:  make(map[string]string),
		Payload: make([]byte, 0),
	}

	if len(req.URL.RawQuery) > r.maxQureyLength {
		return nil, errors.New("URL query too long")
	}

	values := make(map[string]string)
	for k, v := range urlValues {
		values[k] = strings.Join(v, ";")
	}

	request.Header = values
	req.Body = http.MaxBytesReader(rw, req.Body, r.maxBytesReader)
	defer req.Body.Close()

	switch r.resolveContentType(req) {
	case "application/json", "application/x-protobuf":
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		request.Payload = body

	case "application/x-www-form-urlencoded":
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
		formValues := make(map[string]string)
		for k, v := range req.Form {
			formValues[k] = strings.Join(v, ";")
		}

		body, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}
		request.Payload = body

	default:
		return nil, errors.New("mime not supported")
	}

	return request, nil
}

// NewCall 创建Call路由
func NewCall(config conf.RPCClient, logger coreslog.Logger, opts ...CallOptions) *Call {
	c := &Call{
		c:              config,
		commandSecret:  make([]byte, 0),
		maxQureyLength: defaultMaxQueryLength,
		maxBytesReader: defaultMaxBytesReader,
		logger:         logger,
		pool:           make(map[string]*grpcpool.Pool),
	}

	for _, o := range opts {
		o(c)
	}
	return c
}

// CommandSecretCallOptions 设置Command解决key
func CommandSecretCallOptions(secret string) CallOptions {
	return func(r *Call) {
		r.commandSecret = []byte(secret)
	}
}

// MaxQureyLengthCallOptions 设置最大http query 长度
func MaxQureyLengthCallOptions(num int) CallOptions {
	return func(r *Call) {
		r.maxQureyLength = num
	}
}

// MaxBytesReaderCallOptions 设置最大body size
func MaxBytesReaderCallOptions(num int64) CallOptions {
	return func(r *Call) {
		r.maxBytesReader = num
	}
}
