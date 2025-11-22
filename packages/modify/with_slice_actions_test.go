package modify

import (
	"path/filepath"
	"testing"

	"github.com/dave/dst"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	packagesmatch "github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support"
	supportfile "github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path/internals"
)

type WithSliceHandlerTestSuite struct {
	suite.Suite
	bootstrapDir string
	appFile      string
}

func TestWithSliceHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(WithSliceHandlerTestSuite))
}

func (s *WithSliceHandlerTestSuite) SetupTest() {
	s.bootstrapDir = support.Config.Paths.Bootstrap
	s.appFile = internals.BootstrapApp()
}

func (s *WithSliceHandlerTestSuite) TearDownTest() {
	s.NoError(supportfile.Remove(s.bootstrapDir))
}

func (s *WithSliceHandlerTestSuite) TestNewWithSliceHandler() {
	config := withSliceConfig{
		fileName:        "commands.go",
		withMethodName:  "WithCommands",
		helperFuncName:  "Commands",
		typePackage:     "console",
		typeName:        "Command",
		typeImportPath:  "github.com/goravel/framework/contracts/console",
		fileExistsError: errors.PackageCommandsFileExists,
		stubTemplate:    commands,
		matcherFunc:     packagesmatch.Commands,
	}

	// Create app.go file
	s.Require().NoError(supportfile.PutContent(s.appFile, `package bootstrap

import "github.com/goravel/framework/foundation"

func Boot() {
	foundation.Setup().Run()
}
`))

	handler := newWithSliceHandler(config)

	s.NotNil(handler)
	s.Equal(config.fileName, handler.config.fileName)
	s.Equal(config.withMethodName, handler.config.withMethodName)
	s.Equal(config.helperFuncName, handler.config.helperFuncName)
	s.Equal(config.typePackage, handler.config.typePackage)
	s.Equal(config.typeName, handler.config.typeName)
	s.Equal(config.typeImportPath, handler.config.typeImportPath)
	s.Equal(config.fileExistsError, handler.config.fileExistsError)
	s.NotNil(handler.config.stubTemplate)
	s.NotNil(handler.config.matcherFunc)
	s.Equal(s.appFile, handler.appFilePath)
	s.Contains(handler.filePath, filepath.Join(s.bootstrapDir, "commands.go"))
	s.False(handler.fileExists)
}

func (s *WithSliceHandlerTestSuite) TestNewWithSliceHandler_FileExists() {
	config := withSliceConfig{
		fileName:        "commands.go",
		withMethodName:  "WithCommands",
		helperFuncName:  "Commands",
		typePackage:     "console",
		typeName:        "Command",
		typeImportPath:  "github.com/goravel/framework/contracts/console",
		fileExistsError: errors.PackageCommandsFileExists,
		stubTemplate:    commands,
		matcherFunc:     packagesmatch.Commands,
	}

	// Create both app.go and commands.go
	s.Require().NoError(supportfile.PutContent(s.appFile, `package bootstrap

import "github.com/goravel/framework/foundation"

func Boot() {
	foundation.Setup().Run()
}
`))

	commandsFile := filepath.Join(s.bootstrapDir, "commands.go")
	s.Require().NoError(supportfile.PutContent(commandsFile, `package bootstrap

import "github.com/goravel/framework/contracts/console"

func Commands() []console.Command {
	return []console.Command{}
}
`))

	handler := newWithSliceHandler(config)

	s.NotNil(handler)
	s.True(handler.fileExists)
}

func (s *WithSliceHandlerTestSuite) TestAddItem_NoWithMethod_NoFile() {
	config := withSliceConfig{
		fileName:        "commands.go",
		withMethodName:  "WithCommands",
		helperFuncName:  "Commands",
		typePackage:     "console",
		typeName:        "Command",
		typeImportPath:  "github.com/goravel/framework/contracts/console",
		fileExistsError: errors.PackageCommandsFileExists,
		stubTemplate:    commands,
		matcherFunc:     packagesmatch.Commands,
	}

	appContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	err := handler.AddItem("goravel/app/console/commands", "&commands.ExampleCommand{}")

	s.NoError(err)

	// Verify app.go was updated
	appResult, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.Contains(appResult, "WithCommands(Commands())")

	// Verify commands.go was created
	commandsFile := filepath.Join(s.bootstrapDir, "commands.go")
	s.True(supportfile.Exists(commandsFile))

	commandsResult, err := supportfile.GetContent(commandsFile)
	s.NoError(err)
	s.Contains(commandsResult, "&commands.ExampleCommand{}")
}

