// To avoid import cycle, only be used internally.

package internals

import (
	"path/filepath"

	"github.com/goravel/framework/support"
)

func Abs(paths ...string) string {
	path := filepath.Join(paths...)
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

func BootstrapApp() string {
	return Abs(support.RelativePath, support.Config.Paths.Bootstrap, "app.go")
}

func Facades(path ...string) string {
	path = append([]string{"facades"}, path...)

	return Path(path...)
}

func Path(path ...string) string {
	path = append([]string{support.RelativePath, "app"}, path...)

	return Abs(path...)
}
