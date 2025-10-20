package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	appConfigPath := path.Config("app.go")
	sessionConfigPath := path.Config("session.go")
	sessionFacadePath := path.Facades("session.go")
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	sessionServiceProvider := "&session.ServiceProvider{}"
	envPath := path.Base(".env")
	envExamplePath := path.Base(".env.example")
	env := `
SESSION_DRIVER=file
SESSION_LIFETIME=120
`

	packages.Setup(os.Args).
		Install(
			// Add the session service provider to the providers array in config/app.go
			modify.GoFile(appConfigPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(sessionServiceProvider)),

			// Create config/session.go and the Session facade
			modify.File(sessionConfigPath).Overwrite(stubs.Config(moduleName)),

			// Add the Session facade
			modify.WhenFacade(facades.Session, modify.File(sessionFacadePath).Overwrite(stubs.SessionFacade())),

			// Add configurations to the .env and .env.example files
			modify.WhenFileNotContains(envPath, "SESSION_DRIVER", modify.File(envPath).Append(env)),
			modify.WhenFileNotContains(envExamplePath, "SESSION_DRIVER", modify.File(envExamplePath).Append(env)),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Session},
				// Remove the session service provider from the providers array in config/app.go
				modify.GoFile(appConfigPath).
					Find(match.Providers()).Modify(modify.Unregister(sessionServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),

				// Remove config/session.go
				modify.File(sessionConfigPath).Remove(),
			),

			// Remove the Session facade
			modify.WhenFacade(facades.Session, modify.File(sessionFacadePath).Remove()),
		).
		Execute()
}
