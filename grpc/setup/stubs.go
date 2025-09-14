package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) Config(module string) string {
	content := `package config

import (
	"DummyModule/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("grpc", map[string]any{
		// Configure your server host
		"host": config.Env("GRPC_HOST", ""),

		// Configure your server port
		"port": config.Env("GRPC_PORT", ""),

		// Configure your client host and interceptors.
		// Interceptors can be the group name of UnaryClientInterceptorGroups in app/grpc/kernel.go.
		"clients": map[string]any{
			//"user": map[string]any{
			//	"host":         config.Env("GRPC_USER_HOST", ""),
			//	"port":         config.Env("GRPC_USER_PORT", ""),
			//	"interceptors": []string{},
			//},
		},
	})
}
`

	return strings.ReplaceAll(content, "DummyModule", module)
}

func (s Stubs) GrpcFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/grpc"
)

func Grpc() grpc.Grpc {
	return App().MakeGrpc()
}
`
}
