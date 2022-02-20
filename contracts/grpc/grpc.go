package grpc

import "google.golang.org/grpc"

type Grpc interface {
	Run(host string) error
	Server() *grpc.Server
	SetServer(server *grpc.Server)
}
