package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	// http, err := supportfile.GetFrameworkContent("http/setup/config/http.go")
	// if err != nil {
	// 	panic(err)
	// }

	// jwt, err := supportfile.GetFrameworkContent("http/setup/config/jwt.go")
	// if err != nil {
	// 	panic(err)
	// }

	// cors, err := supportfile.GetFrameworkContent("http/setup/config/cors.go")
	// if err != nil {
	// 	panic(err)
	// }

	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&http.ServiceProvider{}")),
			// modify.File(path.Config("http.go")).Overwrite(http),
			// modify.File(path.Config("jwt.go")).Overwrite(jwt),
			// modify.File(path.Config("cors.go")).Overwrite(cors),
		).
		Uninstall(
			modify.GoFile(path.Config("app.go")).
				Find(match.Providers()).Modify(modify.Unregister("&http.ServiceProvider{}")).
				Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			// modify.File(path.Config("http.go")).Remove(),
			// modify.File(path.Config("jwt.go")).Remove(),
			// modify.File(path.Config("cors.go")).Remove(),
		).
		Execute()
}
