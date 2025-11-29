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

// AddJob adds job to the foundation.Setup() chain in the Boot function.
// If WithJobs doesn't exist, it creates a new jobs.go file in the bootstrap directory based on the stubs.go:jobs template,
// then adds WithJobs(Jobs()) to foundation.Setup(), add the job to Jobs().
// If WithJobs exists, it appends the job to []queue.Job if the jobs.go file doesn't exist,
// or appends to the Jobs() function if the jobs.go file exists.
// This function also ensures the configuration package and job package are imported when creating WithJobs.
//
// Returns an error if jobs.go exists but WithJobs is not registered in foundation.Setup(), as the jobs.go file
// should only be created when adding WithJobs to Setup().
//
// Parameters:
//   - pkg: Package path of the job (e.g., "goravel/app/jobs")
//   - job: Job expression to add (e.g., "&jobs.ExampleJob{}")
//
// Example usage:
//
//	AddJob("goravel/app/jobs", "&jobs.ExampleJob{}")
//
// This transforms (when jobs.go doesn't exist and WithJobs doesn't exist):
//
//	foundation.Setup().WithConfig(config.Boot).Run()
//
// Into:
//
//	foundation.Setup().WithJobs(Jobs()).WithConfig(config.Boot).Run()
//
// And creates bootstrap/jobs.go:
//
//	package bootstrap
//	import "github.com/goravel/framework/contracts/queue"
//	func Jobs() []queue.Job {
//	  return []queue.Job{&jobs.ExampleJob{}}
//	}
//
// If WithJobs already exists but jobs.go doesn't:
//
//	foundation.Setup().WithJobs([]queue.Job{
//	  &jobs.ExistingJob{},
//	}).Run()
//
// It appends the new job:
//
//	foundation.Setup().WithJobs([]queue.Job{
//	  &jobs.ExistingJob{},
//	  &jobs.ExampleJob{},
//	}).Run()
//
// If WithJobs exists with Jobs() call and jobs.go exists, it appends to Jobs() function.
func AddJob(pkg, job string) error {
	config := withSliceConfig{
		fileName:        "jobs.go",
		withMethodName:  "WithJobs",
		helperFuncName:  "Jobs",
		typePackage:     "queue",
		typeName:        "Job",
		typeImportPath:  "github.com/goravel/framework/contracts/queue",
		fileExistsError: errors.PackageJobsFileExists,
		stubTemplate:    jobs,
		matcherFunc:     match.Jobs,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, job)
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

// AddProvider adds service provider to the foundation.Setup() chain in the Boot function.
// If WithProviders doesn't exist, it creates a new providers.go file in the bootstrap directory based on the stubs.go:providers template,
// then add WithProviders(Providers()) to foundation.Setup(), add the provider to Providers().
// If WithProviders exists, it appends the provider to []foundation.ServiceProvider if the providers.go file doesn't exist,
// or appends to the Providers() function if the providers.go file exists.
// This function also ensures the configuration package and provider package are imported when creating WithProviders.
//
// Returns an error if providers.go exists but WithProviders is not registered in foundation.Setup(), as the providers.go file
// should only be created when adding WithProviders to Setup().
//
// Parameters:
//   - pkg: Package path of the provider (e.g., "goravel/app/providers")
//   - provider: Provider expression to add (e.g., "&providers.AppServiceProvider{}")
//
// Example usage:
//
//	AddProvider("goravel/app/providers", "&providers.AppServiceProvider{}")
//
// This transforms (when providers.go doesn't exist and WithProviders doesn't exist):
//
//	foundation.Setup().WithConfig(config.Boot).Run()
//
// Into:
//
//	foundation.Setup().WithProviders(Providers()).WithConfig(config.Boot).Run()
//
// And creates bootstrap/providers.go:
//
//	package bootstrap
//	import "github.com/goravel/framework/contracts/foundation"
//	func Providers() []foundation.ServiceProvider {
//	  return []foundation.ServiceProvider{&providers.AppServiceProvider{}}
//	}
//
// If WithProviders already exists but providers.go doesn't:
//
//	foundation.Setup().WithProviders([]foundation.ServiceProvider{
//	  &providers.ExistingProvider{},
//	}).Run()
//
// It appends the new provider:
//
//	foundation.Setup().WithProviders([]foundation.ServiceProvider{
//	  &providers.ExistingProvider{},
//	  &providers.AppServiceProvider{},
//	}).Run()
//
// If WithProviders exists with Providers() call and providers.go exists, it appends to Providers() function.
func AddProvider(pkg, provider string) error {
	config := withSliceConfig{
		fileName:        "providers.go",
		withMethodName:  "WithProviders",
		helperFuncName:  "Providers",
		typePackage:     "foundation",
		typeName:        "ServiceProvider",
		typeImportPath:  "github.com/goravel/framework/contracts/foundation",
		fileExistsError: errors.PackageProvidersFileExists,
		stubTemplate:    providers,
		matcherFunc:     match.Providers,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, provider)
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

// ExprExists checks if an expression exists in a slice of expressions.
// It uses structural equality comparison via ExprIndex.
//
// Parameters:
//   - x: Slice of expressions to search in
//   - y: Expression to search for
//
// Returns true if the expression exists in the slice, false otherwise.
//
// Example:
//
//	exprs := []dst.Expr{
//		&dst.Ident{Name: "foo"},
//		&dst.Ident{Name: "bar"},
//	}
//	target := &dst.Ident{Name: "foo"}
//	if ExprExists(exprs, target) {
//		fmt.Println("Expression found")
//	}
func ExprExists(x []dst.Expr, y dst.Expr) bool {
	return ExprIndex(x, y) >= 0
}

// ExprIndex returns the index of the first occurrence of an expression in a slice.
// It uses structural equality comparison to match expressions.
//
// Parameters:
//   - x: Slice of expressions to search in
//   - y: Expression to search for
//
// Returns the index of the first occurrence, or -1 if not found.
//
// Example:
//
//	exprs := []dst.Expr{
//		&dst.Ident{Name: "foo"},
//		&dst.Ident{Name: "bar"},
//		&dst.Ident{Name: "baz"},
//	}
//	target := &dst.Ident{Name: "bar"}
//	index := ExprIndex(exprs, target) // returns 1
func ExprIndex(x []dst.Expr, y dst.Expr) int {
	return slices.IndexFunc(x, func(expr dst.Expr) bool {
		return match.EqualNode(y).MatchNode(expr)
	})
}

// IsUsingImport checks if an imported package is actually used in the file.
// It inspects the AST for selector expressions that reference the package.
//
// Parameters:
//   - df: The parsed Go file to inspect
//   - path: Import path of the package (e.g., "github.com/goravel/framework/contracts/console")
//   - name: Optional package name. If not provided, uses the last segment of the path
//
// Returns true if the package is used anywhere in the file, false otherwise.
//
// Example:
//
//	file, _ := decorator.Parse(src)
//	// Check if "console" package from "github.com/goravel/framework/contracts/console" is used
//	if IsUsingImport(file, "github.com/goravel/framework/contracts/console") {
//		fmt.Println("console package is being used")
//	}
//	// Or specify a custom name
//	if IsUsingImport(file, "github.com/goravel/framework/contracts/console", "customName") {
//		fmt.Println("Package with alias 'customName' is being used")
//	}
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

// KeyExists checks if a key exists in a slice of key-value expressions.
// It uses structural equality comparison via KeyIndex.
//
// Parameters:
//   - kvs: Slice of expressions (expected to contain KeyValueExpr)
//   - key: Key expression to search for
//
// Returns true if the key exists in any KeyValueExpr, false otherwise.
//
// Example:
//
//	kvExprs := []dst.Expr{
//		&dst.KeyValueExpr{
//			Key:   &dst.Ident{Name: "name"},
//			Value: &dst.BasicLit{Value: `"John"`},
//		},
//		&dst.KeyValueExpr{
//			Key:   &dst.Ident{Name: "age"},
//			Value: &dst.BasicLit{Value: "30"},
//		},
//	}
//	targetKey := &dst.Ident{Name: "name"}
//	if KeyExists(kvExprs, targetKey) {
//		fmt.Println("Key found")
//	}
func KeyExists(kvs []dst.Expr, key dst.Expr) bool {
	return KeyIndex(kvs, key) >= 0
}

// KeyIndex returns the index of a key in a slice of key-value expressions.
// It searches for KeyValueExpr nodes and compares their keys using structural equality.
//
// Parameters:
//   - kvs: Slice of expressions (expected to contain KeyValueExpr)
//   - key: Key expression to search for
//
// Returns the index of the first KeyValueExpr with matching key, or -1 if not found.
//
// Example:
//
//	kvExprs := []dst.Expr{
//		&dst.KeyValueExpr{
//			Key:   &dst.Ident{Name: "name"},
//			Value: &dst.BasicLit{Value: `"John"`},
//		},
//		&dst.KeyValueExpr{
//			Key:   &dst.Ident{Name: "age"},
//			Value: &dst.BasicLit{Value: "30"},
//		},
//	}
//	targetKey := &dst.Ident{Name: "age"}
//	index := KeyIndex(kvExprs, targetKey) // returns 1
func KeyIndex(kvs []dst.Expr, key dst.Expr) int {
	return slices.IndexFunc(kvs, func(expr dst.Expr) bool {
		if kv, ok := expr.(*dst.KeyValueExpr); ok {
			return match.EqualNode(key).MatchNode(kv.Key)
		}
		return false
	})
}

// MustParseExpr parses a Go expression from a string and returns its AST node.
// It wraps the expression in a minimal valid Go program to parse it, then extracts
// and returns the expression node with proper decorations and newlines.
//
// Parameters:
//   - x: String representation of a Go expression
//
// Returns the parsed expression as a dst.Node, with decorations preserved.
// Panics if the expression cannot be parsed.
//
// Example:
//
//	// Parse a simple expression
//	node := MustParseExpr("&commands.ExampleCommand{}")
//	// Returns a UnaryExpr node representing the address-of operation
//
//	// Parse a composite literal
//	node := MustParseExpr(`map[string]interface{}{"key": "value"}`)
//	// Returns a CompositeLit node
//
//	// Parse a function call
//	node := MustParseExpr("fmt.Println(\"hello\")")
//	// Returns a CallExpr node
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

// RemoveProvider removes a service provider from the foundation.Setup() chain in the Boot function.
// If providers.go exists, it removes the provider from the Providers() function in that file.
// If providers.go doesn't exist, it removes the provider from the inline array in app.go.
// This function also cleans up unused imports after removing the provider.
//
// Parameters:
//   - pkg: Package path of the provider (e.g., "goravel/app/providers")
//   - provider: Provider expression to remove (e.g., "&providers.AppServiceProvider{}")
//
// Example usage:
//
//	RemoveProvider("goravel/app/providers", "&providers.AppServiceProvider{}")
//
// If providers.go exists with:
//
//	func Providers() []foundation.ServiceProvider {
//	  return []foundation.ServiceProvider{
//	    &providers.AppServiceProvider{},
//	    &providers.OtherProvider{},
//	  }
//	}
//
// After removal:
//
//	func Providers() []foundation.ServiceProvider {
//	  return []foundation.ServiceProvider{
//	    &providers.OtherProvider{},
//	  }
//	}
//
// If providers.go doesn't exist and app.go has:
//
//	foundation.Setup().WithProviders([]foundation.ServiceProvider{
//	  &providers.AppServiceProvider{},
//	  &providers.OtherProvider{},
//	}).Run()
//
// After removal:
//
//	foundation.Setup().WithProviders([]foundation.ServiceProvider{
//	  &providers.OtherProvider{},
//	}).Run()
func RemoveProvider(pkg, provider string) error {
	config := withSliceConfig{
		fileName:       "providers.go",
		withMethodName: "WithProviders",
		helperFuncName: "Providers",
		typePackage:    "foundation",
		typeName:       "ServiceProvider",
		typeImportPath: "github.com/goravel/framework/contracts/foundation",
		matcherFunc:    match.Providers,
	}

	handler := newWithSliceHandler(config)
	return handler.RemoveItem(pkg, provider)
}

// WrapNewline adds newline decorations to specific AST nodes for better formatting.
// It traverses the AST and adds Before/After newlines to KeyValueExpr, UnaryExpr,
// and FuncType result nodes to improve code readability.
//
// Parameters:
//   - node: Any dst.Node to process
//
// Returns the same node with newline decorations applied.
//
// Example:
//
//	// Parse and wrap an expression
//	expr := MustParseExpr("&commands.ExampleCommand{}")
//	// The UnaryExpr will have newlines before and after
//
//	// For a composite literal with key-value pairs:
//	node := MustParseExpr(`map[string]int{"a": 1, "b": 2}`)
//	wrapped := WrapNewline(node)
//	// Each KeyValueExpr will have newlines for better formatting:
//	// map[string]int{
//	//     "a": 1,
//	//     "b": 2,
//	// }
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

// isThirdParty determines if an import path refers to a third-party package.
// Third-party packages typically contain a domain (e.g., ".com", ".org") in their path.
// This heuristic is taken from golang.org/x/tools/imports package.
//
// Parameters:
//   - importPath: The import path to check
//
// Returns true if the import path appears to be a third-party package, false for standard library.
//
// Example:
//
//	isThirdParty("fmt") // false - standard library
//	isThirdParty("encoding/json") // false - standard library
//	isThirdParty("github.com/goravel/framework") // true - third party
//	isThirdParty("example.com/package") // true - third party
func isThirdParty(importPath string) bool {
	// Third party package import path usually contains "." (".com", ".org", ...)
	// This logic is taken from golang.org/x/tools/imports package.
	return strings.Contains(importPath, ".")
}

// isTopName checks if an expression is a top-level unresolved identifier with the given name.
// An identifier is considered "top-level" and "unresolved" when it has no associated object,
// meaning it refers to a package name or other non-local identifier.
//
// Parameters:
//   - n: Expression to check
//   - name: Expected identifier name
//
// Returns true if n is an Ident with the given name and no associated object, false otherwise.
//
// Example:
//
//	// In the expression "fmt.Println", "fmt" is a top-level identifier
//	selectorExpr := &dst.SelectorExpr{
//		X:   &dst.Ident{Name: "fmt", Obj: nil},
//		Sel: &dst.Ident{Name: "Println"},
//	}
//	isTopName(selectorExpr.X, "fmt") // true
//
//	// A local variable has an associated Obj, so it's not a top-level name
//	localVar := &dst.Ident{Name: "x", Obj: &dst.Object{}}
//	isTopName(localVar, "x") // false
func isTopName(n dst.Expr, name string) bool {
	id, ok := n.(*dst.Ident)
	return ok && id.Name == name && id.Obj == nil
}
