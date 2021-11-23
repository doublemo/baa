package imf

import (
	"context"

	corespb "github.com/doublemo/baa/cores/proto/pb"
)

type Server struct {
	corespb.UnimplementedServiceServer
}

func (s *Server) Call(ctx context.Context, req *corespb.Request) (*corespb.Response, error) {
	return r.Handler(req)
}

func (s *Server) BidirectionalStreaming(stream corespb.Service_BidirectionalStreamingServer) error {
	return nil
}

func NewServer() *Server {
	return &Server{}
}
