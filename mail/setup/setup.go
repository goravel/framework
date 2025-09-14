package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}

	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&mail.ServiceProvider{}")),
			modify.File(path.Config("mail.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade("Mail", modify.File(path.Facades("mail.go")).Overwrite(stubs.MailFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Mail"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&mail.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("mail.go")).Remove(),
			),
			modify.WhenFacade("Mail", modify.File(path.Facades("mail.go")).Remove()),
		).
		Execute()
}
