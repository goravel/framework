// To avoid import cycle, only be used internally.

package internals

import (
	"path/filepath"

	"github.com/goravel/framework/support"
)

func Abs(paths ...string) string {
	paths = append([]string{support.RelativePath}, paths...)
	path := filepath.Join(paths...)
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

func BootstrapApp() string {
	bootstrap := support.PathToSlice(support.Config.Paths.Bootstrap)
	bootstrap = append(bootstrap, "app.go")

	return Abs(bootstrap...)
}

func Facades(path ...string) string {
	facades := support.PathToSlice(support.Config.Paths.Facades)
	path = append(facades, path...)

	return Abs(path...)
}

func Path(path ...string) string {
	app := support.PathToSlice(support.Config.Paths.App)
	path = append(app, path...)

	return Abs(path...)
}
