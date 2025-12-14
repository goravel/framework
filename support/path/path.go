package path

import (
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/path/internals"
)

/******************************************
DONT USE BELOW FUNCTIONS IN THE FRAMEWORK TO AVOID IMPORT CYCLE.
INJECT THE APP AND USE app.*Path() INSTEAD.
*******************************************/

func App(paths ...string) string {
	return facades.App().Path(paths...)
}

func Base(paths ...string) string {
	return facades.App().BasePath(paths...)
}

func Bootstrap(paths ...string) string {
	return facades.App().BootstrapPath(paths...)
}

func Config(paths ...string) string {
	return facades.App().ConfigPath(paths...)
}

func Migration(paths ...string) string {
	paths = append(internals.ToSlice(support.Config.Paths.Migrations), paths...)

	return internals.Abs(paths...)
}

func Model(paths ...string) string {
	return facades.App().ModelPath(paths...)
}

func Database(paths ...string) string {
	return facades.App().DatabasePath(paths...)
}

func Executable(paths ...string) string {
	return facades.App().ExecutablePath(paths...)
}

func Facades(paths ...string) string {
	return facades.App().FacadesPath(paths...)
}

func Storage(paths ...string) string {
	return facades.App().StoragePath(paths...)
}

func Resource(paths ...string) string {
	return facades.App().ResourcePath(paths...)
}

func Route(paths ...string) string {
	paths = append(internals.ToSlice(support.Config.Paths.Routes), paths...)

	return internals.Abs(paths...)
}

func Lang(paths ...string) string {
	return facades.App().LangPath(paths...)
}

func Public(paths ...string) string {
	return facades.App().PublicPath(paths...)
}

func View(paths ...string) string {
	paths = append(internals.ToSlice(support.Config.Paths.Views), paths...)

	return internals.Abs(paths...)
}
