package console

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	contractsmodular "github.com/goravel/framework/contracts/modular"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type MakeTestSuite struct {
	suite.Suite
	make *Make
}

func TestMakeTestSuite(t *testing.T) {
	suite.Run(t, new(MakeTestSuite))
}

func (s *MakeTestSuite) SetupTest() {
	s.make = &Make{
		name: "Lowercase",
		root: filepath.Join("app", "rules"),
	}
}

func (s *MakeTestSuite) TestGetFilePath() {
	pwd, _ := os.Getwd()
	s.Equal(filepath.Join(pwd, s.make.root, "lowercase.go"), s.make.GetFilePath())

	s.make.name = "user/Lowercase"
	s.Equal(filepath.Join(pwd, s.make.root, "user", "lowercase.go"), s.make.GetFilePath())
}

func (s *MakeTestSuite) TestGetSignature() {
	s.Equal("Lowercase", s.make.GetSignature())

	s.make.name = "user/Lowercase"
	s.Equal("UserLowercase", s.make.GetSignature())

	s.make.name = "user/Lowercase/Uppercase"
	s.Equal("UserLowercaseUppercase", s.make.GetSignature())
}

func (s *MakeTestSuite) TestGetStructName() {
	s.Equal("Lowercase", s.make.GetStructName())

	s.make.name = "lowercase"
	s.Equal("Lowercase", s.make.GetStructName())

	s.make.name = "user/Lowercase"
	s.Equal("Lowercase", s.make.GetStructName())
}

func (s *MakeTestSuite) TestGetPackageImportPath() {
	s.Contains(s.make.GetPackageImportPath(), "/app/rules")

	s.make.name = "user/Lowercase"
	s.Contains(s.make.GetPackageImportPath(), "/app/rules/user")

	s.make.name = "user/Lowercase/Uppercase"
	s.Contains(s.make.GetPackageImportPath(), "/app/rules/user/Lowercase")
}

func (s *MakeTestSuite) TestGetPackageName() {
	s.Equal("rules", s.make.GetPackageName())

	s.make.name = "user/Lowercase"
	s.Equal("user", s.make.GetPackageName())

	// Test with forward slashes in root (cross-platform compatibility)
	s.make.name = "Auth"
	s.make.root = "app/http/middleware"
	s.Equal("middleware", s.make.GetPackageName())

	// Test with nested paths using forward slashes
	s.make.name = "user/Custom"
	s.make.root = "app/http/middleware"
	s.Equal("user", s.make.GetPackageName())

	// Test with Windows-style backslashes in root
	s.make.name = "Auth"
	s.make.root = `app\http\middleware`
	s.Equal("middleware", s.make.GetPackageName())

	// Test with mixed separators (Windows path with forward slash in name)
	s.make.name = "user/Custom"
	s.make.root = `app\http\middleware`
	s.Equal("user", s.make.GetPackageName())

	// Test with Windows-style path at different depth
	s.make.name = "Verify"
	s.make.root = `app\http`
	s.Equal("http", s.make.GetPackageName())
}

func (s *MakeTestSuite) TestGetFolderPath() {
	s.Empty(s.make.GetFolderPath())

	s.make.name = "user/Lowercase"
	s.Equal("user", s.make.GetFolderPath())
}

