package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

type Grpc interface {
	// Client gets the gRPC client instance.
	Client(ctx context.Context, name string) (*grpc.ClientConn, error)
	// Listen starts the gRPC server with the given listener.
	Listen(l net.Listener) error
	// Run starts the gRPC server.
	Run(host ...string) error
	// Server gets the gRPC server instance.
	Server() *grpc.Server
	// Shutdown stops the gRPC server.
	Shutdown(force ...bool)
	// UnaryServerInterceptors sets the gRPC server interceptors.
	UnaryServerInterceptors([]grpc.UnaryServerInterceptor)
	// UnaryClientInterceptorGroups sets the gRPC client interceptor groups.
	UnaryClientInterceptorGroups(map[string][]grpc.UnaryClientInterceptor)
}
