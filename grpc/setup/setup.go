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
	appServiceProviderPath := path.App("providers", "app_service_provider.go")
	kernelPath := path.App("grpc", "kernel.go")
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	grpcServiceProvider := "&grpc.ServiceProvider{}"
	unaryServerInterceptors := "facades.Grpc().UnaryServerInterceptors(grpc.Kernel{}.UnaryServerInterceptors())"
	unaryClientInterceptorGroups := "facades.Grpc().UnaryClientInterceptorGroups(grpc.Kernel{}.UnaryClientInterceptorGroups())"
	facadesImport := fmt.Sprintf("%s/app/facades", moduleName)
	grpcImport := fmt.Sprintf("%s/app/grpc", moduleName)

	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(grpcServiceProvider)),
			modify.File(path.Config("grpc.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.File(kernelPath).Overwrite(stubs.Kernel()),
			modify.GoFile(appServiceProviderPath).
				Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
				Find(match.Imports()).Modify(modify.AddImport(grpcImport)).
				Find(match.RegisterFunc()).Modify(modify.Add(unaryServerInterceptors)).
				Find(match.RegisterFunc()).Modify(modify.Add(unaryClientInterceptorGroups)),
			modify.WhenFacade("Grpc", modify.File(path.Facades("grpc.go")).Overwrite(stubs.GrpcFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Grpc"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister(grpcServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.GoFile(appServiceProviderPath).
					Find(match.RegisterFunc()).Modify(modify.Remove(unaryServerInterceptors)).
					Find(match.RegisterFunc()).Modify(modify.Remove(unaryClientInterceptorGroups)).
					Find(match.Imports()).Modify(modify.RemoveImport(facadesImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(grpcImport)),
				modify.File(path.Config("grpc.go")).Remove(),
				modify.File(kernelPath).Remove(),
			),
			modify.WhenFacade("Grpc", modify.File(path.Facades("grpc.go")).Remove()),
		).
		Execute()
}
