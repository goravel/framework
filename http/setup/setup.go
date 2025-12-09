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
	packageName := setup.PackageName()
	httpServiceProvider := "&http.ServiceProvider{}"
	modulePath := setup.ModulePath()

	setup.Install(
		// Add the http service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(modulePath, httpServiceProvider),

		// Create config/http.go, config/jwt.go, config/cors.go
		modify.File(httpConfigPath).Overwrite(stubs.HttpConfig(packageName)),
		modify.File(jwtConfigPath).Overwrite(stubs.JwtConfig(packageName)),
		modify.File(corsConfigPath).Overwrite(stubs.CorsConfig(packageName)),

		// Register the Http, RateLimiter, View facades
		modify.WhenFacade(httpFacade, modify.File(httpFacadePath).Overwrite(stubs.HttpFacade())),
		modify.WhenFacade(rateLimiterFacade, modify.File(rateLimiterFacadePath).Overwrite(stubs.RateLimiterFacade())),
		modify.WhenFacade(viewFacade, modify.File(viewFacadePath).Overwrite(stubs.ViewFacade())),
	).Uninstall(
		modify.WhenNoFacades([]string{httpFacade, rateLimiterFacade, viewFacade},
			// Remove config/http.go, config/jwt.go, config/cors.go
			modify.File(httpConfigPath).Remove(),
			modify.File(jwtConfigPath).Remove(),
			modify.File(corsConfigPath).Remove(),

			// Remove the http service provider from the providers array in bootstrap/providers.go
			modify.RemoveProviderApply(modulePath, httpServiceProvider),
		),

		// Remove the Http, RateLimiter, View facades
		modify.WhenFacade(httpFacade, modify.File(httpFacadePath).Remove()),
		modify.WhenFacade(rateLimiterFacade, modify.File(rateLimiterFacadePath).Remove()),
		modify.WhenFacade(viewFacade, modify.File(viewFacadePath).Remove()),
	).Execute()
}
