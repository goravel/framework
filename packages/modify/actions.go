package modify

import (
	"go/token"
	"slices"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
)

// Add adds an expression to the matched specified function.
func Add(expression string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		expr := MustParseExpr(expression).(dst.Expr)
		stmt := &dst.ExprStmt{
			X: expr,
		}

		node := cursor.Node().(*dst.FuncDecl)
		node.Body.List = append(node.Body.List, stmt)
	}
}

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

// AddConfig adds a configuration key with the given expression to the config file.
func AddConfig(name, expression string, annotations ...string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		var value *dst.CompositeLit
		switch node := cursor.Node().(type) {
		case *dst.KeyValueExpr:
			value = node.Value.(*dst.CompositeLit)
		case *dst.CallExpr:
			value = node.Args[1].(*dst.CompositeLit)
		}

		key := WrapNewline(&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(name)})
		newExpr := WrapNewline(&dst.KeyValueExpr{
			Key:   key,
			Value: WrapNewline(MustParseExpr(expression)).(dst.Expr),
		})

		// Add annotations as comments if provided
		if len(annotations) > 0 {
			var comments dst.Decorations
			for _, annotation := range annotations {
				// Ensure the annotation starts with "//" for proper comment formatting
				if !strings.HasPrefix(annotation, "//") {
					annotation = "// " + annotation
				}
				comments = append(comments, annotation)
			}
			newExpr.Decs.Start = comments
		}

		existExprIndex := KeyIndex(value.Elts, key)

		if existExprIndex >= 0 {
			value.Elts[existExprIndex] = newExpr
		} else {
			// add config
			value.Elts = append(value.Elts, newExpr)
		}
	}
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
	appFilePath := support.Config.Paths.App

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

// AddImport adds an import statement to the file.
func AddImport(path string, name ...string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		node := cursor.Node().(*dst.GenDecl)

		// Check if import already exists
		for _, spec := range node.Specs {
			if importSpec, ok := spec.(*dst.ImportSpec); ok {
				if importSpec.Path.Value == strconv.Quote(path) {
					// Import already exists, no need to add
					return
				}
			}
		}

		// import spec
		im := &dst.ImportSpec{
			Path: &dst.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(path),
			},
		}
		if len(name) > 0 {
			im.Name = &dst.Ident{
				Name: name[0],
			}
		}

		// Insert third-party imports at the top and others at the bottom.
		// When formatting the source code, this helps group and sort imports
		// into stdlib, third-party, and local packages.
		if isThirdParty(path) {
			node.Specs = append([]dst.Spec{WrapNewline(im)}, node.Specs...)
			return
		}
		node.Specs = append(node.Specs, WrapNewline(im))
	}
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

func CreateImport(node dst.Node) error {
	importDecl := &dst.GenDecl{
		Tok: token.IMPORT,
	}

	f := node.(*dst.File)

	newDecls := make([]dst.Decl, 0, len(f.Decls)+1)
	newDecls = append(newDecls, f.Decls[0], importDecl) // package and import

	if len(f.Decls) > 1 {
		newDecls = append(newDecls, f.Decls[1:]...) // others
	}

	f.Decls = newDecls

	return nil
}

// Register adds a registration to the matched specified array.
func Register(expression string, before ...string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		expr := MustParseExpr(expression).(dst.Expr)
		node := cursor.Node().(*dst.CompositeLit)
		if ExprExists(node.Elts, expr) {
			color.Warningln(errors.PackageRegistrationDuplicate.Args(expression))
			return
		}
		if len(before) > 0 {
			// check if before is "*" and insert registration at the beginning
			if before[0] == "*" {
				node.Elts = slices.Insert(node.Elts, 0, expr)
				return
			}

			// check if beforeExpr is existing and insert registration before it
			beforeExpr := MustParseExpr(before[0]).(dst.Expr)
			if i := ExprIndex(node.Elts, beforeExpr); i >= 0 {
				node.Elts = slices.Insert(node.Elts, i, expr)
				return
			}

			color.Warningln(errors.PackageRegistrationNotFound.Args(before[0]))
		}

		node.Elts = append(node.Elts, expr)
	}
}

// Remove removes an expression from the matched specified function.
func Remove(expression string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		expr := MustParseExpr(expression).(dst.Expr)
		stmt := &dst.ExprStmt{
			X: expr,
		}
		node := cursor.Node().(*dst.FuncDecl)
		node.Body.List = slices.DeleteFunc(node.Body.List, func(ex dst.Stmt) bool {
			return match.EqualNode(stmt).MatchNode(ex)
		})
	}
}

