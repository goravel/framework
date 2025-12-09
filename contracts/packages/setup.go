package packages

import (
	"github.com/goravel/framework/contracts/packages/modify"
)

type Setup interface {
	// Execute runs the setup command based on the provided arguments.
	Execute()
	// ModulePath returns the module path of package, it may be a sub-package, eg: github.com/goravel/framework/auth.
	ModulePath() string
	// PackageName returns the package name of application, eg: goravel.
	PackageName() string
	// Install adds the provided modifiers to be executed during installation.
	Install(modifiers ...modify.Apply) Setup
	// Uninstall adds the provided modifiers to be executed during uninstallation.
	Uninstall(modifiers ...modify.Apply) Setup
}
