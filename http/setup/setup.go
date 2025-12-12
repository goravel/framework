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
	httpFacade := "Http"
	rateLimiterFacade := "RateLimiter"
	viewFacade := "View"
	httpConfigPath := path.Config("http.go")
	jwtConfigPath := path.Config("jwt.go")
	corsConfigPath := path.Config("cors.go")
	httpFacadePath := path.Facades("http.go")
	rateLimiterFacadePath := path.Facades("rate_limiter.go")
	viewFacadePath := path.Facades("view.go")
	packageName := setup.Paths().Main().Package()
	httpServiceProvider := "&http.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	configPackage := setup.Paths().Config().Package()
	facadesPackage := setup.Paths().Facades().Package()

	setup.Install(
		// Add the http service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, httpServiceProvider),

		// Create config/http.go, config/jwt.go, config/cors.go
		modify.File(httpConfigPath).Overwrite(stubs.HttpConfig(configPackage, packageName)),
		modify.File(jwtConfigPath).Overwrite(stubs.JwtConfig(configPackage, packageName)),
		modify.File(corsConfigPath).Overwrite(stubs.CorsConfig(configPackage, packageName)),

		// Register the Http, RateLimiter, View facades
		modify.WhenFacade(httpFacade, modify.File(httpFacadePath).Overwrite(stubs.HttpFacade(facadesPackage))),
		modify.WhenFacade(rateLimiterFacade, modify.File(rateLimiterFacadePath).Overwrite(stubs.RateLimiterFacade(facadesPackage))),
		modify.WhenFacade(viewFacade, modify.File(viewFacadePath).Overwrite(stubs.ViewFacade(facadesPackage))),
	).Uninstall(
		modify.WhenNoFacades([]string{httpFacade, rateLimiterFacade, viewFacade},
			// Remove config/http.go, config/jwt.go, config/cors.go
			modify.File(httpConfigPath).Remove(),
			modify.File(jwtConfigPath).Remove(),
			modify.File(corsConfigPath).Remove(),

			// Remove the http service provider from the providers array in bootstrap/providers.go
			modify.RemoveProviderApply(moduleImport, httpServiceProvider),
		),

		// Remove the Http, RateLimiter, View facades
		modify.WhenFacade(httpFacade, modify.File(httpFacadePath).Remove()),
		modify.WhenFacade(rateLimiterFacade, modify.File(rateLimiterFacadePath).Remove()),
		modify.WhenFacade(viewFacade, modify.File(viewFacadePath).Remove()),
	).Execute()
}
