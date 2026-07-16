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
		modify.WhenFacade(facades.Notification,
			modify.RegisterProvider(moduleImport, notificationServiceProvider),
			modify.File(notificationFacadePath).Overwrite(stubs.NotificationFacade(facadesPackage)),
		),
	).Uninstall(
		modify.WhenFacade(facades.Notification,
			modify.UnregisterProvider(moduleImport, notificationServiceProvider),
			modify.File(notificationFacadePath).Remove(),
		),
	).Execute()
}
