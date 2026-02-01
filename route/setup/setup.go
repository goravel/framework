package main

import (
	"os"

	contractsmodify "github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	routeFacadePath := path.Facade("route.go")
	routesImport := setup.Paths().Routes().Import()
	webFunc := setup.Paths().Routes().Package() + ".Web()"
	webRoutePath := path.Route("web.go")
	jwtConfigPath := path.Config("jwt.go")
	corsConfigPath := path.Config("cors.go")
	welcomeTmplPath := path.Resource("views", "welcome.tmpl")
	controllerPath := path.App("http", "controllers", "user_controller.go")
	publicGitignorePath := path.Public(".gitignore")
	routeServiceProvider := "&route.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()
	facadesImport := setup.Paths().Facades().Import()
	configPackage := setup.Paths().Config().Package()
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
		modify.RegisterProvider(moduleImport, routeServiceProvider),

		// Create resources/views/welcome.tmpl, app/http/controllers/user_controller.go, public/.gitignore, routes/web.go, config/jwt.go, config/cors.go
		modify.File(welcomeTmplPath).Overwrite(stubs.WelcomeTmpl()),
		modify.File(controllerPath).Overwrite(stubs.Controller()),
		modify.File(publicGitignorePath).Overwrite(""),
		modify.File(webRoutePath).Overwrite(stubs.Routes(setup.Paths().Routes().Package(), setup.Paths().App().Import(), setup.Paths().Facades().Import(), facadesPackage)),
		modify.File(jwtConfigPath).Overwrite(stubs.JwtConfig(configPackage, facadesImport, facadesPackage)),
		modify.File(corsConfigPath).Overwrite(stubs.CorsConfig(configPackage, facadesImport, facadesPackage)),

		// Add the Web function to WithRouting
		modify.RegisterRoute(routesImport, webFunc),
		// Register the Route facade
		modify.File(routeFacadePath).Overwrite(stubs.RouteFacade(facadesPackage)),

		// Add configurations to the .env and .env.example files
		modify.WhenFileNotContains(envPath, "APP_URL", modify.File(envPath).Append(env)),
		modify.WhenFileNotContains(envExamplePath, "APP_URL", modify.File(envExamplePath).Append(env)),
	).Uninstall(
		// Remove the Route facade
		modify.File(routeFacadePath).Remove(),

		// Remove the Web function from WithRouting
		modify.UnregisterRoute(routesImport, webFunc),

		// Remove resources folder, app folder, public folder, routes/web.go, config/jwt.go, config/cors.go
		modify.File(webRoutePath).Remove(),
		modify.Call(func(options []contractsmodify.Option) error {
			if err := file.Remove(path.App("http")); err != nil {
				return err
			}
			if err := file.Remove(path.Public()); err != nil {
				return err
			}
			if err := file.Remove(path.Resource()); err != nil {
				return err
			}

			return nil
		}),
		modify.File(jwtConfigPath).Remove(),
		modify.File(corsConfigPath).Remove(),

		// Remove the route service provider from the providers array in bootstrap/providers.go
		modify.UnregisterProvider(moduleImport, routeServiceProvider),
	).Execute()
}
