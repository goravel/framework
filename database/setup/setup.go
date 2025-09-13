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
				Find(match.Providers()).Modify(modify.Register("&database.ServiceProvider{}")),
			modify.File(path.Config("database.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.File(path.Database("kernel.go")).Overwrite(stubs.Kernel(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade("DB", modify.File(path.Facades("db.go")).Overwrite(stubs.DBFacade())),
			modify.WhenFacade("Orm", modify.File(path.Facades("orm.go")).Overwrite(stubs.OrmFacade())),
			modify.WhenFacade("Schema", modify.File(path.Facades("schema.go")).Overwrite(stubs.SchemaFacade())),
			modify.WhenFacade("Seeder", modify.File(path.Facades("seeder.go")).Overwrite(stubs.SeederFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"DB", "Orm", "Schema", "Seeder"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&database.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("database.go")).Remove(),
				modify.File(path.Database("kernel.go")).Remove(),
			),
			modify.WhenFacade("DB", modify.File(path.Facades("db.go")).Remove()),
			modify.WhenFacade("Orm", modify.File(path.Facades("orm.go")).Remove()),
			modify.WhenFacade("Schema", modify.File(path.Facades("schema.go")).Remove()),
			modify.WhenFacade("Seeder", modify.File(path.Facades("seeder.go")).Remove()),
		).
		Execute()
}
