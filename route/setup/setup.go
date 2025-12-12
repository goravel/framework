package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	routesPath := support.Config.Paths.Routes
	routeFacadePath := path.Facades("route.go")
	packageName := setup.Paths().Main().Package()
	routesPackage := packageName + "/" + routesPath
	webFunc := routesPath + ".Web"
	webRoutePath := path.Base(routesPath, "web.go")
	welcomeTmplPath := path.Base(support.Config.Paths.Resources, "views", "welcome.tmpl")
	routeServiceProvider := "&route.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	envPath := path.Base(".env")
	envExamplePath := path.Base(".env.example")
	env := `
APP_URL=http://localhost
APP_HOST=127.0.0.1
APP_PORT=3000

JWT_SECRET=
`

	setup.Install(
		// Add the route service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, routeServiceProvider),

		// Create resources/views/welcome.tmpl and routes/web.go
		modify.File(welcomeTmplPath).Overwrite(stubs.WelcomeTmpl()),
		modify.File(webRoutePath).Overwrite(stubs.Routes(setup.Paths().Routes().Package(), packageName)),

		// Add the Web function to WithRouting
		modify.AddRouteApply(routesPackage, webFunc),

		// Register the Route facade
		modify.File(routeFacadePath).Overwrite(stubs.RouteFacade(setup.Paths().Facades().Package())),

		// Add configurations to the .env and .env.example files
		modify.WhenFileNotContains(envPath, "APP_URL", modify.File(envPath).Append(env)),
		modify.WhenFileNotContains(envExamplePath, "APP_URL", modify.File(envExamplePath).Append(env)),
	).Uninstall(
		// Remove the Route facade
		modify.File(routeFacadePath).Remove(),

		// Remove the Web function from WithRouting
		modify.RemoveRouteApply(routesPackage, webFunc),

		// Remove resources/views/welcome.tmpl and routes/web.go
		modify.File(webRoutePath).Remove(),
		modify.File(welcomeTmplPath).Remove(),

		// Remove the route service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, routeServiceProvider),
	).Execute()
}
