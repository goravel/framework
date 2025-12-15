package path

import (
	"github.com/goravel/framework/packages"
	packagespaths "github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
)

func App(paths ...string) string {
	return packages.Paths().App().String(paths...)
}

func Base(paths ...string) string {
	return packagespaths.Abs(paths...)
}

func Bootstrap(paths ...string) string {
	return packages.Paths().Bootstrap().String(paths...)
}

func Config(paths ...string) string {
	return packages.Paths().Config().String(paths...)
}

func Database(paths ...string) string {
	return packages.Paths().Database().String(paths...)
}

func Executable(paths ...string) string {
	paths = append([]string{support.RootPath}, paths...)

	return Base(paths...)
}

func Facade(paths ...string) string {
	return packages.Paths().Facades().String(paths...)
}

func Lang(paths ...string) string {
	return packages.Paths().Lang().String(paths...)
}

func Migration(paths ...string) string {
	return packages.Paths().Migrations().String(paths...)
}

func Model(paths ...string) string {
	return packages.Paths().Models().String(paths...)
}

func Public(paths ...string) string {
	return packages.Paths().Public().String(paths...)
}

func Resource(paths ...string) string {
	return packages.Paths().Resources().String(paths...)
}

func Route(paths ...string) string {
	return packages.Paths().Routes().String(paths...)
}

func Storage(paths ...string) string {
	return packages.Paths().Storage().String(paths...)
}

func View(paths ...string) string {
	return packages.Paths().Views().String(paths...)
}
