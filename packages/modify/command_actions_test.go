package modify

import (
	"path/filepath"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support"
	supportfile "github.com/goravel/framework/support/file"
)

type CommandActionsTestSuite struct {
	suite.Suite
}

func TestCommandActionsTestSuite(t *testing.T) {
	suite.Run(t, new(CommandActionsTestSuite))
}

func (s *CommandActionsTestSuite) TestAddCommand() {
	tests := []struct {
		name              string
		appContent        string
		commandsContent   string // empty if file doesn't exist
		pkg               string
		command           string
		expectedApp       string
		expectedCommands  string // empty if file shouldn't be created
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add command when WithCommands doesn't exist and commands.go doesn't exist",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.ExampleCommand{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands(Commands()).WithConfig(config.Boot).Run()
}
`,
			expectedCommands: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExampleCommand{},
	}
}
`,
		},
		{
			name: "add command when WithCommands exists with Commands() and commands.go exists",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/bootstrap"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands(Commands()).WithConfig(config.Boot).Run()
}
`,
			commandsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
	}
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.NewCommand{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/bootstrap"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands(Commands()).WithConfig(config.Boot).Run()
}
`,
			expectedCommands: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
		&commands.NewCommand{},
	}
}
`,
		},
		{
			name: "add command when WithCommands exists with inline array",
			appContent: `package bootstrap

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
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.NewCommand{}",
			expectedApp: `package bootstrap

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
			&commands.NewCommand{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "error when commands.go exists but WithCommands doesn't exist",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			commandsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
	}
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.NewCommand{}",
			wantErr: true,
		},
		{
			name: "add command when WithCommands doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.FirstCommand{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
		WithCommands(Commands()).Run()
}
`,
			expectedCommands: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.FirstCommand{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempDir := s.T().TempDir()
			bootstrapDir := filepath.Join(tempDir, "bootstrap")

			appFile := filepath.Join(bootstrapDir, "app.go")
			commandsFile := filepath.Join(bootstrapDir, "commands.go")

			s.Require().NoError(supportfile.PutContent(appFile, tt.appContent))

			if tt.commandsContent != "" {
				s.Require().NoError(supportfile.PutContent(commandsFile, tt.commandsContent))
			}

			// Override Config.Paths.App for testing
			originalAppPath := support.Config.Paths.App
			support.Config.Paths.App = appFile
			defer func() {
				support.Config.Paths.App = originalAppPath
			}()

			err := AddCommand(tt.pkg, tt.command)

			if tt.wantErr {
				s.Require().Error(err)
				if tt.expectedErrString != "" {
					s.Contains(err.Error(), tt.expectedErrString)
				}
				return
			}

			s.Require().NoError(err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			s.Require().NoError(err)
			s.Equal(tt.expectedApp, appContent)

			// Verify commands.go content if expected
			if tt.expectedCommands != "" {
				commandsContent, err := supportfile.GetContent(commandsFile)
				s.Require().NoError(err)
				s.Equal(tt.expectedCommands, commandsContent)
			}
		})
	}
}

func (s *CommandActionsTestSuite) Test_checkWithCommandsExists() {
	tests := []struct {
		name     string
		content  string
		expected bool
		wantErr  bool
	}{
		{
			name: "WithCommands exists in chain",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/bootstrap"
)

func Boot() {
	foundation.Setup().WithCommands(Commands()).Run()
}
`,
			expected: true,
		},
		{
			name: "WithCommands exists with inline array",
			content: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().WithCommands([]console.Command{}).Run()
}
`,
			expected: true,
		},
		{
			name: "WithCommands doesn't exist",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`,
			expected: false,
		},
		{
			name: "WithCommands doesn't exist in complex chain",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).Run()
}
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "app.go")
			s.Require().NoError(supportfile.PutContent(tempFile, tt.content))

			result, err := checkWithCommandsExists(tempFile)

			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *CommandActionsTestSuite) Test_createCommandsFile() {
	tests := []struct {
		name            string
		expectedContent string
	}{
		{
			name: "create commands.go file with correct structure",
			expectedContent: `package bootstrap

import "github.com/goravel/framework/contracts/console"

func Commands() []console.Command {
	return []console.Command{}
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "commands.go")

			err := createCommandsFile(tempFile)
			s.NoError(err)

			content, err := supportfile.GetContent(tempFile)
			s.Require().NoError(err)
			s.Equal(tt.expectedContent, content)
		})
	}
}

