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
	aiConfigPath := path.Config("ai.go")
	aiFacadePath := path.Facade("ai.go")
	modulePath := setup.Paths().Module().Import()
	aiServiceProvider := "&ai.ServiceProvider{}"
	facadesPackage := setup.Paths().Facades().Package()
	env := `
AI_PROVIDER=
`

	setup.Install(
		// Add the ai service provider to the providers array in bootstrap/providers.go
		modify.RegisterProvider(modulePath, aiServiceProvider),

		// Create config/ai.go
		modify.File(aiConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

		// Add the AI facade
		modify.File(aiFacadePath).Overwrite(stubs.AIFacade(facadesPackage)),

		// Add configurations to the .env and .env.example files
		modify.WhenFileExists(path.Base(".env"), modify.WhenFileNotContains(path.Base(".env"), "AI_PROVIDER", modify.File(path.Base(".env")).Append(env))),
		modify.WhenFileExists(path.Base(".env.example"), modify.WhenFileNotContains(path.Base(".env.example"), "AI_PROVIDER", modify.File(path.Base(".env.example")).Append(env))),
	).Uninstall(

		// Remove the AI facade
		modify.File(aiFacadePath).Remove(),

		// Remove config/ai.go
		modify.File(aiConfigPath).Remove(),

		// Remove the ai service provider from the providers array in bootstrap/providers.go
		modify.UnregisterProvider(modulePath, aiServiceProvider),
	).Execute()
}