func (s *WithSliceHandlerTestSuite) TestAddItem_NoWithMethod_FileExists() {
	config := withSliceConfig{
		fileName:        "commands.go",
		withMethodName:  "WithCommands",
		helperFuncName:  "Commands",
		typePackage:     "console",
		typeName:        "Command",
		typeImportPath:  "github.com/goravel/framework/contracts/console",
		fileExistsError: errors.PackageCommandsFileExists,
		stubTemplate:    commands,
		matcherFunc:     packagesmatch.Commands,
	}

	appContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	// Create commands.go file
	commandsFile := filepath.Join(s.bootstrapDir, "commands.go")
	s.Require().NoError(supportfile.PutContent(commandsFile, `package bootstrap

import "github.com/goravel/framework/contracts/console"

func Commands() []console.Command {
	return []console.Command{}
}
`))

	handler := newWithSliceHandler(config)
	err := handler.AddItem("goravel/app/console/commands", "&commands.ExampleCommand{}")

	s.Error(err)
	s.Equal(errors.PackageCommandsFileExists, err)
}

func (s *WithSliceHandlerTestSuite) TestAddItem_WithMethodExists_FileExists() {
	config := withSliceConfig{
		fileName:        "commands.go",
		withMethodName:  "WithCommands",
		helperFuncName:  "Commands",
		typePackage:     "console",
		typeName:        "Command",
		typeImportPath:  "github.com/goravel/framework/contracts/console",
		fileExistsError: errors.PackageCommandsFileExists,
		stubTemplate:    commands,
		matcherFunc:     packagesmatch.Commands,
	}

	appContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithCommands(Commands()).WithConfig(config.Boot).Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	// Create commands.go with existing command
	commandsFile := filepath.Join(s.bootstrapDir, "commands.go")
	s.Require().NoError(supportfile.PutContent(commandsFile, `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
	}
}
`))

	handler := newWithSliceHandler(config)
	err := handler.AddItem("goravel/app/console/commands", "&commands.NewCommand{}")

	s.NoError(err)

	// Verify commands.go was updated
	commandsResult, err := supportfile.GetContent(commandsFile)
	s.NoError(err)
	s.Contains(commandsResult, "&commands.ExistingCommand{}")
	s.Contains(commandsResult, "&commands.NewCommand{}")
}

