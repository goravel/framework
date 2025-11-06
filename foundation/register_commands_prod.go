//go:build production

package foundation

import (
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation/console"
)

func getRegisteredCommands(r *Application) []contractsconsole.Command {
	return []contractsconsole.Command{
		console.NewAboutCommand(r),
	}
}
