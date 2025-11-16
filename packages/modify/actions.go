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
// If WithCommand doesn't exist, it creates one. If it exists, it appends to the command to []console.Command.
// This function also ensures the configuration package and command package are imported when creating WithCommand.
//
// Parameters:
//   - pkg: Package path of the command (e.g., "goravel/app/console/commands")
//   - command: Command expression to add (e.g., "&commands.ExampleCommand{}")
//
// Example usage:
//
//	AddCommand("goravel/app/console/commands", "&commands.ExampleCommand{}")
//
// This transforms:
//
//	foundation.Setup().WithConfig(config.Boot).Run()
//
// Into:
//
//	foundation.Setup().WithCommand(commands []console.Command{
//	  &commands.ExampleCommand{},
//	}).WithConfig(config.Boot).Run()
//
// If WithCommand already exists:
//
//	foundation.Setup().WithCommand(commands []console.Command{
//	  &commands.ExistingCommand{},
//	}).Run()
//
// It appends the new middleware:
//
//	foundation.Setup().WithCommand(commands []console.Command{
//	  &commands.ExistingCommand{},
//	  &commands.ExampleCommand{},
//	}).Run()
func AddCommand(pkg, command string) error {
	appFilePath := support.Config.Paths.App

	if err := addCommandImports(appFilePath, pkg); err != nil {
		return err
	}

	return GoFile(appFilePath).Find(match.FoundationSetup()).Modify(foundationSetupCommand(command)).Apply()
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
// If WithMiddleware doesn't exist, it creates one. If it exists, it appends to it using handler.Append().
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
// If WithMigrations doesn't exist, it creates one. If it exists, it appends the migration to []schema.Migration.
// This function also ensures the configuration package and migration package are imported when creating WithMigrations.
//
// Parameters:
//   - pkg: Package path of the migration (e.g., "goravel/database/migrations")
//   - migration: Migration expression to add (e.g., "&migrations.ExampleMigration{}")
//
// Example usage:
//
//	AddMigration("goravel/database/migrations", "&migrations.ExampleMigration{}")
//
// This transforms:
//
//	foundation.Setup().WithConfig(config.Boot).Run()
//
// Into:
//
//	foundation.Setup().WithMigrations(migrations []schema.Migration{
//	  &migrations.ExampleMigration{},
//	}).WithConfig(config.Boot).Run()
//
// If WithMigrations already exists:
//
//	foundation.Setup().WithMigrations(migrations []schema.Migration{
//	  &migrations.ExistingMigration{},
//	}).Run()
//
// It appends the new middleware:
//
//	foundation.Setup().WithMigrations(migrations []schema.Migration{
//	  &migrations.ExistingMigration{},
//	  &migrations.ExampleMigration{},
//	}).Run()
func AddMigration(pkg, migration string) error {
	appFilePath := support.Config.Paths.App

	if err := addMigrationImports(appFilePath, pkg); err != nil {
		return err
	}

	return GoFile(appFilePath).Find(match.FoundationSetup()).Modify(foundationSetupMigration(migration)).Apply()
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

		// insert registration at the end
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

// addCommandImports adds the required imports for command package and console package.
func addCommandImports(appFilePath, pkg string) error {
	importMatchers := match.Imports()
	if err := GoFile(appFilePath).FindOrCreate(importMatchers, CreateImport).Modify(AddImport(pkg)).Apply(); err != nil {
		return err
	}

	consoleImportPath := "github.com/goravel/framework/contracts/console"
	return GoFile(appFilePath).Find(importMatchers).Modify(AddImport(consoleImportPath)).Apply()
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

// appendToExistingCommand appends command to an existing WithCommand call.
func appendToExistingCommand(withCommandCall *dst.CallExpr, commandExpr dst.Expr) {
	if len(withCommandCall.Args) == 0 {
		return
	}

	compositeLit, ok := withCommandCall.Args[0].(*dst.CompositeLit)
	if !ok {
		return
	}

	// Append the command to the composite literal
	compositeLit.Elts = append(compositeLit.Elts, commandExpr)
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

// createWithCommand creates a new WithCommand call and inserts it into the chain.
func createWithCommand(setupCall *dst.CallExpr, parentOfSetup *dst.SelectorExpr, commandExpr dst.Expr) {
	compositeLit := &dst.CompositeLit{
		Type: &dst.ArrayType{
			Elt: &dst.SelectorExpr{
				X:   &dst.Ident{Name: "console"},
				Sel: &dst.Ident{Name: "Command"},
			},
		},
		Elts: []dst.Expr{commandExpr},
	}

	newWithCommandCall := &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X: setupCall,
			Sel: &dst.Ident{
				Name: "WithCommands",
				Decs: dst.IdentDecorations{
					NodeDecs: dst.NodeDecs{
						Before: dst.NewLine,
					},
				},
			},
		},
		Args: []dst.Expr{compositeLit},
	}

	// Insert WithCommand into the chain
	parentOfSetup.X = newWithCommandCall
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

// findFoundationSetupCallsForCommand walks the chain to find Setup() and WithCommand() calls.
func findFoundationSetupCallsForCommand(callExpr *dst.CallExpr) (setupCall, withCommandCall *dst.CallExpr, parentOfSetup *dst.SelectorExpr) {
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
					// Check if this is WithCommand
					if innerSel.Sel.Name == "WithCommands" {
						withCommandCall = innerCall
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

// foundationSetupCommand returns an action that modifies the foundation.Setup() chain for commands.
func foundationSetupCommand(command string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		stmt := cursor.Node().(*dst.ExprStmt)

		if !containsFoundationSetup(stmt) {
			return
		}

		callExpr, ok := stmt.X.(*dst.CallExpr)
		if !ok {
			return
		}

		setupCall, withCommandCall, parentOfSetup := findFoundationSetupCallsForCommand(callExpr)
		if setupCall == nil || parentOfSetup == nil {
			return
		}

		commandExpr := MustParseExpr(command).(dst.Expr)

		if withCommandCall != nil {
			appendToExistingCommand(withCommandCall, commandExpr)
		} else {
			createWithCommand(setupCall, parentOfSetup, commandExpr)
		}
	}
}

// foundationSetupMigration returns an action that modifies the foundation.Setup() chain for migrations.
func foundationSetupMigration(migration string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		stmt := cursor.Node().(*dst.ExprStmt)

		if !containsFoundationSetup(stmt) {
			return
		}

		callExpr, ok := stmt.X.(*dst.CallExpr)
		if !ok {
			return
		}

		setupCall, withMigrationsCall, parentOfSetup := findFoundationSetupCallsForMigration(callExpr)
		if setupCall == nil || parentOfSetup == nil {
			return
		}

		migrationExpr := MustParseExpr(migration).(dst.Expr)

		if withMigrationsCall != nil {
			appendToExistingMigration(withMigrationsCall, migrationExpr)
		} else {
			createWithMigrations(setupCall, parentOfSetup, migrationExpr)
		}
	}
}

// addMigrationImports adds the required imports for migration package and schema package.
func addMigrationImports(appFilePath, pkg string) error {
	importMatchers := match.Imports()
	if err := GoFile(appFilePath).FindOrCreate(importMatchers, CreateImport).Modify(AddImport(pkg)).Apply(); err != nil {
		return err
	}

	schemaImportPath := "github.com/goravel/framework/contracts/database/schema"
	return GoFile(appFilePath).Find(importMatchers).Modify(AddImport(schemaImportPath)).Apply()
}

// appendToExistingMigration appends migration to an existing WithMigrations call.
func appendToExistingMigration(withMigrationsCall *dst.CallExpr, migrationExpr dst.Expr) {
	if len(withMigrationsCall.Args) == 0 {
		return
	}

	compositeLit, ok := withMigrationsCall.Args[0].(*dst.CompositeLit)
	if !ok {
		return
	}

	// Append the migration to the composite literal
	compositeLit.Elts = append(compositeLit.Elts, migrationExpr)
}

// createWithMigrations creates a new WithMigrations call and inserts it into the chain.
func createWithMigrations(setupCall *dst.CallExpr, parentOfSetup *dst.SelectorExpr, migrationExpr dst.Expr) {
	compositeLit := &dst.CompositeLit{
		Type: &dst.ArrayType{
			Elt: &dst.SelectorExpr{
				X:   &dst.Ident{Name: "schema"},
				Sel: &dst.Ident{Name: "Migration"},
			},
		},
		Elts: []dst.Expr{migrationExpr},
	}

	newWithMigrationsCall := &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X: setupCall,
			Sel: &dst.Ident{
				Name: "WithMigrations",
				Decs: dst.IdentDecorations{
					NodeDecs: dst.NodeDecs{
						Before: dst.NewLine,
					},
				},
			},
		},
		Args: []dst.Expr{compositeLit},
	}

	// Insert WithMigrations into the chain
	parentOfSetup.X = newWithMigrationsCall
}

// findFoundationSetupCallsForMigration walks the chain to find Setup() and WithMigrations() calls.
func findFoundationSetupCallsForMigration(callExpr *dst.CallExpr) (setupCall, withMigrationsCall *dst.CallExpr, parentOfSetup *dst.SelectorExpr) {
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
					// Check if this is WithMigrations
					if innerSel.Sel.Name == "WithMigrations" {
						withMigrationsCall = innerCall
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
