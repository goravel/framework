package foundation

import (
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	"github.com/goravel/framework/database"
	"github.com/goravel/framework/route"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/facades"
)

func init() {
	//Create a new application instance.
	app := Application{}

	app.registerBaseServiceProviders()
	app.bootBaseServiceProviders()
}

type Application struct {
}

//Boot Register and bootstrap all of the configured service providers.
func (app *Application) Boot() {
	app.registerConfiguredServiceProviders()
	app.bootConfiguredServiceProviders()
}

//BootHttpKernel Bootstrap the http kernel, add http middlewares.
func (app *Application) BootHttpKernel(kernel support.Kernel) {
	facades.Route.Use(kernel.Middleware()...)
}

//getBaseServiceProviders Get all of the base service providers.
func (app *Application) getBaseServiceProviders() []support.ServiceProvider {
	return []support.ServiceProvider{
		&config.ServiceProvider{},
	}
}

//getConfiguredServiceProviders Get all of the configured service providers.
func (app *Application) getConfiguredServiceProviders() []support.ServiceProvider {
	configuredServiceProviders := []support.ServiceProvider{
		&database.ServiceProvider{},
		&console.ServiceProvider{},
		&route.ServiceProvider{},
	}

	configuredServiceProviders = append(configuredServiceProviders, facades.Config.Get("app.providers").([]support.ServiceProvider)...)

	return configuredServiceProviders
}

//registerBaseServiceProviders Register all of the base service providers.
func (app *Application) registerBaseServiceProviders() {
	app.registerServiceProviders(app.getBaseServiceProviders())
}

//bootBaseServiceProviders Bootstrap all of the base service providers.
func (app *Application) bootBaseServiceProviders() {
	app.bootServiceProviders(app.getBaseServiceProviders())
}

//registerConfiguredServiceProviders Register all of the configured service providers.
func (app *Application) registerConfiguredServiceProviders() {
	app.registerServiceProviders(app.getConfiguredServiceProviders())
}

//bootConfiguredServiceProviders Bootstrap all of the configured service providers.
func (app *Application) bootConfiguredServiceProviders() {
	app.bootServiceProviders(app.getConfiguredServiceProviders())
}

//registerServiceProviders Register service providers.
func (app *Application) registerServiceProviders(serviceProviders []support.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		app.register(serviceProvider)
	}
}

//bootServiceProviders Bootstrap service providers.
func (app *Application) bootServiceProviders(serviceProviders []support.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		app.boot(serviceProvider)
	}
}

//register Register a service provider.
func (app *Application) register(serviceProvider support.ServiceProvider) {
	serviceProvider.Register()
}

//boot Bootstrap a service provider.
func (app *Application) boot(serviceProvider support.ServiceProvider) {
	serviceProvider.Boot()
}
