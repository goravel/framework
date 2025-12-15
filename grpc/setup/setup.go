package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	routesImport := setup.Paths().Routes().Import()
	routesPackage := setup.Paths().Routes().Package()
	grpcFunc := routesPackage + ".Grpc"
	configPath := path.Config("grpc.go")
	facadePath := path.Facades("grpc.go")
	grpcRoutePath := path.Route("grpc.go")
	grpcServiceProvider := "&grpc.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()
	env := `
GRPC_HOST=
GRPC_PORT=
`

	setup.Install(
		// Add the grpc service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, grpcServiceProvider),

		// Create config/grpc.go, routes/grpc.go
		modify.File(configPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),
		modify.File(grpcRoutePath).Overwrite(stubs.Routes(setup.Paths().Routes().Package())),

		// Add the Grpc function to WithRouting
		modify.AddRouteApply(routesImport, grpcFunc),

		// Register the Grpc facade
		modify.File(facadePath).Overwrite(stubs.GrpcFacade(facadesPackage)),

		// Add configurations to the .env and .env.example files
		modify.WhenFileNotContains(path.Base(".env"), "GRPC_HOST", modify.File(path.Base(".env")).Append(env)),
		modify.WhenFileNotContains(path.Base(".env.example"), "GRPC_HOST", modify.File(path.Base(".env.example")).Append(env)),
	).Uninstall(
		// Remove the Grpc facade
		modify.File(facadePath).Remove(),

		// Remove the Grpc function from WithRouting
		modify.RemoveRouteApply(routesImport, grpcFunc),

		// Remove config/grpc.go, routes/grpc.go
		modify.File(configPath).Remove(),
		modify.File(grpcRoutePath).Remove(),

		// Remove the gRPC service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, grpcServiceProvider),
	).Execute()
}
