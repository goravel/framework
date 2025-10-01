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
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	globalMiddleware := "facades.Route().GlobalMiddleware(http.Kernel{}.Middleware()...)"
	facadesImport := fmt.Sprintf("%s/app/facades", moduleName)
	httpImport := fmt.Sprintf("%s/app/http", moduleName)
	routesImport := fmt.Sprintf("%s/routes", moduleName)
	routesWeb := "routes.Web()"
	routesPath := path.Base("routes", "web.go")

	packages.Setup(os.Args).
		Install(
			// Add the route service provider to the providers array in config/app.go
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&route.ServiceProvider{}")),

			// Create routes/web.go
			modify.File(routesPath).Overwrite(stubs.Routes()),

			// Modify app/providers/app_service_provider.go to register the HTTP global middleware
			modify.GoFile(appServiceProviderPath).
				Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
				Find(match.Imports()).Modify(modify.AddImport(httpImport)).
				Find(match.Imports()).Modify(modify.AddImport(routesImport)).
				Find(match.BootFunc()).Modify(modify.Add(globalMiddleware)).
				Find(match.BootFunc()).Modify(modify.Add(routesWeb)),

			// Register the Route facade
			modify.WhenFacade("Route", modify.File(path.Facades("route.go")).Overwrite(stubs.RouteFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Route"},
				// Modify app/providers/app_service_provider.go to unregister the HTTP global middleware
				modify.GoFile(appServiceProviderPath).
					Find(match.BootFunc()).Modify(modify.Remove(globalMiddleware)).
					Find(match.BootFunc()).Modify(modify.Remove(routesWeb)).
					Find(match.Imports()).Modify(modify.RemoveImport(facadesImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(httpImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(routesImport)),

				// Remove routes/web.go
				modify.File(routesPath).Remove(),

				// Remove the route service provider from the providers array in config/app.go
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&route.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),

			// Remove the Route facade
			modify.WhenFacade("Route", modify.File(path.Facades("route.go")).Remove()),
		).
		Execute()
}
