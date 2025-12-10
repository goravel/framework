package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) Config(pkg, main string) string {
	content := `package DummyPackage

import (
	"DummyMain/app/facades"
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

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyMain", main)

	return content
}

func (s Stubs) GrpcFacade(pkg string) string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/grpc"
)

func Grpc() grpc.Grpc {
	return App().MakeGrpc()
}
`
}

func (s Stubs) Routes(pkg string) string {
	return `package DummyPackage

func Grpc() {

}
`
}
