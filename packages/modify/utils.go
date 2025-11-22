package modify

import (
	"slices"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support/path/internals"
)

// AddCommand adds command to the foundation.Setup() chain in the Boot function.
// If WithCommands doesn't exist, it creates a new commands.go file in the bootstrap directory based on the stubs.go:commands template,
// then add WithCommands(Commands()) to foundation.Setup(), add the command to Commands().
// If WithCommands exists, it appends the command to []console.Command if the commands.go file doesn't exist,
// or appends to the Commands() function if the commands.go file exists.
// This function also ensures the configuration package and command package are imported when creating WithCommands.
//
// Returns an error if commands.go exists but WithCommands is not registered in foundation.Setup(), as the commands.go file
// should only be created when adding WithCommands to Setup().
//
// Parameters:
//   - pkg: Package path of the command (e.g., "goravel/app/console/commands")
//   - command: Command expression to add (e.g., "&commands.ExampleCommand{}")
//
// Example usage:
//
//	AddCommand("goravel/app/console/commands", "&commands.ExampleCommand{}")
//
// This transforms (when commands.go doesn't exist and WithCommands doesn't exist):
//
//	foundation.Setup().WithConfig(config.Boot).Run()
//
// Into:
//
//	foundation.Setup().WithCommands(Commands()).WithConfig(config.Boot).Run()
//
// And creates bootstrap/commands.go:
//
//	package bootstrap
//	import "github.com/goravel/framework/contracts/console"
//	func Commands() []console.Command {
//	  return []console.Command{&commands.ExampleCommand{}}
//	}
//
// If WithCommands already exists but commands.go doesn't:
//
//	foundation.Setup().WithCommands([]console.Command{
//	  &commands.ExistingCommand{},
//	}).Run()
//
// It appends the new command:
//
//	foundation.Setup().WithCommands([]console.Command{
//	  &commands.ExistingCommand{},
//	  &commands.ExampleCommand{},
//	}).Run()
//
// If WithCommands exists with Commands() call and commands.go exists, it appends to Commands() function.
func AddCommand(pkg, command string) error {
	config := withSliceConfig{
		fileName:        "commands.go",
		withMethodName:  "WithCommands",
		helperFuncName:  "Commands",
		typePackage:     "console",
		typeName:        "Command",
		typeImportPath:  "github.com/goravel/framework/contracts/console",
		fileExistsError: errors.PackageCommandsFileExists,
		stubTemplate:    commands,
		matcherFunc:     match.Commands,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, command)
}

// AddMiddleware adds middleware to the foundation.Setup() chain in the Boot function.
// If WithMiddleware doesn't exist, it creates one. If it exists, it appends the middleware using handler.Append().
// This function also ensures the configuration package and middleware package are imported when creating WithMiddleware.
//
// Parameters:
//   - pkg: Package path of the middleware (e.g., "goravel/app/http/middleware")
//   - middleware: Middleware expression to add (e.g., "&Auth{}")
//
// Example usage:
//
//	AddMiddleware("goravel/app/http/middleware", "&Auth{}")
//
// This transforms:
//
//	foundation.Setup().WithConfig(config.Boot).Run()
//
// Into:
//
//	foundation.Setup().WithMiddleware(func(handler configuration.Middleware) {
//	    handler.Append(&middleware.Auth{})
//	}).WithConfig(config.Boot).Run()
//
// If WithMiddleware already exists:
//
//	foundation.Setup().WithMiddleware(func(handler configuration.Middleware) {
//	    handler.Append(&middleware.Existing{})
//	}).Run()
//
// It appends the new middleware:
//
//	foundation.Setup().WithMiddleware(func(handler configuration.Middleware) {
//	    handler.Append(&middleware.Existing{}, &middleware.Auth{})
//	}).Run()
func AddMiddleware(pkg, middleware string) error {
	appFilePath := internals.BootstrapApp()

	if err := addMiddlewareImports(appFilePath, pkg); err != nil {
		return err
	}

	return GoFile(appFilePath).Find(match.FoundationSetup()).Modify(foundationSetupMiddleware(middleware)).Apply()
}

