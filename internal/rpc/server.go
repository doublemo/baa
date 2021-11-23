package rpc

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/doublemo/baa/internal/conf"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// basepath is the root directory of this package.
var basepath string

var (
	ErrMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	ErrInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

// NewServer 创建GRPC server
func NewServer(c conf.RPC) (*grpc.Server, error) {
	opts := []grpc.ServerOption{}
	if len(c.Key) > 0 && len(c.Salt) > 0 {
		cert, err := tls.LoadX509KeyPair(c.Salt, c.Key)
		if err != nil {
			return nil, err
		}

		opts = append(opts,
			grpc.ChainStreamInterceptor(ensureStreamValidToken(c), grpc_prometheus.StreamServerInterceptor),
			grpc.ChainUnaryInterceptor(ensureValidToken(c), grpc_prometheus.UnaryServerInterceptor),
			grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
		)
	}

	return grpc.NewServer(opts...), nil
}

func ensureValidToken(config conf.RPC) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrMissingMetadata
		}
		// The keys within metadata.MD are normalized to lowercase.
		// See: https://godoc.org/google.golang.org/grpc/metadata#New
		authorization := md["authorization"]
		if len(authorization) < 1 {
			return nil, ErrInvalidToken
		}

		token := strings.TrimPrefix(authorization[0], "Bearer ")
		if token != config.ServiceSecurityKey {
			return nil, ErrInvalidToken
		}

		// Continue execution of handler after ensuring a valid token.
		return handler(ctx, req)
	}
}

//func(srv interface{}, ss ServerStream, info *StreamServerInfo, handler StreamHandler)
func ensureStreamValidToken(config conf.RPC) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return ErrMissingMetadata
		}

		authorization := md["authorization"]
		if len(authorization) < 1 {
			return ErrInvalidToken
		}

		token := strings.TrimPrefix(authorization[0], "Bearer ")
		if token != config.ServiceSecurityKey {
			return ErrInvalidToken
		}

		// Continue execution of handler after ensuring a valid token.
		return handler(srv, ss)
	}
}
