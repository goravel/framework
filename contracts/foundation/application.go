package foundation

import (
	"context"

	"github.com/goravel/framework/contracts/console"
)

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
	// LangPath get the path to the language files.
	LangPath(path string) string
	// PublicPath get the path to the public directory.
	PublicPath(path string) string
	// ExecutablePath get the path to the executable of the running Goravel application.
	ExecutablePath() (string, error)
	// Publishes register the given paths to be published by the "vendor:publish" command.
	Publishes(packageName string, paths map[string]string, groups ...string)
	// CurrentLocale get the current application locale.
	CurrentLocale(ctx context.Context) string
	// SetLocale set the current application locale.
	SetLocale(ctx context.Context, locale string) context.Context
	// Version gets the version number of the application.
	Version() string
	// IsLocale get the current application locale.
	IsLocale(ctx context.Context, locale string) bool
	// SetJson set the JSON implementation.
	SetJson(json Json)
	// GetJson get the JSON implementation.
	GetJson() Json
}