func (s *WithSliceHandlerTestSuite) TestAddItem_WithMethodExists_NoFile_InlineArray() {
	config := withSliceConfig{
		fileName:        "commands.go",
		withMethodName:  "WithCommands",
		helperFuncName:  "Commands",
		typePackage:     "console",
		typeName:        "Command",
		typeImportPath:  "github.com/goravel/framework/contracts/console",
		fileExistsError: errors.PackageCommandsFileExists,
		stubTemplate:    commands,
		matcherFunc:     packagesmatch.Commands,
	}

	appContent := `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands([]console.Command{
			&commands.ExistingCommand{},
		}).WithConfig(config.Boot).Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	err := handler.AddItem("goravel/app/console/commands", "&commands.NewCommand{}")

	s.NoError(err)

	// Verify app.go was updated with inline addition
	appResult, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.Contains(appResult, "&commands.ExistingCommand{}")
	s.Contains(appResult, "&commands.NewCommand{}")
}

func (s *WithSliceHandlerTestSuite) TestCheckWithMethodExists() {
	config := withSliceConfig{
		fileName:       "commands.go",
		withMethodName: "WithCommands",
	}

	// Test when method exists
	appContent := `package bootstrap

import "github.com/goravel/framework/foundation"

func Boot() {
	foundation.Setup().WithCommands(Commands()).Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	exists, err := handler.checkWithMethodExists()

	s.NoError(err)
	s.True(exists)

	// Test when method doesn't exist
	appContent = `package bootstrap

import "github.com/goravel/framework/foundation"

func Boot() {
	foundation.Setup().Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler = newWithSliceHandler(config)
	exists, err = handler.checkWithMethodExists()

	s.NoError(err)
	s.False(exists)
}

func (s *WithSliceHandlerTestSuite) TestCreateFile() {
	config := withSliceConfig{
		fileName:     "commands.go",
		stubTemplate: commands,
	}

	s.Require().NoError(supportfile.PutContent(s.appFile, `package bootstrap`))

	handler := newWithSliceHandler(config)
	err := handler.createFile()

	s.NoError(err)

	commandsFile := filepath.Join(s.bootstrapDir, "commands.go")
	s.True(supportfile.Exists(commandsFile))

	content, err := supportfile.GetContent(commandsFile)
	s.NoError(err)
	s.Equal(commands(), content)
}

func (s *WithSliceHandlerTestSuite) TestAddImports() {
	config := withSliceConfig{
		fileName:       "commands.go",
		typeImportPath: "github.com/goravel/framework/contracts/console",
	}

	appContent := `package bootstrap

import "github.com/goravel/framework/foundation"

func Boot() {
	foundation.Setup().Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	err := handler.addImports("goravel/app/console/commands")

	s.NoError(err)

	content, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.Contains(content, `"goravel/app/console/commands"`)
	s.Contains(content, `"github.com/goravel/framework/contracts/console"`)
}

func (s *WithSliceHandlerTestSuite) TestAddItemToFile() {
	config := withSliceConfig{
		fileName:    "commands.go",
		matcherFunc: packagesmatch.Commands,
	}

	s.Require().NoError(supportfile.PutContent(s.appFile, `package bootstrap`))

	commandsFile := filepath.Join(s.bootstrapDir, "commands.go")
	commandsContent := `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
)

func Commands() []console.Command {
	return []console.Command{}
}
`
	s.Require().NoError(supportfile.PutContent(commandsFile, commandsContent))

	handler := newWithSliceHandler(config)
	err := handler.addItemToFile("goravel/app/console/commands", "&commands.ExampleCommand{}")

	s.NoError(err)

	content, err := supportfile.GetContent(commandsFile)
	s.NoError(err)
	s.Contains(content, "&commands.ExampleCommand{}")
	s.Contains(content, `"goravel/app/console/commands"`)
}

func (s *WithSliceHandlerTestSuite) TestAppendToExisting() {
	config := withSliceConfig{
		typePackage: "console",
		typeName:    "Command",
	}

	handler := newWithSliceHandler(config)

	// Test with valid composite literal
	withCall := &dst.CallExpr{
		Args: []dst.Expr{
			&dst.CompositeLit{
				Elts: []dst.Expr{
					&dst.Ident{Name: "existing"},
				},
			},
		},
	}

	itemExpr := &dst.Ident{Name: "new"}
	handler.appendToExisting(withCall, itemExpr)

	compositeLit := withCall.Args[0].(*dst.CompositeLit)
	s.Len(compositeLit.Elts, 2)
	s.Equal("existing", compositeLit.Elts[0].(*dst.Ident).Name)
	s.Equal("new", compositeLit.Elts[1].(*dst.Ident).Name)
}

func (s *WithSliceHandlerTestSuite) TestAppendToExisting_EmptyArgs() {
	config := withSliceConfig{
		typePackage: "console",
		typeName:    "Command",
	}

	handler := newWithSliceHandler(config)

	// Test with empty args
	withCall := &dst.CallExpr{
		Args: []dst.Expr{},
	}

	itemExpr := &dst.Ident{Name: "new"}
	handler.appendToExisting(withCall, itemExpr)

	// Should not panic and should not add anything
	s.Len(withCall.Args, 0)
}

func (s *WithSliceHandlerTestSuite) TestAppendToExisting_NotCompositeLit() {
	config := withSliceConfig{
		typePackage: "console",
		typeName:    "Command",
	}

	handler := newWithSliceHandler(config)

	// Test with non-composite literal arg
	withCall := &dst.CallExpr{
		Args: []dst.Expr{
			&dst.Ident{Name: "notACompositeLit"},
		},
	}

	itemExpr := &dst.Ident{Name: "new"}
	handler.appendToExisting(withCall, itemExpr)

	// Should not panic and should not modify
	s.Len(withCall.Args, 1)
	s.Equal("notACompositeLit", withCall.Args[0].(*dst.Ident).Name)
}

func (s *WithSliceHandlerTestSuite) TestCreateWithMethod() {
	config := withSliceConfig{
		withMethodName: "WithCommands",
		typePackage:    "console",
		typeName:       "Command",
	}

	handler := newWithSliceHandler(config)

	setupCall := &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X:   &dst.Ident{Name: "foundation"},
			Sel: &dst.Ident{Name: "Setup"},
		},
	}

	parentOfSetup := &dst.SelectorExpr{
		X:   setupCall,
		Sel: &dst.Ident{Name: "Run"},
	}

	itemExpr := &dst.Ident{Name: "&commands.NewCommand{}"}

	handler.createWithMethod(setupCall, parentOfSetup, itemExpr)

	// Verify the chain was modified
	newWithCall, ok := parentOfSetup.X.(*dst.CallExpr)
	s.True(ok)
	s.NotNil(newWithCall)

	withSel, ok := newWithCall.Fun.(*dst.SelectorExpr)
	s.True(ok)
	s.Equal("WithCommands", withSel.Sel.Name)

	// Verify the composite literal
	s.Len(newWithCall.Args, 1)
	compositeLit, ok := newWithCall.Args[0].(*dst.CompositeLit)
	s.True(ok)
	s.Len(compositeLit.Elts, 1)
}