// AddMigration adds migration to the foundation.Setup() chain in the Boot function.
// If WithMigrations doesn't exist, it creates a new migrations.go file in the bootstrap directory based on the stubs.go:migrations template,
// then add WithMigrations(Migrations()) to foundation.Setup(), add the migration to Migrations().
// If WithMigrations exists, it appends the migration to []schema.Migration if the migrations.go file doesn't exist,
// or appends to the Migrations() function if the migrations.go file exists.
// This function also ensures the configuration package and migration package are imported when creating WithMigrations.
//
// Returns an error if migrations.go exists but WithMigrations is not registered in foundation.Setup(), as the migrations.go file
// should only be created when adding WithMigrations to Setup().
//
// Parameters:
//   - pkg: Package path of the migration (e.g., "goravel/database/migrations")
//   - migration: Migration expression to add (e.g., "&migrations.ExampleMigration{}")
//
// Example usage:
//
//	AddMigration("goravel/database/migrations", "&migrations.ExampleMigration{}")
//
// This transforms (when migrations.go doesn't exist and WithMigrations doesn't exist):
//
//	foundation.Setup().WithConfig(config.Boot).Run()
//
// Into:
//
//	foundation.Setup().WithMigrations(Migrations()).WithConfig(config.Boot).Run()
//
// And creates bootstrap/migrations.go:
//
//	package bootstrap
//	import "github.com/goravel/framework/contracts/database/schema"
//	func Migrations() []schema.Migration {
//	  return []schema.Migration{&migrations.ExampleMigration{}}
//	}
//
// If WithMigrations already exists but migrations.go doesn't:
//
//	foundation.Setup().WithMigrations([]schema.Migration{
//	  &migrations.ExistingMigration{},
//	}).Run()
//
// It appends the new migration:
//
//	foundation.Setup().WithMigrations([]schema.Migration{
//	  &migrations.ExistingMigration{},
//	  &migrations.ExampleMigration{},
//	}).Run()
//
// If WithMigrations exists with Migrations() call and migrations.go exists, it appends to Migrations() function.
func AddMigration(pkg, migration string) error {
	config := withSliceConfig{
		fileName:        "migrations.go",
		withMethodName:  "WithMigrations",
		helperFuncName:  "Migrations",
		typePackage:     "schema",
		typeName:        "Migration",
		typeImportPath:  "github.com/goravel/framework/contracts/database/schema",
		fileExistsError: errors.PackageMigrationsFileExists,
		stubTemplate:    migrations,
		matcherFunc:     match.Migrations,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, migration)
}

// AddSeeder adds seeder to the foundation.Setup() chain in the Boot function.
// If WithSeeders doesn't exist, it creates a new seeders.go file in the bootstrap directory based on the stubs.go:seeders template,
// then adds WithSeeders(Seeders()) to foundation.Setup(), and adds the seeder to Seeders().
// If WithSeeders exists, it appends the seeder to []seeder.Seeder if the seeders.go file doesn't exist,
// or appends to the Seeders() function if the seeders.go file exists.
// This function also ensures the configuration package and seeder package are imported when creating WithSeeders.
//
// Returns an error if seeders.go exists but WithSeeders is not registered in foundation.Setup(), as the seeders.go file
// should only be created when adding WithSeeders to Setup().
//
// Parameters:
//   - pkg: Package path of the seeder (e.g., "goravel/database/seeders")
//   - seeder: Seeder expression to add (e.g., "&seeders.ExampleSeeder{}")
//
// Example usage:
//
//	AddSeeder("goravel/database/seeders", "&seeders.ExampleSeeder{}")
//
// This transforms (when seeders.go doesn't exist and WithSeeders doesn't exist):
//
//	foundation.Setup().WithConfig(config.Boot).Run()
//
// Into:
//
//	foundation.Setup().WithSeeders(Seeders()).WithConfig(config.Boot).Run()
//
// And creates bootstrap/seeders.go:
//
//	package bootstrap
//	import "github.com/goravel/framework/contracts/database/seeder"
//	func Seeders() []seeder.Seeder {
//	  return []seeder.Seeder{&seeders.ExampleSeeder{}}
//	}
//
// If WithSeeders already exists but seeders.go doesn't:
//
//	foundation.Setup().WithSeeders([]seeder.Seeder{
//	  &seeders.ExistingSeeder{},
//	}).Run()
//
// It appends the new seeder:
//
//	foundation.Setup().WithSeeders([]seeder.Seeder{
//	  &seeders.ExistingSeeder{},
//	  &seeders.ExampleSeeder{},
//	}).Run()
//
// If WithSeeders exists with Seeders() call and seeders.go exists, it appends to Seeders() function.
func AddSeeder(pkg, seeder string) error {
	config := withSliceConfig{
		fileName:        "seeders.go",
		withMethodName:  "WithSeeders",
		helperFuncName:  "Seeders",
		typePackage:     "seeder",
		typeName:        "Seeder",
		typeImportPath:  "github.com/goravel/framework/contracts/database/seeder",
		fileExistsError: errors.PackageSeedersFileExists,
		stubTemplate:    seeders,
		matcherFunc:     match.Seeders,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, seeder)
}

