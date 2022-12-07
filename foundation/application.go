package foundation

import (
	"os"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support"
)

func init() {
	//Create a new application instance.
	app := Application{}

	app.registerBaseServiceProviders()
	app.bootBaseServiceProviders()
}

type Application struct {
}

//Boot Register and bootstrap configured service providers.
func (app *Application) Boot() {
	app.registerConfiguredServiceProviders()
	app.bootConfiguredServiceProviders()

	app.bootArtisan()
	app.setRootPath()
}

func (app *Application) setRootPath() {
	support.RootPath = getCurrentAbPath()
}

//bootArtisan Boot artisan command.
func (app *Application) bootArtisan() {
	facades.Artisan.Run(os.Args, true)
}

//getBaseServiceProviders Get base service providers.
func (app *Application) getBaseServiceProviders() []contracts.ServiceProvider {
	return []contracts.ServiceProvider{
		&config.ServiceProvider{},
	}
}

//getConfiguredServiceProviders Get configured service providers.
func (app *Application) getConfiguredServiceProviders() []contracts.ServiceProvider {
	return facades.Config.Get("app.providers").([]contracts.ServiceProvider)
}

//registerBaseServiceProviders Register base service providers.
func (app *Application) registerBaseServiceProviders() {
	app.registerServiceProviders(app.getBaseServiceProviders())
}

//bootBaseServiceProviders Bootstrap base service providers.
func (app *Application) bootBaseServiceProviders() {
	app.bootServiceProviders(app.getBaseServiceProviders())
}

//registerConfiguredServiceProviders Register configured service providers.
func (app *Application) registerConfiguredServiceProviders() {
	app.registerServiceProviders(app.getConfiguredServiceProviders())
}

//bootConfiguredServiceProviders Bootstrap configured service providers.
func (app *Application) bootConfiguredServiceProviders() {
	app.bootServiceProviders(app.getConfiguredServiceProviders())
}

//registerServiceProviders Register service providers.
func (app *Application) registerServiceProviders(serviceProviders []contracts.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Register()
	}
}

//bootServiceProviders Bootstrap service providers.
func (app *Application) bootServiceProviders(serviceProviders []contracts.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Boot()
	}
}

//RunningInConsole Determine if the application is running in the console.
func (app *Application) RunningInConsole() bool {
	args := os.Args

	return len(args) >= 2 && args[1] == "artisan"
}