func (s *WithSliceHandlerTestSuite) TestFindFoundationSetupCalls() {
	config := withSliceConfig{
		withMethodName: "WithCommands",
	}

	handler := newWithSliceHandler(config)

	// Create a chain: foundation.Setup().WithCommands(...).Run()
	setupCall := &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X:   &dst.Ident{Name: "foundation"},
			Sel: &dst.Ident{Name: "Setup"},
		},
	}

	withCall := &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X:   setupCall,
			Sel: &dst.Ident{Name: "WithCommands"},
		},
	}

	runCall := &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X:   withCall,
			Sel: &dst.Ident{Name: "Run"},
		},
	}

	foundSetup, foundWith, parentOfSetup := handler.findFoundationSetupCalls(runCall)

	s.NotNil(foundSetup)
	s.NotNil(foundWith)
	s.NotNil(parentOfSetup)
	s.Equal(setupCall, foundSetup)
	s.Equal(withCall, foundWith)
}

func (s *WithSliceHandlerTestSuite) TestFindFoundationSetupCalls_NoWithMethod() {
	config := withSliceConfig{
		withMethodName: "WithCommands",
	}

	handler := newWithSliceHandler(config)

	// Create a chain without WithCommands: foundation.Setup().Run()
	setupCall := &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X:   &dst.Ident{Name: "foundation"},
			Sel: &dst.Ident{Name: "Setup"},
		},
	}

	runCall := &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X:   setupCall,
			Sel: &dst.Ident{Name: "Run"},
		},
	}

	foundSetup, foundWith, parentOfSetup := handler.findFoundationSetupCalls(runCall)

	s.NotNil(foundSetup)
	s.Nil(foundWith)
	s.NotNil(parentOfSetup)
	s.Equal(setupCall, foundSetup)
}

func (s *WithSliceHandlerTestSuite) TestSetupInline() {
	config := withSliceConfig{
		withMethodName: "WithCommands",
		typePackage:    "console",
		typeName:       "Command",
	}

	appContent := `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
)

func Boot() {
	foundation.Setup().
		WithCommands([]console.Command{
			&commands.ExistingCommand{},
		}).Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	action := handler.setupInline("&commands.NewCommand{}")

	err := GoFile(s.appFile).Find(packagesmatch.FoundationSetup()).Modify(action).Apply()
	s.NoError(err)

	result, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.Contains(result, "&commands.ExistingCommand{}")
	s.Contains(result, "&commands.NewCommand{}")
}

func (s *WithSliceHandlerTestSuite) TestSetupInline_CreateWithMethod() {
	config := withSliceConfig{
		withMethodName: "WithCommands",
		typePackage:    "console",
		typeName:       "Command",
	}

	appContent := `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
)

func Boot() {
	foundation.Setup().Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	action := handler.setupInline("&commands.NewCommand{}")

	err := GoFile(s.appFile).Find(packagesmatch.FoundationSetup()).Modify(action).Apply()
	s.NoError(err)

	result, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.Contains(result, "WithCommands([]console.Command{")
	s.Contains(result, "&commands.NewCommand{}")
}

func (s *WithSliceHandlerTestSuite) TestSetupWithFunction() {
	config := withSliceConfig{
		withMethodName: "WithCommands",
		helperFuncName: "Commands",
	}

	appContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	action := handler.setupWithFunction()

	err := GoFile(s.appFile).Find(packagesmatch.FoundationSetup()).Modify(action).Apply()
	s.NoError(err)

	result, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.Contains(result, "WithCommands(Commands())")
}

func (s *WithSliceHandlerTestSuite) TestAddItem_WithMigrations() {
	config := withSliceConfig{
		fileName:        "migrations.go",
		withMethodName:  "WithMigrations",
		helperFuncName:  "Migrations",
		typePackage:     "schema",
		typeName:        "Migration",
		typeImportPath:  "github.com/goravel/framework/contracts/database/schema",
		fileExistsError: errors.PackageMigrationsFileExists,
		stubTemplate:    migrations,
		matcherFunc:     packagesmatch.Migrations,
	}

	appContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	err := handler.AddItem("goravel/database/migrations", "&migrations.CreateUsersTable{}")

	s.NoError(err)

	// Verify app.go was updated
	appResult, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.Contains(appResult, "WithMigrations(Migrations())")

	// Verify migrations.go was created
	migrationsFile := filepath.Join(s.bootstrapDir, "migrations.go")
	s.True(supportfile.Exists(migrationsFile))

	migrationsResult, err := supportfile.GetContent(migrationsFile)
	s.NoError(err)
	s.Contains(migrationsResult, "&migrations.CreateUsersTable{}")
}
