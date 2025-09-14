package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	module := packages.GetModuleNameFromArgs(os.Args)
	stubs := Stubs{}

	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&http.ServiceProvider{}")),
			modify.File(path.Config("http.go")).Overwrite(stubs.HttpConfig(module)),
			modify.File(path.Config("jwt.go")).Overwrite(stubs.JwtConfig(module)),
			modify.File(path.Config("cors.go")).Overwrite(stubs.CorsConfig(module)),
			modify.WhenFacade("Http", modify.File(path.Facades("http.go")).Overwrite(stubs.HttpFacade())),
			modify.WhenFacade("RateLimiter", modify.File(path.Facades("rate_limiter.go")).Overwrite(stubs.RateLimiterFacade())),
			modify.WhenFacade("View", modify.File(path.Facades("view.go")).Overwrite(stubs.ViewFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Http", "RateLimiter", "View"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&http.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("http.go")).Remove(),
				modify.File(path.Config("jwt.go")).Remove(),
				modify.File(path.Config("cors.go")).Remove(),
			),
			modify.WhenFacade("Http", modify.File(path.Facades("http.go")).Remove()),
			modify.WhenFacade("RateLimiter", modify.File(path.Facades("rate_limiter.go")).Remove()),
			modify.WhenFacade("View", modify.File(path.Facades("view.go")).Remove()),
		).
		Execute()
}
