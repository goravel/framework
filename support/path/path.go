package path

import (
	"github.com/goravel/framework/facades"
)

func App(paths ...string) string {
	finalPath := ""
	if len(paths) >= 1 {
		finalPath = paths[0]
	}

	return facades.App().Path(finalPath)
}

func Base(paths ...string) string {
	finalPath := ""
	if len(paths) >= 1 {
		finalPath = paths[0]
	}

	return facades.App().BasePath(finalPath)
}

func Config(paths ...string) string {
	finalPath := ""
	if len(paths) >= 1 {
		finalPath = paths[0]
	}

	return facades.App().ConfigPath(finalPath)
}

func Database(paths ...string) string {
	finalPath := ""
	if len(paths) >= 1 {
		finalPath = paths[0]
	}

	return facades.App().DatabasePath(finalPath)
}

func Storage(paths ...string) string {
	finalPath := ""
	if len(paths) >= 1 {
		finalPath = paths[0]
	}

	return facades.App().StoragePath(finalPath)
}

func Public(paths ...string) string {
	finalPath := ""
	if len(paths) >= 1 {
		finalPath = paths[0]
	}

	return facades.App().PublicPath(finalPath)
}
