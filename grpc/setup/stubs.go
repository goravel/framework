package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) Config(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
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
	content = strings.ReplaceAll(content, "DummyFacadesImport", facadesImport)
	content = strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage)

	return content
}

func (s Stubs) GrpcFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/grpc"
)

func Grpc() grpc.Grpc {
	return App().MakeGrpc()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}

func (s Stubs) Routes(pkg string) string {
	content := `package DummyPackage

func Grpc() {

}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
