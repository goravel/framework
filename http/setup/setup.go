package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	httpFacade := "Http"
	rateLimiterFacade := "RateLimiter"
	viewFacade := "View"
	providersBootstrapPath := path.Bootstrap("providers.go")
	httpConfigPath := path.Config("http.go")
	jwtConfigPath := path.Config("jwt.go")
	corsConfigPath := path.Config("cors.go")
	httpFacadePath := path.Facades("http.go")
	rateLimiterFacadePath := path.Facades("rate_limiter.go")
	viewFacadePath := path.Facades("view.go")
	kernelPath := path.App("http", "kernel.go")
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	httpServiceProvider := "&http.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			// Add the HTTP service provider to the providers array in config/app.go
			modify.GoFile(providersBootstrapPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(httpServiceProvider)),

			// Create config/http.go, config/jwt.go, config/cors.go, app/http/kernel.go
			modify.File(httpConfigPath).Overwrite(stubs.HttpConfig(moduleName)),
			modify.File(jwtConfigPath).Overwrite(stubs.JwtConfig(moduleName)),
			modify.File(corsConfigPath).Overwrite(stubs.CorsConfig(moduleName)),
			modify.File(kernelPath).Overwrite(stubs.Kernel()),

			// Register the Http, RateLimiter, View facades
			modify.WhenFacade(httpFacade, modify.File(httpFacadePath).Overwrite(stubs.HttpFacade())),
			modify.WhenFacade(rateLimiterFacade, modify.File(rateLimiterFacadePath).Overwrite(stubs.RateLimiterFacade())),
			modify.WhenFacade(viewFacade, modify.File(viewFacadePath).Overwrite(stubs.ViewFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{httpFacade, rateLimiterFacade, viewFacade},
				// Remove the HTTP service provider from the providers array in config/app.go
				modify.GoFile(providersBootstrapPath).
					Find(match.Providers()).Modify(modify.Unregister(httpServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),

				// Remove config/http.go, config/jwt.go, config/cors.go, app/http/kernel.go
				modify.File(httpConfigPath).Remove(),
				modify.File(jwtConfigPath).Remove(),
				modify.File(corsConfigPath).Remove(),
				modify.File(kernelPath).Remove(),
			),

			// Remove the Http, RateLimiter, View facades
			modify.WhenFacade(httpFacade, modify.File(httpFacadePath).Remove()),
			modify.WhenFacade(rateLimiterFacade, modify.File(rateLimiterFacadePath).Remove()),
			modify.WhenFacade(viewFacade, modify.File(viewFacadePath).Remove()),
		).
		Execute()
}