// RemoveConfig removes a configuration key from the config file.
func RemoveConfig(name string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		var value *dst.CompositeLit
		switch node := cursor.Node().(type) {
		case *dst.KeyValueExpr:
			value = node.Value.(*dst.CompositeLit)
		case *dst.CallExpr:
			value = node.Args[1].(*dst.CompositeLit)
		}
		key := WrapNewline(&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(name)})

		// remove config
		value.Elts = slices.DeleteFunc(value.Elts, func(expr dst.Expr) bool {
			if kv, ok := expr.(*dst.KeyValueExpr); ok {
				return match.EqualNode(key).MatchNode(kv.Key)
			}
			return false
		})
	}
}

// RemoveImport removes an import statement from the file.
func RemoveImport(path string, name ...string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		if IsUsingImport(cursor.Parent().(*dst.File), path, name...) {
			return
		}

		node := cursor.Node().(*dst.GenDecl)
		node.Specs = slices.DeleteFunc(node.Specs, func(spec dst.Spec) bool {
			return match.Import(path, name...).MatchNode(spec)
		})
	}
}

// ReplaceConfig replaces a configuration key with the given expression in the config file.
func ReplaceConfig(name, expression string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		var value *dst.CompositeLit
		switch node := cursor.Node().(type) {
		case *dst.KeyValueExpr:
			value = node.Value.(*dst.CompositeLit)
		case *dst.CallExpr:
			value = node.Args[1].(*dst.CompositeLit)
		}
		key := WrapNewline(&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(name)})

		// replace config
		if i := KeyIndex(value.Elts, key); i >= 0 {
			value.Elts[i] = WrapNewline(&dst.KeyValueExpr{
				Key:   key,
				Value: WrapNewline(MustParseExpr(expression)).(dst.Expr),
			})
			return
		}
	}
}

// Unregister remove a registration from the matched specified array.
func Unregister(expression string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		expr := MustParseExpr(expression).(dst.Expr)
		node := cursor.Node().(*dst.CompositeLit)
		node.Elts = slices.DeleteFunc(node.Elts, func(ex dst.Expr) bool {
			return match.EqualNode(expr).MatchNode(ex)
		})
	}
}

// addMiddlewareAppendCall adds a new handler.Append() call to the function literal.
func addMiddlewareAppendCall(funcLit *dst.FuncLit, middlewareArg dst.Expr) {
	// Add newline decorations to middleware argument for proper formatting
	middlewareArg.Decorations().Before = dst.NewLine
	middlewareArg.Decorations().After = dst.NewLine

	appendStmt := &dst.ExprStmt{
		X: &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   &dst.Ident{Name: "handler"},
				Sel: &dst.Ident{Name: "Append"},
			},
			Args: []dst.Expr{middlewareArg},
			Decs: dst.CallExprDecorations{
				NodeDecs: dst.NodeDecs{
					Before: dst.NewLine,
					After:  dst.NewLine,
				},
			},
		},
	}
	funcLit.Body.List = append(funcLit.Body.List, appendStmt)
}

// addMiddlewareImports adds the required imports for middleware and configuration packages.
func addMiddlewareImports(appFilePath, pkg string) error {
	importMatchers := match.Imports()
	if err := GoFile(appFilePath).FindOrCreate(importMatchers, CreateImport).Modify(AddImport(pkg)).Apply(); err != nil {
		return err
	}

	configImportPath := "github.com/goravel/framework/contracts/foundation/configuration"
	return GoFile(appFilePath).Find(importMatchers).Modify(AddImport(configImportPath)).Apply()
}

// appendToExistingMiddleware appends middleware to an existing WithMiddleware call.
func appendToExistingMiddleware(withMiddlewareCall *dst.CallExpr, middlewareExpr dst.Expr) {
	if len(withMiddlewareCall.Args) == 0 {
		return
	}

	funcLit, ok := withMiddlewareCall.Args[0].(*dst.FuncLit)
	if !ok {
		return
	}

	appendCall := findMiddlewareAppendCall(funcLit)
	if appendCall != nil {
		// Ensure the first existing argument doesn't have a newline before it
		if len(appendCall.Args) > 0 {
			appendCall.Args[0].Decorations().Before = dst.None
		}

		// Add newline decorations to the new middleware for proper formatting
		middlewareExpr.Decorations().Before = dst.NewLine
		middlewareExpr.Decorations().After = dst.NewLine

		appendCall.Args = append(appendCall.Args, middlewareExpr)
	} else {
		addMiddlewareAppendCall(funcLit, middlewareExpr)
	}
}

