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
	broadcastConfigPath := path.Config("broadcasting.go")
	broadcastFacadePath := path.Facade("broadcast.go")
	broadcastServiceProvider := "&broadcasting.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	env := `
BROADCAST_CONNECTION=log
PUSHER_APP_KEY=
PUSHER_APP_SECRET=
PUSHER_APP_ID=
PUSHER_HOST=
`

	setup.Install(
		modify.RegisterProvider(moduleImport, broadcastServiceProvider),

		modify.File(broadcastConfigPath).Overwrite(stubs.Config()),

		modify.File(broadcastFacadePath).Overwrite(stubs.BroadcastFacade(setup.Paths().Facades().Package())),

		modify.WhenFileNotContains(path.Base(".env"), "BROADCAST_CONNECTION", modify.File(path.Base(".env")).Append(env)),
		modify.WhenFileNotContains(path.Base(".env.example"), "BROADCAST_CONNECTION", modify.File(path.Base(".env.example")).Append(env)),
	).Uninstall(
		modify.File(broadcastConfigPath).Remove(),

		modify.UnregisterProvider(moduleImport, broadcastServiceProvider),

		modify.File(broadcastFacadePath).Remove(),
	).Execute()
}
