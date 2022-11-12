package grpc

import (
	"context"

	"google.golang.org/grpc"
)

type Grpc interface {
	Run(host string) error
	Server() *grpc.Server
	Client(ctx context.Context, name string) (*grpc.ClientConn, error)
	UnaryServerInterceptors([]grpc.UnaryServerInterceptor)
	UnaryClientInterceptorGroups(map[string][]grpc.UnaryClientInterceptor)
}
