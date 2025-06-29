package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	supportfile "github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

func main() {
	config, err := supportfile.GetFrameworkContent("cache/config/cache.go")
	if err != nil {
		panic(err)
	}

	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&cache.ServiceProvider{}")),
			modify.File(path.Config("cache.go")).Overwrite(config),
		).
		Uninstall(
			modify.GoFile(path.Config("app.go")).
				Find(match.Providers()).Modify(modify.Unregister("&cache.ServiceProvider{}")).
				Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			modify.File(path.Config("cache.go")).Remove(),
		).
		Execute()
}
