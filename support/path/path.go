package path

import (
	stdpath "path"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/packages"
	packagespaths "github.com/goravel/framework/packages/paths"
)

func App(paths ...string) string {
	return packages.Paths().App().Abs(paths...)
}

func Base(paths ...string) string {
	return packagespaths.Abs(paths...)
}

func Bootstrap(paths ...string) string {
	return packages.Paths().Bootstrap().Abs(paths...)
}

// CleanRelative cleans p as a slash-separated path relative to some root and
// reports whether it stays inside that root. "a/./b" and "a/../b" become "a/b"
// and "b"; an absolute path, ".." and anything starting with "../" are refused.
// "" and "." clean to ".", the root itself.
func CleanRelative(p string) (string, bool) {
	p = filepath.ToSlash(p)
	if stdpath.IsAbs(p) || filepath.IsAbs(filepath.FromSlash(p)) {
		return "", false
	}

	cleaned := stdpath.Clean(p)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", false
	}

	return cleaned, true
}

func Config(paths ...string) string {
	return packages.Paths().Config().Abs(paths...)
}

func Database(paths ...string) string {
	return packages.Paths().Database().Abs(paths...)
}

func Executable(paths ...string) string {
	return Base(paths...)
}

func Facade(paths ...string) string {
	return packages.Paths().Facades().Abs(paths...)
}

func Lang(paths ...string) string {
	return packages.Paths().Lang().Abs(paths...)
}

func Migration(paths ...string) string {
	return packages.Paths().Migrations().Abs(paths...)
}

func Model(paths ...string) string {
	return packages.Paths().Models().Abs(paths...)
}

func Public(paths ...string) string {
	return packages.Paths().Public().Abs(paths...)
}

func Resource(paths ...string) string {
	return packages.Paths().Resources().Abs(paths...)
}

func Route(paths ...string) string {
	return packages.Paths().Routes().Abs(paths...)
}

func Storage(paths ...string) string {
	return packages.Paths().Storage().Abs(paths...)
}

func Test(paths ...string) string {
	return packages.Paths().Tests().Abs(paths...)
}

func View(paths ...string) string {
	return packages.Paths().Views().Abs(paths...)
}
