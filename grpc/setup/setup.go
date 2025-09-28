package main

import (
	"fmt"
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	grpcFacade := "Grpc"
	appServiceProviderPath := path.App("providers", "app_service_provider.go")
	appConfigPath := path.Config("app.go")
	configPath := path.Config("grpc.go")
	facadePath := path.Facades("grpc.go")
	kernelPath := path.App("grpc", "kernel.go")
	routesPath := path.Base("routes", "grpc.go")
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	grpcServiceProvider := "&grpc.ServiceProvider{}"
	unaryServerInterceptors := "facades.Grpc().UnaryServerInterceptors(grpc.Kernel{}.UnaryServerInterceptors())"
	unaryClientInterceptorGroups := "facades.Grpc().UnaryClientInterceptorGroups(grpc.Kernel{}.UnaryClientInterceptorGroups())"
	routesGrpc := "routes.Grpc()"
	facadesImport := fmt.Sprintf("%s/app/facades", moduleName)
	grpcImport := fmt.Sprintf("%s/app/grpc", moduleName)
	routesImport := fmt.Sprintf("%s/routes", moduleName)

	packages.Setup(os.Args).
		Install(
			modify.GoFile(appConfigPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(grpcServiceProvider)),
			modify.File(configPath).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.File(routesPath).Overwrite(stubs.Routes()),
			modify.File(kernelPath).Overwrite(stubs.Kernel()),
			modify.GoFile(appServiceProviderPath).
				Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
				Find(match.Imports()).Modify(modify.AddImport(grpcImport)).
				Find(match.Imports()).Modify(modify.AddImport(routesImport)).
				Find(match.RegisterFunc()).Modify(modify.Add(unaryServerInterceptors)).
				Find(match.RegisterFunc()).Modify(modify.Add(unaryClientInterceptorGroups)).
				Find(match.BootFunc()).Modify(modify.Add(routesGrpc)),
			modify.WhenFacade(grpcFacade, modify.File(facadePath).Overwrite(stubs.GrpcFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{grpcFacade},
				modify.GoFile(appConfigPath).
					Find(match.Providers()).Modify(modify.Unregister(grpcServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.GoFile(appServiceProviderPath).
					Find(match.RegisterFunc()).Modify(modify.Remove(unaryServerInterceptors)).
					Find(match.RegisterFunc()).Modify(modify.Remove(unaryClientInterceptorGroups)).
					Find(match.BootFunc()).Modify(modify.Remove(routesGrpc)).
					Find(match.Imports()).Modify(modify.RemoveImport(facadesImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(grpcImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(routesImport)),
				modify.File(configPath).Remove(),
				modify.File(kernelPath).Remove(),
				modify.File(routesPath).Remove(),
			),
			modify.WhenFacade(grpcFacade, modify.File(facadePath).Remove()),
		).
		Execute()
}
