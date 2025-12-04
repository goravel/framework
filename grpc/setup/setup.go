package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	grpcFacade := "Grpc"
	configPath := path.Config("grpc.go")
	facadePath := path.Facades("grpc.go")
	kernelPath := path.App("grpc", "kernel.go")
	routesPath := path.Base("routes", "grpc.go")
	grpcServiceProvider := "&grpc.ServiceProvider{}"
	modulePath := packages.GetModulePath()
	env := `
GRPC_HOST=
GRPC_PORT=
`

	packages.Setup(os.Args).
		Install(
			// Add the grpc service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, grpcServiceProvider),

			// Create config/grpc.go, app/grpc/kernel.go, routes/grpc.go
			modify.File(configPath).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.File(routesPath).Overwrite(stubs.Routes()),
			modify.File(kernelPath).Overwrite(stubs.Kernel()),

			// Register the Grpc facade
			modify.WhenFacade(grpcFacade, modify.File(facadePath).Overwrite(stubs.GrpcFacade())),

			// Add configurations to the .env and .env.example files
			modify.WhenFileNotContains(path.Base(".env"), "GRPC_HOST", modify.File(path.Base(".env")).Append(env)),
			modify.WhenFileNotContains(path.Base(".env.example"), "GRPC_HOST", modify.File(path.Base(".env.example")).Append(env)),
		).
		Uninstall(
			modify.WhenNoFacades([]string{grpcFacade},
				// Remove the gRPC service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, grpcServiceProvider),

				// Remove config/grpc.go, app/grpc/kernel.go, routes/grpc.go
				modify.File(configPath).Remove(),
				modify.File(kernelPath).Remove(),
				modify.File(routesPath).Remove(),
			),

			// Remove the Grpc facade
			modify.WhenFacade(grpcFacade, modify.File(facadePath).Remove()),
		).
		Execute()
}
