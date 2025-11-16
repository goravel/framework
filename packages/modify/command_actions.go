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
	appFilePath := support.Config.Paths.App
	bootstrapDir := filepath.Dir(appFilePath)
	commandsFilePath := filepath.Join(bootstrapDir, "commands.go")
	commandsFileExists := supportfile.Exists(commandsFilePath)

	withCommandsExists, err := checkWithCommandsExists(appFilePath)
	if err != nil {
		return err
	}

	if !withCommandsExists {
		if commandsFileExists {
			return errors.PackageCommandsFileExists
		}

		if err := createCommandsFile(commandsFilePath); err != nil {
			return err
		}

		if err := addCommandToCommandsFile(commandsFilePath, pkg, command); err != nil {
			return err
		}

		return GoFile(appFilePath).Find(match.FoundationSetup()).Modify(foundationSetupCommandWithFunction()).Apply()
	}

	if commandsFileExists {
		if err := addCommandToCommandsFile(commandsFilePath, pkg, command); err != nil {
			return err
		}
		return nil
	}

	if err := addCommandImports(appFilePath, pkg); err != nil {
		return err
	}

	return GoFile(appFilePath).Find(match.FoundationSetup()).Modify(foundationSetupCommandInline(command)).Apply()
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

// addCommandToCommandsFile adds a command to the existing Commands() function in commands.go.
func addCommandToCommandsFile(commandsFilePath, pkg, command string) error {
	// Add the command package import
	importMatchers := match.Imports()
	if err := GoFile(commandsFilePath).FindOrCreate(importMatchers, CreateImport).Modify(AddImport(pkg)).Apply(); err != nil {
		return err
	}

	// Add the command to the Commands() function
	return GoFile(commandsFilePath).Find(match.Commands()).Modify(Register(command)).Apply()
}

// checkWithCommandsExists checks if WithCommands exists in the foundation.Setup() chain.
func checkWithCommandsExists(appFilePath string) (bool, error) {
	content, err := supportfile.GetContent(appFilePath)
	if err != nil {
		return false, err
	}

	return strings.Contains(content, "WithCommands("), nil
}

// createCommandsFile creates a new commands.go file with the Commands() function.
func createCommandsFile(commandsFilePath string) error {
	return supportfile.PutContent(commandsFilePath, commands())
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

// createWithCommands creates a new WithCommands call and inserts it into the chain.
func createWithCommands(setupCall *dst.CallExpr, parentOfSetup *dst.SelectorExpr, commandExpr dst.Expr) {
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

// foundationSetupCommandInline returns an action that modifies the foundation.Setup() chain for commands (inline array).
func foundationSetupCommandInline(command string) modify.Action {
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
			createWithCommands(setupCall, parentOfSetup, commandExpr)
		}
	}
}

// foundationSetupCommandWithFunction returns an action that adds WithCommands(Commands()) to the foundation.Setup() chain.
func foundationSetupCommandWithFunction() modify.Action {
	return func(cursor *dstutil.Cursor) {
		stmt := cursor.Node().(*dst.ExprStmt)

		if !containsFoundationSetup(stmt) {
			return
		}

		callExpr, ok := stmt.X.(*dst.CallExpr)
		if !ok {
			return
		}

		setupCall, _, parentOfSetup := findFoundationSetupCallsForCommand(callExpr)
		if setupCall == nil || parentOfSetup == nil {
			return
		}

		// Create WithCommands(Commands()) call
		newWithCommandsCall := &dst.CallExpr{
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
			Args: []dst.Expr{
				&dst.CallExpr{
					Fun: &dst.Ident{Name: "Commands"},
				},
			},
		}

		// Insert WithCommands into the chain
		parentOfSetup.X = newWithCommandsCall
	}
}
