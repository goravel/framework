package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	notificationConfigPath := path.Config("notification.go")
	notificationFacadePath := path.Facade("notification.go")
	notificationServiceProvider := "&notification.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()
	env := `
NOTIFICATION_CHANNEL=mail
SLACK_WEBHOOK_URL=
SLACK_USERNAME=Goravel
SLACK_ICON_EMOJI=:bell:
NOTIFICATION_DB_CONNECTION=
`

	setup.Install(
		// Avoid duplicate installation when installing dependent facades
		modify.WhenFacade(facades.Notification,
			// Add the notification service provider to the providers array in bootstrap/providers.go
			modify.RegisterProvider(moduleImport, notificationServiceProvider),

			// Create config/notification.go
			modify.File(notificationConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

			// Add the Notification facade
			modify.File(notificationFacadePath).Overwrite(stubs.NotificationFacade(facadesPackage)),

			// Add configurations to the .env and .env.example files
			modify.WhenFileNotContains(path.Base(".env"), "NOTIFICATION_CHANNEL", modify.File(path.Base(".env")).Append(env)),
			modify.WhenFileNotContains(path.Base(".env.example"), "NOTIFICATION_CHANNEL", modify.File(path.Base(".env.example")).Append(env)),
		),
	).Uninstall(
		modify.WhenFacade(facades.Notification,
			// Remove config/notification.go
			modify.File(notificationConfigPath).Remove(),

			// Remove the notification service provider from the providers array in bootstrap/providers.go
			modify.UnregisterProvider(moduleImport, notificationServiceProvider),

			// Remove the Notification facade
			modify.File(notificationFacadePath).Remove(),
		),
	).Execute()
}