// containsFoundationSetup checks if the statement contains a foundation.Setup() call.
func containsFoundationSetup(stmt *dst.ExprStmt) bool {
	var foundSetup bool
	dst.Inspect(stmt, func(n dst.Node) bool {
		if call, ok := n.(*dst.CallExpr); ok {
			if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
				if ident, ok := sel.X.(*dst.Ident); ok {
					if ident.Name == "foundation" && sel.Sel.Name == "Setup" {
						foundSetup = true
						return false
					}
				}
			}
		}
		return true
	})
	return foundSetup
}

// createWithMiddleware creates a new WithMiddleware call and inserts it into the chain.
func createWithMiddleware(setupCall *dst.CallExpr, parentOfSetup *dst.SelectorExpr, middlewareExpr dst.Expr) {
	// Add newline decorations to middleware argument for proper formatting
	middlewareExpr.Decorations().Before = dst.NewLine
	middlewareExpr.Decorations().After = dst.NewLine

	funcLit := &dst.FuncLit{
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{{Name: "handler"}},
						Type: &dst.SelectorExpr{
							X:   &dst.Ident{Name: "configuration"},
							Sel: &dst.Ident{Name: "Middleware"},
						},
					},
				},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun: &dst.SelectorExpr{
							X:   &dst.Ident{Name: "handler"},
							Sel: &dst.Ident{Name: "Append"},
						},
						Args: []dst.Expr{middlewareExpr},
						Decs: dst.CallExprDecorations{
							NodeDecs: dst.NodeDecs{
								Before: dst.NewLine,
								After:  dst.NewLine,
							},
						},
					},
				},
			},
		},
	}

	newWithMiddlewareCall := &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X: setupCall,
			Sel: &dst.Ident{
				Name: "WithMiddleware",
				Decs: dst.IdentDecorations{
					NodeDecs: dst.NodeDecs{
						Before: dst.NewLine,
					},
				},
			},
		},
		Args: []dst.Expr{funcLit},
	}

	// Insert WithMiddleware into the chain
	parentOfSetup.X = newWithMiddlewareCall
}

// findFoundationSetupCallsForMiddleware walks the chain to find Setup() and WithMiddleware() calls.
func findFoundationSetupCallsForMiddleware(callExpr *dst.CallExpr) (setupCall, withMiddlewareCall *dst.CallExpr, parentOfSetup *dst.SelectorExpr) {
	current := callExpr
	for current != nil {
		if sel, ok := current.Fun.(*dst.SelectorExpr); ok {
			if innerCall, ok := sel.X.(*dst.CallExpr); ok {
				if innerSel, ok := innerCall.Fun.(*dst.SelectorExpr); ok {
					// Check if this is the Setup() call
					if innerSel.Sel.Name == "Setup" {
						if ident, ok := innerSel.X.(*dst.Ident); ok && ident.Name == "foundation" {
							setupCall = innerCall
							parentOfSetup = sel
							break
						}
					}
					// Check if this is WithMiddleware
					if innerSel.Sel.Name == "WithMiddleware" {
						withMiddlewareCall = innerCall
					}
				}
				current = innerCall
				continue
			}
		}
		break
	}
	return
}

// findMiddlewareAppendCall finds the handler.Append() call in the function literal.
func findMiddlewareAppendCall(funcLit *dst.FuncLit) *dst.CallExpr {
	for _, stmt := range funcLit.Body.List {
		if exprStmt, ok := stmt.(*dst.ExprStmt); ok {
			if call, ok := exprStmt.X.(*dst.CallExpr); ok {
				if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
					if sel.Sel.Name == "Append" {
						return call
					}
				}
			}
		}
	}
	return nil
}

// foundationSetupMiddleware returns an action that modifies the foundation.Setup() chain.
func foundationSetupMiddleware(middleware string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		stmt := cursor.Node().(*dst.ExprStmt)

		if !containsFoundationSetup(stmt) {
			return
		}

		callExpr, ok := stmt.X.(*dst.CallExpr)
		if !ok {
			return
		}

		setupCall, withMiddlewareCall, parentOfSetup := findFoundationSetupCallsForMiddleware(callExpr)
		if setupCall == nil || parentOfSetup == nil {
			return
		}

		middlewareExpr := MustParseExpr(middleware).(dst.Expr)

		if withMiddlewareCall != nil {
			appendToExistingMiddleware(withMiddlewareCall, middlewareExpr)
		} else {
			createWithMiddleware(setupCall, parentOfSetup, middlewareExpr)
		}
	}
}