func (s *CommandActionsTestSuite) Test_addCommandToCommandsFile() {
	tests := []struct {
		name            string
		initialContent  string
		pkg             string
		command         string
		expectedContent string
	}{
		{
			name: "add command to empty Commands() function",
			initialContent: `package bootstrap

import "github.com/goravel/framework/contracts/console"

func Commands() []console.Command {
	return []console.Command{}
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.ExampleCommand{}",
			expectedContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExampleCommand{},
	}
}
`,
		},
		{
			name: "add command to existing Commands() function",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
	}
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.NewCommand{}",
			expectedContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
		&commands.NewCommand{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "commands.go")
			s.Require().NoError(supportfile.PutContent(tempFile, tt.initialContent))

			err := addCommandToCommandsFile(tempFile, tt.pkg, tt.command)
			s.NoError(err)

			content, err := supportfile.GetContent(tempFile)
			s.Require().NoError(err)
			s.Equal(tt.expectedContent, content)
		})
	}
}

func (s *CommandActionsTestSuite) Test_addCommandImports() {
	tests := []struct {
		name            string
		initialContent  string
		pkg             string
		expectedImports []string
	}{
		{
			name: "add command imports to file with existing imports",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg: "goravel/app/console/commands",
			expectedImports: []string{
				"goravel/app/console/commands",
				"github.com/goravel/framework/contracts/console",
			},
		},
		{
			name: "add command imports when console import already exists",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg: "goravel/app/console/commands",
			expectedImports: []string{
				"goravel/app/console/commands",
				"github.com/goravel/framework/contracts/console",
			},
		},
		{
			name: "add command imports when command package already exists",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg: "goravel/app/console/commands",
			expectedImports: []string{
				"goravel/app/console/commands",
				"github.com/goravel/framework/contracts/console",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "app.go")
			s.Require().NoError(supportfile.PutContent(tempFile, tt.initialContent))

			err := addCommandImports(tempFile, tt.pkg)
			s.NoError(err)

			content, err := supportfile.GetContent(tempFile)
			s.Require().NoError(err)

			for _, expectedImport := range tt.expectedImports {
				s.Contains(content, expectedImport, "Expected import %s to be present", expectedImport)
			}
		})
	}
}

func (s *CommandActionsTestSuite) Test_foundationSetupCommandInline() {
	tests := []struct {
		name            string
		initialContent  string
		commandToAdd    string
		expectedContent string
	}{
		{
			name: "create WithCommands with inline array when it doesn't exist",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			commandToAdd: "&commands.ExampleCommand{}",
			expectedContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands([]console.Command{
			&commands.ExampleCommand{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "append to existing WithCommands inline array",
			initialContent: `package test

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
`,
			commandToAdd: "&commands.NewCommand{}",
			expectedContent: `package test

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
			&commands.NewCommand{},
		}).WithConfig(config.Boot).Run()
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
			commandToAdd: "&commands.ExampleCommand{}",
			expectedContent: `package test

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
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(tempFile, tt.initialContent))

			err := GoFile(tempFile).Find(match.FoundationSetup()).Modify(foundationSetupCommandInline(tt.commandToAdd)).Apply()
			s.NoError(err)

			content, err := supportfile.GetContent(tempFile)
			s.Require().NoError(err)
			s.Equal(tt.expectedContent, content)
		})
	}
}

func (s *CommandActionsTestSuite) Test_foundationSetupCommandWithFunction() {
	tests := []struct {
		name            string
		initialContent  string
		expectedContent string
	}{
		{
			name: "add WithCommands(Commands()) to Setup chain",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			expectedContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands(Commands()).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "add WithCommands(Commands()) at the beginning of chain",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`,
			expectedContent: `package test

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
		WithCommands(Commands()).Run()
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(tempFile, tt.initialContent))

			err := GoFile(tempFile).Find(match.FoundationSetup()).Modify(foundationSetupCommandWithFunction()).Apply()
			s.NoError(err)

			content, err := supportfile.GetContent(tempFile)
			s.Require().NoError(err)
			s.Equal(tt.expectedContent, content)
		})
	}
}

