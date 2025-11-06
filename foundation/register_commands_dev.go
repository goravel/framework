//go:build !production

package foundation

import (
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation/console"
)

func getRegisteredCommands(r *Application) []contractsconsole.Command {
	return []contractsconsole.Command{
		console.NewAboutCommand(r),
		console.NewEnvEncryptCommand(),
		console.NewEnvDecryptCommand(),
		console.NewTestMakeCommand(),
		console.NewPackageMakeCommand(),
		console.NewProviderMakeCommand(),
		console.NewPackageInstallCommand(binding.Bindings, r.Bindings()),
		console.NewPackageUninstallCommand(r, binding.Bindings, r.Bindings()),
		console.NewVendorPublishCommand(r.publishes, r.publishGroups),
	}
}
