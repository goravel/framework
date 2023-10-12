package foundation

import (
	"github.com/goravel/framework/contracts/console"
)

//go:generate mockery --name=Application
type Application interface {
	Container
	// Boot register and bootstrap configured service providers.
	Boot()
	// Commands register the given commands with the console application.
	Commands([]console.Command)
	// Path gets the path respective to "app" directory.
	Path(path string) string
	// BasePath get the base path of the Goravel installation.
	BasePath(path string) string
	// ConfigPath get the path to the configuration files.
	ConfigPath(path string) string
	// DatabasePath get the path to the database directory.
	DatabasePath(path string) string
	// StoragePath get the path to the storage directory.
	StoragePath(path string) string
	// PublicPath get the path to the public directory.
	PublicPath(path string) string
	// Publishes register the given paths to be published by the "vendor:publish" command.
	Publishes(packageName string, paths map[string]string, groups ...string)
}
