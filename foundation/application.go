package foundation

import (
	"os"
	"strings"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support"
)

var App foundation.Application

func init() {
	setEnv()

	app := &Application{Container: NewContainer()}
	app.registerBaseServiceProviders()
	app.bootBaseServiceProviders()

	App = app
}

type Application struct {
	foundation.Container
}

func NewApplication() foundation.Application {
	return App
}

//Boot Register and bootstrap configured service providers.
func (app *Application) Boot() {
	app.registerConfiguredServiceProviders()
	app.bootConfiguredServiceProviders()

	app.bootArtisan()
	setRootPath()
}

//bootArtisan Boot artisan command.
func (app *Application) bootArtisan() {
	app.MakeArtisan().Run(os.Args, true)
}

//getBaseServiceProviders Get base service providers.
func (app *Application) getBaseServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&config.ServiceProvider{},
	}
}

//getConfiguredServiceProviders Get configured service providers.
func (app *Application) getConfiguredServiceProviders() []foundation.ServiceProvider {
	return app.MakeConfig().Get("app.providers").([]foundation.ServiceProvider)
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
func (app *Application) registerServiceProviders(serviceProviders []foundation.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Register(app)
	}
}

//bootServiceProviders Bootstrap service providers.
func (app *Application) bootServiceProviders(serviceProviders []foundation.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Boot(app)
	}
}

func setEnv() {
	args := os.Args
	if strings.HasSuffix(os.Args[0], ".test") {
		support.Env = support.EnvTest
	}
	if len(args) >= 2 {
		if args[1] == "artisan" {
			support.Env = support.EnvArtisan
		}
	}
}

func setRootPath() {
	rootPath := getCurrentAbsolutePath()

	// Hack air path
	airPath := "/storage/temp"
	if strings.HasSuffix(rootPath, airPath) {
		rootPath = strings.ReplaceAll(rootPath, airPath, "")
	}

	support.RootPath = rootPath
}
