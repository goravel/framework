package modify

import (
	"go/token"
	"path/filepath"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support"
	supportfile "github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
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
	s.appFile = path.Bootstrap("app.go")
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
		matcherFunc:     match.Commands,
	}

	s.Run("FileDoesNotExist", func() {
		// Create app.go file
		s.Require().NoError(supportfile.PutContent(s.appFile, `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
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
	})

	s.Run("FileExists", func() {
		// Create both app.go and commands.go
		s.Require().NoError(supportfile.PutContent(s.appFile, `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
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
	})
}

func (s *WithSliceHandlerTestSuite) TestAddItem_NoFoundationSetup() {
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

	// Create app.go file WITHOUT foundation.Setup()
	s.Require().NoError(supportfile.PutContent(s.appFile, `package bootstrap

import "goravel/config"

func Boot() {
	config.Boot()
}
`))

	handler := newWithSliceHandler(config)
	err := handler.AddItem("goravel/app/console/commands", "&commands.ExampleCommand{}")

	// Should return nil (no-op) when foundation.Setup() doesn't exist
	s.NoError(err)

	// Verify app.go was NOT modified
	appResult, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.NotContains(appResult, "WithCommands")

	// Verify commands.go was NOT created
	commandsFile := filepath.Join(s.bootstrapDir, "commands.go")
	s.False(supportfile.Exists(commandsFile))
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
		matcherFunc:     match.Commands,
	}

	appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	err := handler.AddItem("goravel/app/console/commands", "&commands.ExampleCommand{}")

	s.NoError(err)

	// Verify app.go was updated
	appResult, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.Contains(appResult, "WithCommands(Commands)")

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
		matcherFunc:     match.Commands,
	}

	appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
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
		matcherFunc:     match.Commands,
	}

	appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithCommands(Commands).WithConfig(config.Boot).Start()
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
		matcherFunc:     match.Commands,
	}

	appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(func() []console.Command{
			return []console.Command{
				&commands.ExistingCommand{},
			}
		}).WithConfig(config.Boot).Start()
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

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithCommands(Commands).Start()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	exists, err := handler.checkWithMethodExists()

	s.NoError(err)
	s.True(exists)

	// Test when method doesn't exist
	appContent = `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
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

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
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
		matcherFunc: match.Commands,
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

	s.Run("ValidCompositeLiteral", func() {
		handler := newWithSliceHandler(config)

		// Test with valid function literal containing a return statement with composite literal
		withCall := &dst.CallExpr{
			Args: []dst.Expr{
				&dst.FuncLit{
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ReturnStmt{
								Results: []dst.Expr{
									&dst.CompositeLit{
										Elts: []dst.Expr{
											&dst.Ident{Name: "existing"},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		itemExpr := &dst.Ident{Name: "new"}
		handler.appendToExisting(withCall, itemExpr)

		funcLit := withCall.Args[0].(*dst.FuncLit)
		retStmt := funcLit.Body.List[0].(*dst.ReturnStmt)
		compositeLit := retStmt.Results[0].(*dst.CompositeLit)
		s.Len(compositeLit.Elts, 2)
		s.Equal("existing", compositeLit.Elts[0].(*dst.Ident).Name)
		s.Equal("new", compositeLit.Elts[1].(*dst.Ident).Name)
	})

	s.Run("EmptyArgs", func() {
		handler := newWithSliceHandler(config)

		// Test with empty args
		withCall := &dst.CallExpr{
			Args: []dst.Expr{},
		}

		itemExpr := &dst.Ident{Name: "new"}
		handler.appendToExisting(withCall, itemExpr)

		// Should not panic and should not add anything
		s.Len(withCall.Args, 0)
	})

	s.Run("NotCompositeLit", func() {
		handler := newWithSliceHandler(config)

		// Test with non-function literal arg
		withCall := &dst.CallExpr{
			Args: []dst.Expr{
				&dst.Ident{Name: "notAFuncLit"},
			},
		}

		itemExpr := &dst.Ident{Name: "new"}
		handler.appendToExisting(withCall, itemExpr)

		// Should not panic and should not modify
		s.Len(withCall.Args, 1)
		s.Equal("notAFuncLit", withCall.Args[0].(*dst.Ident).Name)
	})
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

	// Verify the function literal wrapping
	s.Len(newWithCall.Args, 1)
	funcLit, ok := newWithCall.Args[0].(*dst.FuncLit)
	s.True(ok)
	s.NotNil(funcLit)

	// Verify the return statement contains the composite literal
	s.Len(funcLit.Body.List, 1)
	retStmt, ok := funcLit.Body.List[0].(*dst.ReturnStmt)
	s.True(ok)
	s.Len(retStmt.Results, 1)
	compositeLit, ok := retStmt.Results[0].(*dst.CompositeLit)
	s.True(ok)
	s.Len(compositeLit.Elts, 1)
}

func (s *WithSliceHandlerTestSuite) TestFindFoundationSetupCalls() {
	config := withSliceConfig{
		withMethodName: "WithCommands",
	}

	s.Run("WithMethodExists", func() {
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
	})

	s.Run("NoWithMethod", func() {
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
	})
}

func (s *WithSliceHandlerTestSuite) TestSetupInline() {
	config := withSliceConfig{
		withMethodName: "WithCommands",
		typePackage:    "console",
		typeName:       "Command",
	}

	s.Run("AppendToExisting", func() {
		appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(func() []console.Command{
			return []console.Command{
				&commands.ExistingCommand{},
			}
		}).Start()
}
`
		s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

		handler := newWithSliceHandler(config)
		action := handler.setupInline("&commands.NewCommand{}")

		err := GoFile(s.appFile).Find(match.FoundationSetup()).Modify(action).Apply()
		s.NoError(err)

		result, err := supportfile.GetContent(s.appFile)
		s.NoError(err)
		s.Contains(result, "&commands.ExistingCommand{}")
		s.Contains(result, "&commands.NewCommand{}")
	})

	s.Run("CreateWithMethod", func() {
		appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
}
`
		s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

		handler := newWithSliceHandler(config)
		action := handler.setupInline("&commands.NewCommand{}")

		err := GoFile(s.appFile).Find(match.FoundationSetup()).Modify(action).Apply()
		s.NoError(err)

		result, err := supportfile.GetContent(s.appFile)
		s.NoError(err)
		s.Contains(result, "WithCommands(func() []console.Command {")
		s.Contains(result, "&commands.NewCommand{}")
	})
}

func (s *WithSliceHandlerTestSuite) TestSetupWithFunction() {
	config := withSliceConfig{
		withMethodName: "WithCommands",
		helperFuncName: "Commands",
	}

	appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	action := handler.setupWithFunction()

	err := GoFile(s.appFile).Find(match.FoundationSetup()).Modify(action).Apply()
	s.NoError(err)

	result, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.Contains(result, "WithCommands(Commands)")
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
		matcherFunc:     match.Migrations,
	}

	appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	err := handler.AddItem("goravel/database/migrations", "&migrations.CreateUsersTable{}")

	s.NoError(err)

	// Verify app.go was updated
	appResult, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.Contains(appResult, "WithMigrations(Migrations)")

	// Verify migrations.go was created
	migrationsFile := filepath.Join(s.bootstrapDir, "migrations.go")
	s.True(supportfile.Exists(migrationsFile))

	migrationsResult, err := supportfile.GetContent(migrationsFile)
	s.NoError(err)
	s.Contains(migrationsResult, "&migrations.CreateUsersTable{}")
}

func (s *WithSliceHandlerTestSuite) TestRemoveItem() {
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

	s.Run("NoWithMethod", func() {
		defer func() {
			s.NoError(supportfile.Remove(s.bootstrapDir))
		}()

		appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`
		s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

		handler := newWithSliceHandler(config)
		err := handler.RemoveItem("goravel/app/console/commands", "&commands.ExampleCommand{}")

		// Should return nil when WithMethod doesn't exist
		s.NoError(err)
	})

	s.Run("WithMethodExists_FileExists", func() {
		defer func() {
			s.NoError(supportfile.Remove(s.bootstrapDir))
		}()

		appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithCommands(Commands).WithConfig(config.Boot).Start()
}
`
		s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

		// Create commands.go with two commands
		commandsFile := filepath.Join(s.bootstrapDir, "commands.go")
		s.Require().NoError(supportfile.PutContent(commandsFile, `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExampleCommand{},
		&commands.OtherCommand{},
	}
}
`))

		handler := newWithSliceHandler(config)
		err := handler.RemoveItem("goravel/app/console/commands", "&commands.ExampleCommand{}")

		s.NoError(err)

		// Verify commands.go was updated - ExampleCommand removed
		commandsResult, err := supportfile.GetContent(commandsFile)
		s.NoError(err)
		s.NotContains(commandsResult, "&commands.ExampleCommand{}")
		s.Contains(commandsResult, "&commands.OtherCommand{}")
	})

	s.Run("WithMethodExists_NoFile_InlineArray", func() {
		defer func() {
			s.NoError(supportfile.Remove(s.bootstrapDir))
		}()

		appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(func() []console.Command {
			return []console.Command{
				&commands.ExampleCommand{},
				&commands.OtherCommand{},
			}
		}).WithConfig(config.Boot).Start()
}
`
		s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

		handler := newWithSliceHandler(config)
		err := handler.RemoveItem("goravel/app/console/commands", "&commands.ExampleCommand{}")

		s.NoError(err)

		// Verify app.go was updated - ExampleCommand removed from inline array
		appResult, err := supportfile.GetContent(s.appFile)
		s.NoError(err)
		s.NotContains(appResult, "&commands.ExampleCommand{}")
		s.Contains(appResult, "&commands.OtherCommand{}")
	})

	s.Run("LastItemInFile", func() {
		defer func() {
			s.NoError(supportfile.Remove(s.bootstrapDir))
		}()

		appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithCommands(Commands).WithConfig(config.Boot).Start()
}
`
		s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

		// Create commands.go with single command
		commandsFile := filepath.Join(s.bootstrapDir, "commands.go")
		s.Require().NoError(supportfile.PutContent(commandsFile, `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExampleCommand{},
	}
}
`))

		handler := newWithSliceHandler(config)
		err := handler.RemoveItem("goravel/app/console/commands", "&commands.ExampleCommand{}")

		s.NoError(err)

		// Verify commands.go was updated - command removed and import cleaned up
		commandsResult, err := supportfile.GetContent(commandsFile)
		s.NoError(err)
		s.NotContains(commandsResult, "&commands.ExampleCommand{}")
	})

	s.Run("CleansUpImport", func() {
		defer func() {
			s.NoError(supportfile.Remove(s.bootstrapDir))
		}()

		appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(func() []console.Command {
			return []console.Command{
				&commands.OnlyCommand{},
			}
		}).WithConfig(config.Boot).Start()
}
`
		s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

		handler := newWithSliceHandler(config)
		err := handler.RemoveItem("goravel/app/console/commands", "&commands.OnlyCommand{}")

		s.NoError(err)

		// Verify the item was removed and import was cleaned up
		appResult, err := supportfile.GetContent(s.appFile)
		s.NoError(err)
		s.NotContains(appResult, "&commands.OnlyCommand{}")
		s.NotContains(appResult, `"goravel/app/console/commands"`)
	})

	s.Run("NoFoundationSetup", func() {
		defer func() {
			s.NoError(supportfile.Remove(s.bootstrapDir))
		}()

		// Create app.go file WITHOUT foundation.Setup()
		appContent := `package bootstrap

import "goravel/config"

func Boot() {
	config.Boot()
}
`
		s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

		handler := newWithSliceHandler(config)
		err := handler.RemoveItem("goravel/app/console/commands", "&commands.ExampleCommand{}")

		// Should return nil (no-op) when foundation.Setup() doesn't exist
		s.NoError(err)

		// Verify app.go was NOT modified
		appResult, err := supportfile.GetContent(s.appFile)
		s.NoError(err)
		s.Equal(appContent, appResult)
	})
}