func (s *CommandActionsTestSuite) Test_appendToExistingCommand() {
	tests := []struct {
		name              string
		initialContent    string
		commandToAdd      string
		expectedArgsCount int
	}{
		{
			name: "append to existing WithCommands inline array",
			initialContent: `package test

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
}`,
			commandToAdd:      "&commands.NewCommand{}",
			expectedArgsCount: 2,
		},
		{
			name: "append to empty inline array",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
		WithCommands([]console.Command{}).Run()
}`,
			commandToAdd:      "&commands.FirstCommand{}",
			expectedArgsCount: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(tempFile, tt.initialContent))

			content, err := supportfile.GetContent(tempFile)
			s.Require().NoError(err)

			file, err := decorator.Parse(content)
			s.Require().NoError(err)

			// Find the WithCommands call
			var withCommandCall *dst.CallExpr
			dst.Inspect(file, func(n dst.Node) bool {
				if call, ok := n.(*dst.CallExpr); ok {
					if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
						if sel.Sel.Name == "WithCommands" {
							withCommandCall = call
							return false
						}
					}
				}
				return true
			})

			s.Require().NotNil(withCommandCall, "WithCommands call not found")

			commandExpr := MustParseExpr(tt.commandToAdd).(dst.Expr)
			appendToExistingCommand(withCommandCall, commandExpr)

			// Verify the command was appended
			s.Require().Len(withCommandCall.Args, 1)
			compositeLit, ok := withCommandCall.Args[0].(*dst.CompositeLit)
			s.Require().True(ok)
			s.Equal(tt.expectedArgsCount, len(compositeLit.Elts))
		})
	}
}

func (s *CommandActionsTestSuite) Test_createWithCommands() {
	tests := []struct {
		name           string
		initialContent string
		commandToAdd   string
	}{
		{
			name: "create WithCommands and insert into chain",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			commandToAdd: "&commands.ExampleCommand{}",
		},
		{
			name: "create WithCommands when multiple chain calls exist",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).Run()
}
`,
			commandToAdd: "&commands.TestCommand{}",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(tempFile, tt.initialContent))

			content, err := supportfile.GetContent(tempFile)
			s.Require().NoError(err)

			file, err := decorator.Parse(content)
			s.Require().NoError(err)

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

			s.NotNil(setupCall, "Expected to find Setup call")
			s.NotNil(parentOfSetup, "Expected to find parent of Setup")

			commandExpr := MustParseExpr(tt.commandToAdd).(dst.Expr)
			createWithCommands(setupCall, parentOfSetup, commandExpr)

			// Verify the structure was created
			s.NotNil(parentOfSetup.X, "Expected parentOfSetup.X to be updated")
			withCommandCall, ok := parentOfSetup.X.(*dst.CallExpr)
			s.True(ok, "Expected parentOfSetup.X to be a CallExpr")

			sel, ok := withCommandCall.Fun.(*dst.SelectorExpr)
			s.True(ok, "Expected WithCommands fun to be a SelectorExpr")
			s.Equal("WithCommands", sel.Sel.Name)

			s.Require().Len(withCommandCall.Args, 1)
			compositeLit, ok := withCommandCall.Args[0].(*dst.CompositeLit)
			s.True(ok, "Expected first argument to be a composite literal")
			s.Equal(1, len(compositeLit.Elts), "Expected exactly 1 element in composite literal")
		})
	}
}

func (s *CommandActionsTestSuite) Test_findFoundationSetupCallsForCommand() {
	tests := []struct {
		name                 string
		initialContent       string
		expectSetup          bool
		expectWithCommand    bool
		expectParentOfSetup  bool
		withCommandArgsCount int
	}{
		{
			name: "find Setup without WithCommands",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			expectSetup:         true,
			expectWithCommand:   false,
			expectParentOfSetup: true,
		},
		{
			name: "find Setup with WithCommands",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands([]console.Command{
			&commands.ExampleCommand{},
		}).WithConfig(config.Boot).Run()
}
`,
			expectSetup:          true,
			expectWithCommand:    true,
			expectParentOfSetup:  true,
			withCommandArgsCount: 1,
		},
		{
			name: "find Setup with complex chain",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).WithMiddleware(mw.Boot).Run()
}
`,
			expectSetup:         true,
			expectWithCommand:   false,
			expectParentOfSetup: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(tempFile, tt.initialContent))

			content, err := supportfile.GetContent(tempFile)
			s.Require().NoError(err)

			file, err := decorator.Parse(content)
			s.Require().NoError(err)

			// Find the outermost call expression (Run())
			var callExpr *dst.CallExpr
			dst.Inspect(file, func(n dst.Node) bool {
				if call, ok := n.(*dst.CallExpr); ok {
					if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
						if sel.Sel.Name == "Run" {
							callExpr = call
							return false
						}
					}
				}
				return true
			})

			s.Require().NotNil(callExpr, "Run() call not found")

			setupCall, withCommandCall, parentOfSetup := findFoundationSetupCallsForCommand(callExpr)

			if tt.expectSetup {
				s.NotNil(setupCall, "Expected to find Setup call")
			} else {
				s.Nil(setupCall, "Did not expect to find Setup call")
			}

			if tt.expectWithCommand {
				s.NotNil(withCommandCall, "Expected to find WithCommands call")
				if tt.withCommandArgsCount > 0 {
					s.Len(withCommandCall.Args, tt.withCommandArgsCount)
				}
			} else {
				s.Nil(withCommandCall, "Did not expect to find WithCommands call")
			}

			if tt.expectParentOfSetup {
				s.NotNil(parentOfSetup, "Expected to find parent of Setup")
			} else {
				s.Nil(parentOfSetup, "Did not expect to find parent of Setup")
			}
		})
	}
}
