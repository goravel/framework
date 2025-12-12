// To avoid import cycle, only be used internally.

package internals

import (
	"path/filepath"
	"strings"

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
	bootstrap := ToSlice(support.Config.Paths.Bootstrap)
	bootstrap = append(bootstrap, "app.go")

	return Abs(bootstrap...)
}

func Facades(path ...string) string {
	facades := ToSlice(support.Config.Paths.Facades)
	path = append(facades, path...)

	return Abs(path...)
}

func Path(path ...string) string {
	app := ToSlice(support.Config.Paths.App)
	path = append(app, path...)

	return Abs(path...)
}

// ToSlice converts a file path string into a slice of its components,
// handling both forward slashes and backslashes, and trimming leading/trailing slashes.
// For example, "app/http/controllers" becomes []string{"app", "http", "controllers"}.
func ToSlice(path string) []string {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}

	return strings.Split(path, "/")
}
