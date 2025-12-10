package packages

import (
	"github.com/goravel/framework/contracts/packages/modify"
)

type Setup interface {
	// Execute runs the setup command based on the provided arguments.
	Execute()
	// Package returns the package information of application.
	Paths() Paths
	// Install adds the provided modifiers to be executed during installation.
	Install(modifiers ...modify.Apply) Setup
	// Uninstall adds the provided modifiers to be executed during uninstallation.
	Uninstall(modifiers ...modify.Apply) Setup
}

type Paths interface {
	// Bootstrap returns the path for the bootstrap package, eg: goravel/bootstrap.
	Bootstrap() Path
	// Config returns the path for the config package, eg: goravel/config.
	Config() Path
	// Facades returns the path for the facades package, eg: goravel/app/facades.
	Facades() Path
	// Main returns the path for the main package, eg: github.com/goravel/goravel.
	Main() Path
	// Module returns the path for the module package, eg: github.com/goravel/framework/auth.
	Module() Path
	// Routes returns the path for the routes package, eg: goravel/routes.
	Routes() Path
	// Tests returns the path for the tests package, eg: goravel/tests.
	Tests() Path
}

type Path interface {
	// Package returns the sub-package name, or the main package name if no sub-package path is specified.
	Package() string
	// Import returns the sub-package import path, or the main package import path if no sub-package path is specified.
	Import() string
}
