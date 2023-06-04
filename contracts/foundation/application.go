package foundation

import (
	"github.com/goravel/framework/contracts/console"
)

//go:generate mockery --name=Application
type Application interface {
	Container
	Boot()
	Commands([]console.Command)
	Path(path string) string
	BasePath(path string) string
	ConfigPath(path string) string
	DatabasePath(path string) string
	StoragePath(path string) string
	PublicPath(path string) string
	Publishes(packageName string, paths map[string]string, groups ...string)
}