func (s *WithSliceHandlerTestSuite) TestRemoveImports() {
	config := withSliceConfig{
		fileName:       "commands.go",
		typeImportPath: "github.com/goravel/framework/contracts/console",
	}

	appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/app/console/unused"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(func() []console.Command {
			return []console.Command{
				&commands.OtherCommand{},
			}
		}).WithConfig(config.Boot).Start()
}
`
	s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

	handler := newWithSliceHandler(config)
	err := handler.removeImports("goravel/app/console/unused")

	s.NoError(err)

	// Verify the unused import is removed
	content, err := supportfile.GetContent(s.appFile)
	s.NoError(err)
	s.NotContains(content, "goravel/app/console/unused")
	// Verify used imports are still present
	s.Contains(content, "goravel/app/console/commands")
}

func (s *WithSliceHandlerTestSuite) TestRemoveItemFromFile() {
	config := withSliceConfig{
		fileName:    "commands.go",
		matcherFunc: match.Commands,
	}

	s.Require().NoError(supportfile.PutContent(s.appFile, `package bootstrap`))

	commandsFile := filepath.Join(s.bootstrapDir, "commands.go")
	commandsContent := `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.Command1{},
		&commands.Command2{},
		&commands.Command3{},
	}
}
`
	s.Require().NoError(supportfile.PutContent(commandsFile, commandsContent))

	handler := newWithSliceHandler(config)
	err := handler.removeItemFromFile("goravel/app/console/commands", "&commands.Command2{}")

	s.NoError(err)

	content, err := supportfile.GetContent(commandsFile)
	s.NoError(err)
	s.Contains(content, "&commands.Command1{}")
	s.NotContains(content, "&commands.Command2{}")
	s.Contains(content, "&commands.Command3{}")
}

func (s *WithSliceHandlerTestSuite) TestRemoveInline() {
	config := withSliceConfig{
		withMethodName: "WithCommands",
		typePackage:    "console",
		typeName:       "Command",
	}

	s.Run("RemoveFromInlineArray", func() {
		appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(func() []console.Command {
			return []console.Command{
				&commands.Command1{},
				&commands.Command2{},
				&commands.Command3{},
			}
		}).Start()
}
`
		s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

		handler := newWithSliceHandler(config)
		action := handler.removeInline("&commands.Command2{}")

		err := GoFile(s.appFile).Find(match.FoundationSetup()).Modify(action).Apply()
		s.NoError(err)

		result, err := supportfile.GetContent(s.appFile)
		s.NoError(err)
		s.Contains(result, "&commands.Command1{}")
		s.NotContains(result, "&commands.Command2{}")
		s.Contains(result, "&commands.Command3{}")
	})

	s.Run("NoWithMethod", func() {
		appContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`
		s.Require().NoError(supportfile.PutContent(s.appFile, appContent))

		handler := newWithSliceHandler(config)
		action := handler.removeInline("&commands.Command1{}")

		// Should not panic when WithMethod doesn't exist
		err := GoFile(s.appFile).Find(match.FoundationSetup()).Modify(action).Apply()
		s.NoError(err)

		result, err := supportfile.GetContent(s.appFile)
		s.NoError(err)
		// Content should remain unchanged
		s.Contains(result, "foundation.Setup().WithConfig(config.Boot).Start()")
	})
}

func (s *WithSliceHandlerTestSuite) TestRemoveFromExisting() {
	config := withSliceConfig{
		typePackage: "console",
		typeName:    "Command",
	}

	s.Run("RemoveMiddleItem", func() {
		handler := newWithSliceHandler(config)

		// Test with valid function literal containing a return statement with composite literal
		withCall := &dst.CallExpr{
			Args: []dst.Expr{
				&dst.FuncLit{
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ReturnStmt{
								Results: []dst.Expr{
									&dst.CompositeLit{
										Elts: []dst.Expr{
											&dst.UnaryExpr{
												Op: token.AND,
												X: &dst.CompositeLit{
													Type: &dst.SelectorExpr{
														X:   &dst.Ident{Name: "commands"},
														Sel: &dst.Ident{Name: "Command1"},
													},
												},
											},
											&dst.UnaryExpr{
												Op: token.AND,
												X: &dst.CompositeLit{
													Type: &dst.SelectorExpr{
														X:   &dst.Ident{Name: "commands"},
														Sel: &dst.Ident{Name: "Command2"},
													},
												},
											},
											&dst.UnaryExpr{
												Op: token.AND,
												X: &dst.CompositeLit{
													Type: &dst.SelectorExpr{
														X:   &dst.Ident{Name: "commands"},
														Sel: &dst.Ident{Name: "Command3"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		itemExpr := MustParseExpr("&commands.Command2{}").(dst.Expr)
		handler.removeFromExisting(withCall, itemExpr)

		funcLit := withCall.Args[0].(*dst.FuncLit)
		retStmt := funcLit.Body.List[0].(*dst.ReturnStmt)
		compositeLit := retStmt.Results[0].(*dst.CompositeLit)
		s.Len(compositeLit.Elts, 2)

		// Verify Command1 and Command3 remain
		unary1 := compositeLit.Elts[0].(*dst.UnaryExpr)
		comp1 := unary1.X.(*dst.CompositeLit)
		sel1 := comp1.Type.(*dst.SelectorExpr)
		s.Equal("Command1", sel1.Sel.Name)

		unary3 := compositeLit.Elts[1].(*dst.UnaryExpr)
		comp3 := unary3.X.(*dst.CompositeLit)
		sel3 := comp3.Type.(*dst.SelectorExpr)
		s.Equal("Command3", sel3.Sel.Name)
	})

	s.Run("EmptyArgs", func() {
		handler := newWithSliceHandler(config)

		// Test with empty args
		withCall := &dst.CallExpr{
			Args: []dst.Expr{},
		}

		itemExpr := &dst.Ident{Name: "toRemove"}
		handler.removeFromExisting(withCall, itemExpr)

		// Should not panic and should not modify anything
		s.Len(withCall.Args, 0)
	})

	s.Run("NotCompositeLit", func() {
		handler := newWithSliceHandler(config)

		// Test with non-function literal arg
		withCall := &dst.CallExpr{
			Args: []dst.Expr{
				&dst.Ident{Name: "notAFuncLit"},
			},
		}

		itemExpr := &dst.Ident{Name: "toRemove"}
		handler.removeFromExisting(withCall, itemExpr)

		// Should not panic and should not modify
		s.Len(withCall.Args, 1)
		s.Equal("notAFuncLit", withCall.Args[0].(*dst.Ident).Name)
	})

	s.Run("ItemNotFound", func() {
		handler := newWithSliceHandler(config)

		// Test removing an item that doesn't exist
		withCall := &dst.CallExpr{
			Args: []dst.Expr{
				&dst.FuncLit{
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ReturnStmt{
								Results: []dst.Expr{
									&dst.CompositeLit{
										Elts: []dst.Expr{
											&dst.UnaryExpr{
												Op: token.AND,
												X: &dst.CompositeLit{
													Type: &dst.SelectorExpr{
														X:   &dst.Ident{Name: "commands"},
														Sel: &dst.Ident{Name: "Command1"},
													},
												},
											},
											&dst.UnaryExpr{
												Op: token.AND,
												X: &dst.CompositeLit{
													Type: &dst.SelectorExpr{
														X:   &dst.Ident{Name: "commands"},
														Sel: &dst.Ident{Name: "Command2"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		itemExpr := MustParseExpr("&commands.NonExistent{}").(dst.Expr)
		handler.removeFromExisting(withCall, itemExpr)

		funcLit := withCall.Args[0].(*dst.FuncLit)
		retStmt := funcLit.Body.List[0].(*dst.ReturnStmt)
		compositeLit := retStmt.Results[0].(*dst.CompositeLit)
		// Should still have both items since the item to remove wasn't found
		s.Len(compositeLit.Elts, 2)
	})
}

func Test_appendToExistingMiddleware(t *testing.T) {
	tests := []struct {
		name              string
		initialContent    string
		middlewareToAdd   string
		expectedArgsCount int
	}{
		{
			name: "append to existing Append call",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Auth{})
		}).Start()
}`,
			middlewareToAdd:   "&middleware.Cors{}",
			expectedArgsCount: 2,
		},
		{
			name: "append to empty function",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
		}).Start()
}`,
			middlewareToAdd:   "&middleware.Auth{}",
			expectedArgsCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := filepath.Join(t.TempDir(), "test.go")
			require.NoError(t, supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			require.NoError(t, err)

			file, err := decorator.Parse(content)
			require.NoError(t, err)

			// Find the WithMiddleware call
			var withMiddlewareCall *dst.CallExpr
			dst.Inspect(file, func(n dst.Node) bool {
				if call, ok := n.(*dst.CallExpr); ok {
					if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
						if sel.Sel.Name == "WithMiddleware" {
							withMiddlewareCall = call
							return false
						}
					}
				}
				return true
			})

			assert.NotNil(t, withMiddlewareCall, "Expected to find WithMiddleware call")

			middlewareExpr := MustParseExpr(tt.middlewareToAdd).(dst.Expr)
			appendToExistingMiddleware(withMiddlewareCall, middlewareExpr)

			funcLit := withMiddlewareCall.Args[0].(*dst.FuncLit)
			appendCall := findMiddlewareAppendCall(funcLit)

			assert.NotNil(t, appendCall, "Expected Append call to exist after modification")
			assert.Equal(t, tt.expectedArgsCount, len(appendCall.Args), "Expected %d arguments in Append call", tt.expectedArgsCount)
		})
	}
}

func Test_addMiddlewareAppendCall(t *testing.T) {
	tests := []struct {
		name            string
		initialContent  string
		middlewareToAdd string
	}{
		{
			name: "add Append to empty function",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
		}).Start()
}`,
			middlewareToAdd: "&middleware.Auth{}",
		},
		{
			name: "add Append to function with other statements",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithMiddleware(func(handler configuration.Middleware) {
		handler.Register(&middleware.Other{})
	}).Start()
}`,
			middlewareToAdd: "&middleware.Cors{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := filepath.Join(t.TempDir(), "test.go")
			require.NoError(t, supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			require.NoError(t, err)

			file, err := decorator.Parse(content)
			require.NoError(t, err)

			// Find the function literal
			var funcLit *dst.FuncLit
			dst.Inspect(file, func(n dst.Node) bool {
				if fl, ok := n.(*dst.FuncLit); ok {
					funcLit = fl
					return false
				}
				return true
			})

			assert.NotNil(t, funcLit, "Expected to find function literal")

			originalStmtCount := len(funcLit.Body.List)
			middlewareExpr := MustParseExpr(tt.middlewareToAdd).(dst.Expr)

			addMiddlewareAppendCall(funcLit, middlewareExpr)

			assert.Equal(t, originalStmtCount+1, len(funcLit.Body.List), "Expected one more statement")

			appendCall := findMiddlewareAppendCall(funcLit)
			assert.NotNil(t, appendCall, "Expected to find newly added Append call")
			assert.Equal(t, 1, len(appendCall.Args), "Expected exactly 1 argument in Append call")
		})
	}
}

func Test_addMiddlewareImports(t *testing.T) {
	tests := []struct {
		name             string
		initialContent   string
		pkg              string
		expectError      bool
		expectedImports  []string
		unexpectedImport string
	}{
		{
			name: "add middleware imports to file with existing imports",
			initialContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:         "github.com/goravel/framework/http/middleware",
			expectError: false,
			expectedImports: []string{
				"github.com/goravel/framework/http/middleware",
				"github.com/goravel/framework/contracts/foundation/configuration",
			},
		},
		{
			name: "add middleware imports when configuration import already exists",
			initialContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:         "github.com/goravel/framework/http/middleware",
			expectError: false,
			expectedImports: []string{
				"github.com/goravel/framework/http/middleware",
				"github.com/goravel/framework/contracts/foundation/configuration",
			},
		},
		{
			name: "add middleware imports when middleware package already exists",
			initialContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:         "github.com/goravel/framework/http/middleware",
			expectError: false,
			expectedImports: []string{
				"github.com/goravel/framework/http/middleware",
				"github.com/goravel/framework/contracts/foundation/configuration",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := filepath.Join(t.TempDir(), "test.go")
			require.NoError(t, supportfile.PutContent(sourceFile, tt.initialContent))

			err := addMiddlewareImports(sourceFile, tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			content, err := supportfile.GetContent(sourceFile)
			require.NoError(t, err)

			for _, expectedImport := range tt.expectedImports {
				assert.Contains(t, content, expectedImport, "Expected import %s to be present", expectedImport)
			}

			if tt.unexpectedImport != "" {
				assert.NotContains(t, content, tt.unexpectedImport)
			}
		})
	}
}

func Test_createWithMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		initialContent  string
		middlewareToAdd string
		expectedContent string
	}{
		{
			name: "create WithMiddleware and insert into chain",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			middlewareToAdd: "&middleware.Auth{}",
			expectedContent: `WithMiddleware(func(handler configuration.Middleware) {
		handler.Append(
			&middleware.Auth{},
		)
	}).WithConfig(config.Boot)`,
		},
		{
			name: "create WithMiddleware when multiple chain calls exist",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).Start()
}
`,
			middlewareToAdd: "&middleware.Cors{}",
			expectedContent: `WithMiddleware(func(handler configuration.Middleware) {
		handler.Append(
			&middleware.Cors{},
		)
	}).WithConfig(config.Boot)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := filepath.Join(t.TempDir(), "test.go")
			require.NoError(t, supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			require.NoError(t, err)

			file, err := decorator.Parse(content)
			require.NoError(t, err)

			// Find the foundation.Setup() call and the chain
			var setupCall *dst.CallExpr
			var parentOfSetup *dst.SelectorExpr
			dst.Inspect(file, func(n dst.Node) bool {
				if call, ok := n.(*dst.CallExpr); ok {
					if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
						if innerCall, ok := sel.X.(*dst.CallExpr); ok {
							if innerSel, ok := innerCall.Fun.(*dst.SelectorExpr); ok {
								if innerSel.Sel.Name == "Setup" {
									setupCall = innerCall
									parentOfSetup = sel
									return false
								}
							}
						}
					}
				}
				return true
			})

			assert.NotNil(t, setupCall, "Expected to find Setup call")
			assert.NotNil(t, parentOfSetup, "Expected to find parent of Setup")

			middlewareExpr := MustParseExpr(tt.middlewareToAdd).(dst.Expr)
			createWithMiddleware(setupCall, parentOfSetup, middlewareExpr)

			// Verify the structure was created
			assert.NotNil(t, parentOfSetup.X, "Expected parentOfSetup.X to be updated")
			withMiddlewareCall, ok := parentOfSetup.X.(*dst.CallExpr)
			assert.True(t, ok, "Expected parentOfSetup.X to be a CallExpr")

			sel, ok := withMiddlewareCall.Fun.(*dst.SelectorExpr)
			assert.True(t, ok, "Expected WithMiddleware fun to be a SelectorExpr")
			assert.Equal(t, "WithMiddleware", sel.Sel.Name)

			require.Len(t, withMiddlewareCall.Args, 1)
			funcLit, ok := withMiddlewareCall.Args[0].(*dst.FuncLit)
			assert.True(t, ok, "Expected first argument to be a function literal")

			appendCall := findMiddlewareAppendCall(funcLit)
			assert.NotNil(t, appendCall, "Expected Append call to exist")
			assert.Equal(t, 1, len(appendCall.Args), "Expected exactly 1 argument in Append call")
		})
	}
}

func Test_containsFoundationSetup(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		expected bool
	}{
		{
			name:     "contains foundation.Setup()",
			stmt:     `foundation.Setup().Run()`,
			expected: true,
		},
		{
			name:     "contains foundation.Setup() in chain",
			stmt:     `foundation.Setup().WithConfig(config.Boot).Run()`,
			expected: true,
		},
		{
			name:     "does not contain foundation.Setup()",
			stmt:     `app.Run()`,
			expected: false,
		},
		{
			name:     "contains Setup() but not from foundation",
			stmt:     `something.Setup().Run()`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := MustParseExpr(tt.stmt).(dst.Expr)
			stmt := &dst.ExprStmt{X: expr}

			result := containsFoundationSetup(stmt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_findFoundationSetupCallsForMiddleware(t *testing.T) {
	tests := []struct {
		name                    string
		initialContent          string
		expectSetup             bool
		expectWithMiddleware    bool
		expectParentOfSetup     bool
		withMiddlewareArgsCount int
	}{
		{
			name: "find Setup without WithMiddleware",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			expectSetup:          true,
			expectWithMiddleware: false,
			expectParentOfSetup:  true,
		},
		{
			name: "find Setup with WithMiddleware",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Auth{})
		}).WithConfig(config.Boot).Start()
}
`,
			expectSetup:             true,
			expectWithMiddleware:    true,
			expectParentOfSetup:     true,
			withMiddlewareArgsCount: 1,
		},
		{
			name: "find Setup with complex chain",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).WithSchedule(schedule.Boot).Start()
}
`,
			expectSetup:          true,
			expectWithMiddleware: false,
			expectParentOfSetup:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := filepath.Join(t.TempDir(), "test.go")
			require.NoError(t, supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			require.NoError(t, err)

			file, err := decorator.Parse(content)
			require.NoError(t, err)

			// Find the main call expression
			var mainCallExpr *dst.CallExpr
			dst.Inspect(file, func(n dst.Node) bool {
				if retStmt, ok := n.(*dst.ReturnStmt); ok {
					if len(retStmt.Results) > 0 {
						if call, ok := retStmt.Results[0].(*dst.CallExpr); ok {
							mainCallExpr = call
							return false
						}
					}
				}
				return true
			})

			assert.NotNil(t, mainCallExpr, "Expected to find main call expression")

			setupCall, withMiddlewareCall, parentOfSetup := findFoundationSetupCallsForMiddleware(mainCallExpr)

			if tt.expectSetup {
				assert.NotNil(t, setupCall, "Expected to find Setup call")
				sel, ok := setupCall.Fun.(*dst.SelectorExpr)
				assert.True(t, ok)
				assert.Equal(t, "Setup", sel.Sel.Name)
			} else {
				assert.Nil(t, setupCall, "Expected not to find Setup call")
			}

			if tt.expectWithMiddleware {
				assert.NotNil(t, withMiddlewareCall, "Expected to find WithMiddleware call")
				sel, ok := withMiddlewareCall.Fun.(*dst.SelectorExpr)
				assert.True(t, ok)
				assert.Equal(t, "WithMiddleware", sel.Sel.Name)
				assert.Equal(t, tt.withMiddlewareArgsCount, len(withMiddlewareCall.Args))
			} else {
				assert.Nil(t, withMiddlewareCall, "Expected not to find WithMiddleware call")
			}

			if tt.expectParentOfSetup {
				assert.NotNil(t, parentOfSetup, "Expected to find parent of Setup")
			} else {
				assert.Nil(t, parentOfSetup, "Expected not to find parent of Setup")
			}
		})
	}
}

func Test_findMiddlewareAppendCall(t *testing.T) {
	tests := []struct {
		name           string
		initialContent string
		expectFound    bool
		expectedArgs   int
	}{
		{
			name: "find Append call with single argument",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithMiddleware(func(handler configuration.Middleware) {
		handler.Append(&middleware.Auth{})
	}).Start()
}`,
			expectFound:  true,
			expectedArgs: 1,
		},
		{
			name: "find Append call with multiple arguments",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Auth{}, &middleware.Cors{})
		}).Start()
}`,
			expectFound:  true,
			expectedArgs: 2,
		},
		{
			name: "return nil when no Append call exists",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
		}).Start()
}`,
			expectFound: false,
		},
		{
			name: "return nil when function has other calls but not Append",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Register(&middleware.Auth{})
		}).Start()
}`,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := filepath.Join(t.TempDir(), "test.go")
			require.NoError(t, supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			require.NoError(t, err)

			file, err := decorator.Parse(content)
			require.NoError(t, err)

			// Find the function literal
			var funcLit *dst.FuncLit
			dst.Inspect(file, func(n dst.Node) bool {
				if fl, ok := n.(*dst.FuncLit); ok {
					funcLit = fl
					return false
				}
				return true
			})

			assert.NotNil(t, funcLit, "Expected to find function literal")

			appendCall := findMiddlewareAppendCall(funcLit)

			if tt.expectFound {
				assert.NotNil(t, appendCall, "Expected to find Append call")
				assert.Equal(t, tt.expectedArgs, len(appendCall.Args), "Expected %d arguments in Append call", tt.expectedArgs)
			} else {
				assert.Nil(t, appendCall, "Expected not to find Append call")
			}
		})
	}
}

func Test_foundationSetupMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		initialContent  string
		middlewareToAdd string
		expectedResult  string
	}{
		{
			name: "modify chain without WithMiddleware",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			middlewareToAdd: "&middleware.Auth{}",
			expectedResult: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&middleware.Auth{},
			)
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "modify chain with existing WithMiddleware",
			initialContent: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Existing{})
		}).WithConfig(config.Boot).Start()
}
`,
			middlewareToAdd: "&middleware.Auth{}",
			expectedResult: `package test

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Existing{},
				&middleware.Auth{},
			)
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "skip non-foundation.Setup() statements",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	app := foundation.NewApplication()
	app.Run()
}
`,
			middlewareToAdd: "&middleware.Auth{}",
			expectedResult: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	app := foundation.NewApplication()
	app.Run()
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := filepath.Join(t.TempDir(), "test.go")
			require.NoError(t, supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			require.NoError(t, err)

			_, err = decorator.Parse(content)
			require.NoError(t, err)

			// Apply the action
			err = GoFile(sourceFile).Find(match.FoundationSetup()).Modify(foundationSetupMiddleware(tt.middlewareToAdd)).Apply()
			assert.NoError(t, err)

			// Read the result
			resultContent, err := supportfile.GetContent(sourceFile)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedResult, resultContent)
		})
	}
}

