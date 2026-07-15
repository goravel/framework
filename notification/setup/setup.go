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
	notificationFacadePath := path.Facade("notification.go")
	notificationServiceProvider := "&notification.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()

	setup.Install(
		// Avoid duplicate installation when installing dependent facades
		modify.WhenFacade(facades.Notification,
			// Add the notification service provider to the providers array in bootstrap/providers.go
			modify.RegisterProvider(moduleImport, notificationServiceProvider),

			// Add the Notification facade
			modify.File(notificationFacadePath).Overwrite(stubs.NotificationFacade(facadesPackage)),
		),
	).Uninstall(
		modify.WhenFacade(facades.Notification,
			// Remove the notification service provider from the providers array in bootstrap/providers.go
			modify.UnregisterProvider(moduleImport, notificationServiceProvider),

			// Remove the Notification facade
			modify.File(notificationFacadePath).Remove(),
		),
	).Execute()
}
