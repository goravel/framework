package foundation

import (
	"os"
	"strings"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support"
)

func init() {
	setEnv()

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
	setRootPath()
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
	rootPath := getCurrentAbPath()

	// Hack air path
	airPath := "/storage/temp"
	if strings.HasSuffix(rootPath, airPath) {
		rootPath = strings.ReplaceAll(rootPath, airPath, "")
	}

	support.RootPath = rootPath
}
