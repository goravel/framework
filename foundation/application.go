package foundation

import (
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/facades"
	"os"
)

func init() {
	//Create a new application instance.
	app := Application{}

	app.registerBaseServiceProviders()
	app.bootBaseServiceProviders()
}

const Version string = "0.1.2"
const EnvironmentFile string = ".env"

type Application struct {
}

//Boot Register and bootstrap configured service providers.
func (app *Application) Boot() {
	app.registerConfiguredServiceProviders()
	app.bootConfiguredServiceProviders()

	app.bootArtisan()
}

//bootArtisan Boot artisan command.
func (app *Application) bootArtisan() {
	facades.Artisan.Run(os.Args)
}

//getBaseServiceProviders Get base service providers.
func (app *Application) getBaseServiceProviders() []support.ServiceProvider {
	return []support.ServiceProvider{
		&config.ServiceProvider{},
	}
}

//getConfiguredServiceProviders Get configured service providers.
func (app *Application) getConfiguredServiceProviders() []support.ServiceProvider {
	return facades.Config.Get("app.providers").([]support.ServiceProvider)
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
func (app *Application) registerServiceProviders(serviceProviders []support.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Register()
	}
}

//bootServiceProviders Bootstrap service providers.
func (app *Application) bootServiceProviders(serviceProviders []support.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Boot()
	}
}

//EnvironmentFile Get the environment file the application is using.
func (app *Application) EnvironmentFile() string {
	return EnvironmentFile
}

//RunningInConsole Determine if the application is running in the console.
func (app *Application) RunningInConsole() bool {
	args := os.Args

	return len(args) > 2 && args[1] == "artisan"
}
