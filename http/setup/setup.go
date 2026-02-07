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
	httpFacadePath := path.Facade("http.go")
	rateLimiterFacadePath := path.Facade("rate_limiter.go")
	viewFacadePath := path.Facade("view.go")
	httpServiceProvider := "&http.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()
	facadesImport := setup.Paths().Facades().Import()
	configPackage := setup.Paths().Config().Package()

	setup.Install(
		// Add the http service provider to the providers array in bootstrap/providers.go
		modify.WhenFileNotContains(path.Bootstrap("providers.go"), httpServiceProvider, modify.RegisterProvider(moduleImport, httpServiceProvider)),

		// Register the Http, RateLimiter, View facades
		modify.WhenFacade(httpFacade,
			// Create config/http.go
			modify.File(httpConfigPath).Overwrite(stubs.HttpConfig(configPackage, facadesImport, facadesPackage)),

			// Create the Http facade
			modify.File(httpFacadePath).Overwrite(stubs.HttpFacade(facadesPackage)),
		),
		modify.WhenFacade(rateLimiterFacade, modify.File(rateLimiterFacadePath).Overwrite(stubs.RateLimiterFacade(facadesPackage))),
		modify.WhenFacade(viewFacade, modify.File(viewFacadePath).Overwrite(stubs.ViewFacade(facadesPackage))),
	).Uninstall(
		modify.WhenNoFacades([]string{httpFacade, rateLimiterFacade, viewFacade},
			// Remove the http service provider from the providers array in bootstrap/providers.go
			modify.UnregisterProvider(moduleImport, httpServiceProvider),
		),

		// Remove the Http, RateLimiter, View facades
		modify.WhenFacade(httpFacade,
			// Remove config/http.go
			modify.File(httpConfigPath).Remove(),

			// Remove the Http facade
			modify.File(httpFacadePath).Remove(),
		),
		modify.WhenFacade(rateLimiterFacade, modify.File(rateLimiterFacadePath).Remove()),
		modify.WhenFacade(viewFacade, modify.File(viewFacadePath).Remove()),
	).Execute()
}
