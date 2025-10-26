package main

import (
	"fmt"
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	routeFacadePath := path.Facades("route.go")
	providersBootstrapPath := path.Bootstrap("providers.go")
	appServiceProviderPath := path.App("providers", "app_service_provider.go")
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	globalMiddleware := "facades.Route().GlobalMiddleware(http.Kernel{}.Middleware()...)"
	facadesImport := fmt.Sprintf("%s/app/facades", moduleName)
	httpImport := fmt.Sprintf("%s/app/http", moduleName)
	routesImport := fmt.Sprintf("%s/routes", moduleName)
	routesWeb := "routes.Web()"
	routesPath := path.Base("routes", "web.go")
	welcomeTmplPath := path.Base("resources", "views", "welcome.tmpl")
	routeServiceProvider := "&route.ServiceProvider{}"
	envPath := path.Base(".env")
	envExamplePath := path.Base(".env.example")
	env := `
APP_URL=http://localhost
APP_HOST=127.0.0.1
APP_PORT=3000

JWT_SECRET=
`

	packages.Setup(os.Args).
		Install(
			// Add the route service provider to the providers array in config/app.go
			modify.GoFile(providersBootstrapPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(routeServiceProvider)),

			// Create resources/views/welcome.tmpl and routes/web.go
			modify.File(welcomeTmplPath).Overwrite(stubs.WelcomeTmpl()),
			modify.File(routesPath).Overwrite(stubs.Routes(moduleName)),

			// Modify app/providers/app_service_provider.go to register the HTTP global middleware
			modify.GoFile(appServiceProviderPath).
				Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
				Find(match.Imports()).Modify(modify.AddImport(httpImport)).
				Find(match.Imports()).Modify(modify.AddImport(routesImport)).
				Find(match.BootFunc()).Modify(modify.Add(globalMiddleware)).
				Find(match.BootFunc()).Modify(modify.Add(routesWeb)),

			// Register the Route facade
			modify.WhenFacade(facades.Route, modify.File(routeFacadePath).Overwrite(stubs.RouteFacade())),

			// Add configurations to the .env and .env.example files
			modify.WhenFileNotContains(envPath, "APP_URL", modify.File(envPath).Append(env)),
			modify.WhenFileNotContains(envExamplePath, "APP_URL", modify.File(envExamplePath).Append(env)),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Route},
				// Modify app/providers/app_service_provider.go to unregister the HTTP global middleware
				modify.GoFile(appServiceProviderPath).
					Find(match.BootFunc()).Modify(modify.Remove(globalMiddleware)).
					Find(match.BootFunc()).Modify(modify.Remove(routesWeb)).
					Find(match.Imports()).Modify(modify.RemoveImport(facadesImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(httpImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(routesImport)),

				// Remove resources/views/welcome.tmpl and routes/web.go
				modify.File(routesPath).Remove(),
				modify.File(welcomeTmplPath).Remove(),

				// Remove the route service provider from the providers array in config/app.go
				modify.GoFile(providersBootstrapPath).
					Find(match.Providers()).Modify(modify.Unregister(routeServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),

			// Remove the Route facade
			modify.WhenFacade(facades.Route, modify.File(routeFacadePath).Remove()),
		).
		Execute()
}
