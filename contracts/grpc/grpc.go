package grpc

import (
	"context"

	"google.golang.org/grpc"
)

//go:generate mockery --name=Grpc
type Grpc interface {
	// Run starts the gRPC server.
	Run(host ...string) error
	// Server gets the gRPC server instance.
	Server() *grpc.Server
	// Client gets the gRPC client instance.
	Client(ctx context.Context, name string) (*grpc.ClientConn, error)
	// UnaryServerInterceptors sets the gRPC server interceptors.
	UnaryServerInterceptors([]grpc.UnaryServerInterceptor)
	// UnaryClientInterceptorGroups sets the gRPC client interceptor groups.
	UnaryClientInterceptorGroups(map[string][]grpc.UnaryClientInterceptor)
}