func ExprExists(x []dst.Expr, y dst.Expr) bool {
	return ExprIndex(x, y) >= 0
}

func ExprIndex(x []dst.Expr, y dst.Expr) int {
	return slices.IndexFunc(x, func(expr dst.Expr) bool {
		return match.EqualNode(y).MatchNode(expr)
	})
}

func IsUsingImport(df *dst.File, path string, name ...string) bool {
	if len(name) == 0 {
		split := strings.Split(path, "/")
		name = append(name, split[len(split)-1])
	}

	var used bool
	dst.Inspect(df, func(n dst.Node) bool {
		sel, ok := n.(*dst.SelectorExpr)
		if ok && isTopName(sel.X, name[0]) {
			used = true

			return false
		}
		return true
	})

	return used
}

func KeyExists(kvs []dst.Expr, key dst.Expr) bool {
	return KeyIndex(kvs, key) >= 0
}

func KeyIndex(kvs []dst.Expr, key dst.Expr) int {
	return slices.IndexFunc(kvs, func(expr dst.Expr) bool {
		if kv, ok := expr.(*dst.KeyValueExpr); ok {
			return match.EqualNode(key).MatchNode(kv.Key)
		}
		return false
	})
}

func MustParseExpr(x string) (node dst.Node) {
	src := "package p\nvar _ = " + x
	file, err := decorator.Parse(src)
	if err != nil {
		panic(err)
	}

	spec := file.Decls[0].(*dst.GenDecl).Specs[0].(*dst.ValueSpec)
	expr := spec.Values[0]

	// handle outer comments for expr
	expr.Decorations().Start = file.Decls[0].(*dst.GenDecl).Decorations().Start
	expr.Decorations().End = file.Decls[0].(*dst.GenDecl).Decorations().End

	return WrapNewline(expr)
}

func WrapNewline[T dst.Node](node T) T {
	dst.Inspect(node, func(n dst.Node) bool {
		switch v := n.(type) {
		case *dst.KeyValueExpr, *dst.UnaryExpr:
			v.Decorations().After = dst.NewLine
			v.Decorations().Before = dst.NewLine
		case *dst.FuncType:
			v.Results.Decorations().After = dst.NewLine
			v.Results.Decorations().Before = dst.NewLine
		}

		return true
	})

	return node
}

func isThirdParty(importPath string) bool {
	// Third party package import path usually contains "." (".com", ".org", ...)
	// This logic is taken from golang.org/x/tools/imports package.
	return strings.Contains(importPath, ".")
}

// isTopName returns true if n is a top-level unresolved identifier with the given name.
func isTopName(n dst.Expr, name string) bool {
	id, ok := n.(*dst.Ident)
	return ok && id.Name == name && id.Obj == nil
}
