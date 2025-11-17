package path

import (
	"github.com/goravel/framework/facades"
)

func App(paths ...string) string {
	return facades.App().Path(paths...)
}

func Base(paths ...string) string {
	return facades.App().BasePath(paths...)
}

func Config(paths ...string) string {
	return facades.App().ConfigPath(paths...)
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

func Lang(paths ...string) string {
	return facades.App().LangPath(paths...)
}

func Public(paths ...string) string {
	return facades.App().PublicPath(paths...)
}
