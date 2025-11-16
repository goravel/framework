package modify

import (
	"path/filepath"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support"
	supportfile "github.com/goravel/framework/support/file"
)

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
	appFilePath := support.Config.Paths.App
	bootstrapDir := filepath.Dir(appFilePath)
	migrationsFilePath := filepath.Join(bootstrapDir, "migrations.go")
	migrationsFileExists := supportfile.Exists(migrationsFilePath)

	withMigrationsExists, err := checkWithMigrationsExists(appFilePath)
	if err != nil {
		return err
	}

	if !withMigrationsExists {
		if migrationsFileExists {
			return errors.PackageMigrationsFileExists
		}

		if err := createMigrationsFile(migrationsFilePath); err != nil {
			return err
		}

		if err := addMigrationToMigrationsFile(migrationsFilePath, pkg, migration); err != nil {
			return err
		}

		return GoFile(appFilePath).Find(match.FoundationSetup()).Modify(foundationSetupMigrationWithFunction()).Apply()
	}

	if migrationsFileExists {
		if err := addMigrationToMigrationsFile(migrationsFilePath, pkg, migration); err != nil {
			return err
		}
		return nil
	}

	if err := addMigrationImports(appFilePath, pkg); err != nil {
		return err
	}

	return GoFile(appFilePath).Find(match.FoundationSetup()).Modify(foundationSetupMigrationInline(migration)).Apply()
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

// addMigrationToMigrationsFile adds a migration to the existing Migrations() function in migrations.go.
func addMigrationToMigrationsFile(migrationsFilePath, pkg, migration string) error {
	// Add the migration package import
	importMatchers := match.Imports()
	if err := GoFile(migrationsFilePath).FindOrCreate(importMatchers, CreateImport).Modify(AddImport(pkg)).Apply(); err != nil {
		return err
	}

	// Add the migration to the Migrations() function
	return GoFile(migrationsFilePath).Find(match.Migrations()).Modify(Register(migration)).Apply()
}

// checkWithMigrationsExists checks if WithMigrations exists in the foundation.Setup() chain.
func checkWithMigrationsExists(appFilePath string) (bool, error) {
	content, err := supportfile.GetContent(appFilePath)
	if err != nil {
		return false, err
	}

	// Simple string check - if WithMigrations appears in the chain, it exists
	return strings.Contains(content, "WithMigrations("), nil
}

// createMigrationsFile creates a new migrations.go file with the Migrations() function.
func createMigrationsFile(migrationsFilePath string) error {
	return supportfile.PutContent(migrationsFilePath, migrations())
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

// foundationSetupMigrationInline returns an action that modifies the foundation.Setup() chain for migrations (inline array).
func foundationSetupMigrationInline(migration string) modify.Action {
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

// foundationSetupMigrationWithFunction returns an action that adds WithMigrations(Migrations()) to the foundation.Setup() chain.
func foundationSetupMigrationWithFunction() modify.Action {
	return func(cursor *dstutil.Cursor) {
		stmt := cursor.Node().(*dst.ExprStmt)

		if !containsFoundationSetup(stmt) {
			return
		}

		callExpr, ok := stmt.X.(*dst.CallExpr)
		if !ok {
			return
		}

		setupCall, _, parentOfSetup := findFoundationSetupCallsForMigration(callExpr)
		if setupCall == nil || parentOfSetup == nil {
			return
		}

		// Create WithMigrations(Migrations()) call
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
			Args: []dst.Expr{
				&dst.CallExpr{
					Fun: &dst.Ident{Name: "Migrations"},
				},
			},
		}

		// Insert WithMigrations into the chain
		parentOfSetup.X = newWithMigrationsCall
	}
}