func TestNewMake(t *testing.T) {
	var (
		name string

		mockCtx = mocksconsole.NewContext(t)
		ttype   = "rule"
		root    = filepath.Join("app", "rules")
	)

	tests := []struct {
		name        string
		setup       func()
		expectMake  *Make
		expectError error
	}{
		{
			name: "Sad path - name is empty",
			setup: func() {
				name = ""
				mockCtx.EXPECT().Ask("Enter the rule name", mock.Anything).Return("", errors.ConsoleEmptyFieldValue.Args("rule")).Once()
			},
			expectMake:  nil,
			expectError: errors.ConsoleEmptyFieldValue.Args("rule"),
		},
		{
			name: "Sad path - name already exists",
			setup: func() {
				name = "Uppercase"
				assert.Nil(t, file.PutContent(filepath.Join(root, "uppercase.go"), ""))
				mockCtx.EXPECT().Option("module").Return("").Once()
				mockCtx.EXPECT().OptionBool("force").Return(false).Once()
			},
			expectMake:  nil,
			expectError: errors.ConsoleFileAlreadyExists.Args("rule"),
		},
		{
			name: "Happy path - name already exists, but force is true",
			setup: func() {
				name = "Uppercase"
				assert.Nil(t, file.PutContent(filepath.Join(root, "uppercase.go"), ""))
				mockCtx.EXPECT().Option("module").Return("").Once()
				mockCtx.EXPECT().OptionBool("force").Return(true).Once()
			},
			expectMake:  &Make{name: "Lowercase", root: root},
			expectError: nil,
		},
		{
			name: "Happy path - name is not empty",
			setup: func() {
				name = "Lowercase"
				mockCtx.EXPECT().Option("module").Return("").Once()
				mockCtx.On("OptionBool", "force").Return(false).Once()
			},
			expectMake:  &Make{name: "Lowercase", root: root},
			expectError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			m, err := NewMake(mockCtx, ttype, name, root)
			if test.expectError != nil {
				assert.Equal(t, test.expectError, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, m)
				assert.Nil(t, file.Remove("app"))
			}
		})
	}
}

// testModule sets its own directory, so it is a module of the application.
type testModule struct {
	name string
	path string
}

func (r *testModule) Name() string { return r.name }
func (r *testModule) Path() string { return r.path }

func TestNewMakeInModule(t *testing.T) {
	registry := modular.NewRegistry("")
	// *flag.FlagSet has Name() and lives outside the application: a third-party module.
	assert.NoError(t, registry.Register(
		&testModule{name: "blog-posts", path: "modules/blog-posts"},
		flag.NewFlagSet("shop", flag.ContinueOnError),
	))

	tests := []struct {
		name        string
		module      string
		registry    contractsmodular.Registry
		root        string
		expectRoot  string
		expectError error
	}{
		{name: "app directory is dropped", module: "blog-posts", registry: registry, root: "app/http/controllers", expectRoot: "modules/blog-posts/http/controllers"},
		{name: "name is converted to kebab case", module: "BlogPosts", registry: registry, root: "app/models", expectRoot: "modules/blog-posts/models"},
		{name: "other directories keep their shape", module: "blog-posts", registry: registry, root: "database/migrations", expectRoot: "modules/blog-posts/database/migrations"},
		{name: "unknown module", module: "news", registry: registry, root: "app/models", expectError: errors.ModularModuleNotFound.Args("news")},
		{name: "third-party module", module: "shop", registry: registry, root: "app/models", expectError: errors.ModularModuleNotLocal.Args("shop")},
		{name: "no registry", module: "blog-posts", root: "app/models", expectError: errors.ModularRegistryNotSet},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modular.SetDefault(test.registry)
			defer modular.SetDefault(nil)

			mockCtx := mocksconsole.NewContext(t)
			mockCtx.EXPECT().Option("module").Return(test.module).Once()
			if test.expectError == nil {
				mockCtx.EXPECT().OptionBool("force").Return(false).Once()
			}

			m, err := NewMake(mockCtx, "model", "admin/User", test.root)
			if test.expectError != nil {
				assert.Equal(t, test.expectError, err)
				assert.Nil(t, m)

				return
			}

			pwd, _ := os.Getwd()
			assert.NoError(t, err)
			assert.Equal(t, "modules/blog-posts", m.GetModuleDir())
			assert.Equal(t, filepath.Join(pwd, filepath.FromSlash(test.expectRoot), "admin", "user.go"), m.GetFilePath())
			assert.Equal(t, env.MainPath()+"/"+test.expectRoot+"/admin", m.GetPackageImportPath())
			assert.Equal(t, "admin", m.GetPackageName())
		})
	}
}

