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
	broadcastFacadePath := path.Facade("broadcast.go")
	broadcastServiceProvider := "&broadcasting.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	setup.Install(
		modify.RegisterProvider(moduleImport, broadcastServiceProvider),
		modify.File(broadcastFacadePath).Overwrite(stubs.BroadcastFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		modify.File(broadcastFacadePath).Remove(),
		modify.UnregisterProvider(moduleImport, broadcastServiceProvider),
	).Execute()
}