func (s *WithSliceHandlerTestSuite) TestCreateWithMethod_TypedSlice() {
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

	itemExpr := &dst.Ident{Name: "&commands.Command{}"}

	handler.createWithMethod(setupCall, parentOfSetup, itemExpr)

	// Verify the chain was modified
	newWithCall, ok := parentOfSetup.X.(*dst.CallExpr)
	s.True(ok)
	s.NotNil(newWithCall)

	withSel, ok := newWithCall.Fun.(*dst.SelectorExpr)
	s.True(ok)
	s.Equal("WithCommands", withSel.Sel.Name)

	// Verify the function literal wrapping
	s.Len(newWithCall.Args, 1)
	funcLit, ok := newWithCall.Args[0].(*dst.FuncLit)
	s.True(ok)
	s.NotNil(funcLit)

	// Verify the return statement contains the composite literal with []console.Command type
	s.Len(funcLit.Body.List, 1)
	retStmt, ok := funcLit.Body.List[0].(*dst.ReturnStmt)
	s.True(ok)
	s.Len(retStmt.Results, 1)
	compositeLit, ok := retStmt.Results[0].(*dst.CompositeLit)
	s.True(ok)

	arrayType, ok := compositeLit.Type.(*dst.ArrayType)
	s.True(ok)

	selectorExpr, ok := arrayType.Elt.(*dst.SelectorExpr)
	s.True(ok)
	s.Equal("console", selectorExpr.X.(*dst.Ident).Name)
	s.Equal("Command", selectorExpr.Sel.Name)

	s.Len(compositeLit.Elts, 1)
}
