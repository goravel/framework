package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	routeFacadePath := path.Facades("route.go")
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	routesPath := path.Base("routes", "web.go")
	welcomeTmplPath := path.Base("resources", "views", "welcome.tmpl")
	routeServiceProvider := "&route.ServiceProvider{}"
	modulePath := packages.GetModulePath()
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
			// Add the route service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, routeServiceProvider),

			// Create resources/views/welcome.tmpl and routes/web.go
			modify.File(welcomeTmplPath).Overwrite(stubs.WelcomeTmpl()),
			modify.File(routesPath).Overwrite(stubs.Routes(moduleName)),

			// Register the Route facade
			modify.WhenFacade(facades.Route, modify.File(routeFacadePath).Overwrite(stubs.RouteFacade())),

			// Add configurations to the .env and .env.example files
			modify.WhenFileNotContains(envPath, "APP_URL", modify.File(envPath).Append(env)),
			modify.WhenFileNotContains(envExamplePath, "APP_URL", modify.File(envExamplePath).Append(env)),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Route},
				// Remove resources/views/welcome.tmpl and routes/web.go
				modify.File(routesPath).Remove(),
				modify.File(welcomeTmplPath).Remove(),

				// Remove the route service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, routeServiceProvider),
			),

			// Remove the Route facade
			modify.WhenFacade(facades.Route, modify.File(routeFacadePath).Remove()),
		).
		Execute()
}
