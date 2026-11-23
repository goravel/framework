package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	contractspackages "github.com/goravel/framework/contracts/packages"
	contractsmodify "github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)

	setup.Install(installActions(setup)...).Uninstall(uninstallActions(setup)...).Execute()
}

func installActions(setup contractspackages.Setup) []contractsmodify.Apply {
	stubs := Stubs{}
	sessionConfigPath := path.Config("session.go")
	sessionFacadePath := path.Facade("session.go")
	sessionServiceProvider := "&session.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()
	envPath := path.Base(".env")
	envExamplePath := path.Base(".env.example")

	return []contractsmodify.Apply{
		// Avoid duplicate installation when installing drivers
		modify.WhenFacade(facades.Session,
			// Add the session service provider to the providers array in bootstrap/providers.go
			modify.RegisterProvider(moduleImport, sessionServiceProvider),

			// Create config/session.go and the Session facade
			modify.File(sessionConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

			// Add the Session facade
			modify.File(sessionFacadePath).Overwrite(stubs.SessionFacade(facadesPackage)),

			// Add the session driver and lifetime to the .env and .env.example files
			modify.WhenFileExists(envPath, modify.EnvWhenMissing(envPath, "SESSION_DRIVER", "file")),
			modify.WhenFileExists(envPath, modify.EnvWhenMissing(envPath, "SESSION_LIFETIME", "120")),
			modify.WhenFileExists(envExamplePath, modify.EnvWhenMissing(envExamplePath, "SESSION_DRIVER", "file")),
			modify.WhenFileExists(envExamplePath, modify.EnvWhenMissing(envExamplePath, "SESSION_LIFETIME", "120")),
		),

		// Selecting the file driver overwrites the selector, so a stale value cannot survive.
		modify.WhenDriver("file",
			modify.WhenFileExists(envPath, modify.Env(envPath, "SESSION_DRIVER", "file")),
			modify.WhenFileExists(envExamplePath, modify.Env(envExamplePath, "SESSION_DRIVER", "file")),
		),
	}
}

func uninstallActions(setup contractspackages.Setup) []contractsmodify.Apply {
	sessionConfigPath := path.Config("session.go")
	sessionFacadePath := path.Facade("session.go")
	sessionServiceProvider := "&session.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	return []contractsmodify.Apply{
		modify.WhenFacade(facades.Session,
			// Remove config/session.go
			modify.File(sessionConfigPath).Remove(),

			// Remove the session service provider from the providers array in bootstrap/providers.go
			modify.UnregisterProvider(moduleImport, sessionServiceProvider),

			// Remove the Session facade
			modify.File(sessionFacadePath).Remove(),
		),
	}
}