func TestModuleRoot(t *testing.T) {
	assert.Equal(t, "modules/blog/http/controllers", ModuleRoot("modules/blog", "app/http/controllers"))
	assert.Equal(t, "modules/blog", ModuleRoot("modules/blog", "app"))
	assert.Equal(t, "modules/blog/application/models", ModuleRoot("modules/blog", "application/models"))
	assert.Equal(t, "modules/blog/tests", ModuleRoot("modules/blog", filepath.Join("tests")))
}

func TestHandleEach(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
		failOn    string
		wantNames []string
	}{
		{name: "runs makeOne for every name", arguments: []string{"User", "Post"}, wantNames: []string{"User", "Post"}},
		{name: "continues after a failure", arguments: []string{"User", "Post", "Comment"}, failOn: "Post", wantNames: []string{"User", "Post", "Comment"}},
		{name: "returns a single empty name when no arguments are given", wantNames: []string{""}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockCtx := mocksconsole.NewContext(t)
			mockCtx.EXPECT().Arguments().Return(test.arguments).Once()
			if test.failOn != "" {
				mockCtx.EXPECT().Error(assert.AnError.Error()).Once()
			}

			var names []string
			HandleEach(mockCtx, func(_ console.Context, name string) error {
				names = append(names, name)
				if test.failOn != "" && name == test.failOn {
					return assert.AnError
				}

				return nil
			})

			assert.Equal(t, test.wantNames, names)
		})
	}
}

func TestConfirmToProceed(t *testing.T) {
	var (
		mockCtx *mocksconsole.Context
	)

	beforeEach := func() {
		mockCtx = mocksconsole.NewContext(t)
	}

	tests := []struct {
		name         string
		env          string
		setup        func()
		expectResult bool
	}{
		{
			name:         "env is not production",
			setup:        func() {},
			expectResult: true,
		},
		{
			name: "the force option is true",
			env:  "production",
			setup: func() {
				mockCtx.EXPECT().OptionBool("force").Return(true).Once()
			},
			expectResult: true,
		},
		{
			name: "confirm returns false",
			env:  "production",
			setup: func() {
				mockCtx.EXPECT().OptionBool("force").Return(false).Once()
				mockCtx.EXPECT().Confirm("Are you sure you want to run this command?").Return(false).Once()
			},
			expectResult: false,
		},
		{
			name: "confirm returns true",
			env:  "production",
			setup: func() {
				mockCtx.EXPECT().OptionBool("force").Return(false).Once()
				mockCtx.EXPECT().Confirm("Are you sure you want to run this command?").Return(true).Once()
			},
			expectResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()
			result := ConfirmToProceed(mockCtx, tt.env)

			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestTwoColumnDetail(t *testing.T) {
	tests := []struct {
		name   string
		first  string
		second string
		output string
	}{
		{
			name:  "only has first column",
			first: color.Yellow().Sprint("Name"),
			output: "  " + color.Yellow().Sprint("Name") + " " +
				color.Gray().Sprint(strings.Repeat(".", pterm.GetTerminalWidth()-len("Name")-5)) + "  ",
		},
		{
			name:   "has first and second column",
			first:  "Test",
			second: color.Green().Sprint("Passed"),
			output: "  Test " + color.Gray().Sprint(
				strings.Repeat(".", pterm.GetTerminalWidth()-len("Test")-len("Passed")-6),
			) + " " + color.Green().Sprint("Passed") + "  ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.output, TwoColumnDetail(tt.first, tt.second))
		})
	}
}

func TestModuleFlag(t *testing.T) {
	moduleFlag, ok := ModuleFlag().(*command.StringFlag)

	assert.True(t, ok)
	assert.Equal(t, "module", moduleFlag.Name)
	assert.NotEmpty(t, moduleFlag.Usage)
}
