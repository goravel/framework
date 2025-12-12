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
	mailConfigPath := path.Config("mail.go")
	mailFacadePath := path.Facades("mail.go")
	packageName := setup.Paths().Main().Package()
	modulePath := setup.Paths().Module().Import()
	mailServiceProvider := "&mail.ServiceProvider{}"
	env := `
MAIL_HOST=
MAIL_PORT=
MAIL_USERNAME=
MAIL_PASSWORD=
MAIL_FROM_ADDRESS=
MAIL_FROM_NAME=
`

	setup.Install(
		// Add the mail service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(modulePath, mailServiceProvider),

		// Create config/mail.go and the Mail facade
		modify.File(mailConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), packageName)),

		// Add the Mail facade
		modify.File(mailFacadePath).Overwrite(stubs.MailFacade(setup.Paths().Facades().Package())),

		// Add configurations to the .env and .env.example files
		modify.WhenFileNotContains(path.Base(".env"), "MAIL_HOST", modify.File(path.Base(".env")).Append(env)),
		modify.WhenFileNotContains(path.Base(".env.example"), "MAIL_HOST", modify.File(path.Base(".env.example")).Append(env)),
	).Uninstall(
		// Remove config/mail.go
		modify.File(mailConfigPath).Remove(),

		// Remove the mail service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(modulePath, mailServiceProvider),

		// Remove the Mail facade
		modify.File(mailFacadePath).Remove(),
	).Execute()
}
