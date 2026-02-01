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
	sessionConfigPath := path.Config("session.go")
	sessionFacadePath := path.Facade("session.go")
	sessionServiceProvider := "&session.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()
	envPath := path.Base(".env")
	envExamplePath := path.Base(".env.example")
	env := `
SESSION_DRIVER=file
SESSION_LIFETIME=120
`

	setup.Install(
		// Add the session service provider to the providers array in bootstrap/providers.go
		modify.RegisterProvider(moduleImport, sessionServiceProvider),

		// Create config/session.go and the Session facade
		modify.File(sessionConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

		// Add the Session facade
		modify.File(sessionFacadePath).Overwrite(stubs.SessionFacade(facadesPackage)),

		// Add configurations to the .env and .env.example files
		modify.WhenFileNotContains(envPath, "SESSION_DRIVER", modify.File(envPath).Append(env)),
		modify.WhenFileNotContains(envExamplePath, "SESSION_DRIVER", modify.File(envExamplePath).Append(env)),
	).Uninstall(
		// Remove config/session.go
		modify.File(sessionConfigPath).Remove(),

		// Remove the session service provider from the providers array in bootstrap/providers.go
		modify.UnregisterProvider(moduleImport, sessionServiceProvider),

		// Remove the Session facade
		modify.File(sessionFacadePath).Remove(),
	).Execute()
}
