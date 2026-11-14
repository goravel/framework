package console

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/support/file"
)

type ViewMakeCommandTestSuite struct {
	suite.Suite
}

func TestViewMakeCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ViewMakeCommandTestSuite))
}

func (s *ViewMakeCommandTestSuite) TestSignature() {
	cmd := &ViewMakeCommand{}
	s.Equal("make:view", cmd.Signature())
}

func (s *ViewMakeCommandTestSuite) TestDescription() {
	cmd := &ViewMakeCommand{}
	s.Equal("Create a new view file", cmd.Description())
}

func (s *ViewMakeCommandTestSuite) TestExtend() {
	cmd := &ViewMakeCommand{}
	extend := cmd.Extend()

	s.Equal("make", extend.Category)
	s.Len(extend.Flags, 2)

	// Test BoolFlag
	boolFlag, ok := extend.Flags[0].(*command.BoolFlag)
	s.True(ok)
	s.Equal("force", boolFlag.Name)
	s.Equal([]string{"f"}, boolFlag.Aliases)
	s.Equal("Create the view even if it already exists", boolFlag.Usage)
}

func (s *ViewMakeCommandTestSuite) TestEmptyName() {
	viewMakeCommand := &ViewMakeCommand{}
	mockContext := mocksconsole.NewContext(s.T())

	mockContext.EXPECT().Arguments().Return([]string{""}).Once()
	mockContext.EXPECT().Error("the view name name cannot be empty").Once()
	s.Nil(viewMakeCommand.Handle(mockContext))
}

func (s *ViewMakeCommandTestSuite) TestCreateSuccess() {
	viewMakeCommand := &ViewMakeCommand{}
	mockContext := mocksconsole.NewContext(s.T())

	mockContext.EXPECT().Arguments().Return([]string{"welcome"}).Once()
	mockContext.EXPECT().Option("module").Return("").Once()

	expectedPath := filepath.Join("resources", "views", "welcome.tmpl")
	mockContext.EXPECT().Success("View created successfully").Once()

	s.Nil(viewMakeCommand.Handle(mockContext))
	s.True(file.Exists(expectedPath))
	s.Nil(file.Remove("resources"))
}

func (s *ViewMakeCommandTestSuite) TestGetStub() {
	cmd := &ViewMakeCommand{}
	stub := cmd.getStub()

	s.NotEmpty(stub)
	s.Contains(stub, "DummyDefinition")
	s.Contains(stub, "{{ define")
	s.Contains(stub, "{{ end }}")
}

func (s *ViewMakeCommandTestSuite) TestPopulateStub() {
	cmd := &ViewMakeCommand{}

	tests := []struct {
		name           string
		stub           string
		definition     string
		expectedResult string
	}{
		{
			name:           "basic replacement",
			stub:           "{{ define \"DummyDefinition\" }}\n<h1>Welcome</h1>\n{{ end }}\n",
			definition:     "welcome.tmpl",
			expectedResult: "{{ define \"welcome.tmpl\" }}\n<h1>Welcome</h1>\n{{ end }}\n",
		},
		{
			name:           "subdirectory path",
			stub:           "{{ define \"DummyDefinition\" }}\n<h1>Welcome</h1>\n{{ end }}\n",
			definition:     "admin/dashboard.tmpl",
			expectedResult: "{{ define \"admin/dashboard.tmpl\" }}\n<h1>Welcome</h1>\n{{ end }}\n",
		},
		{
			name:           "custom definition name",
			stub:           "{{ define \"DummyDefinition\" }}\n<h1>Welcome</h1>\n{{ end }}\n",
			definition:     "home",
			expectedResult: "{{ define \"home\" }}\n<h1>Welcome</h1>\n{{ end }}\n",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := cmd.populateStub(tt.stub, tt.definition)
			s.Equal(tt.expectedResult, result)
		})
	}
}

func (s *ViewMakeCommandTestSuite) TestHandleMultipleNames() {
	s.T().Cleanup(func() {
		_ = file.Remove("resources")
	})

	viewMakeCommand := &ViewMakeCommand{}
	mockContext := mocksconsole.NewContext(s.T())

	mockContext.EXPECT().Arguments().Return([]string{"welcome", "dashboard"}).Once()
	mockContext.EXPECT().Option("module").Return("").Times(2)
	mockContext.EXPECT().Success("View created successfully").Times(2)

	s.Nil(viewMakeCommand.Handle(mockContext))
	s.FileExists(filepath.Join("resources", "views", "welcome.tmpl"))
	s.FileExists(filepath.Join("resources", "views", "dashboard.tmpl"))
}

// blogModule sets its own directory, so it is a module of the application.
type blogModule struct{}

func (r *blogModule) Name() string { return "blog" }
func (r *blogModule) Path() string { return "modules/blog" }

func (s *ViewMakeCommandTestSuite) TestModule() {
	registry := modular.NewRegistry("")
	s.NoError(registry.Register(&blogModule{}))
	modular.SetDefault(registry)
	s.T().Cleanup(func() {
		modular.SetDefault(nil)
		_ = file.Remove("modules")
	})

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Arguments().Return([]string{"posts/index"}).Once()
	mockContext.EXPECT().Option("module").Return("Blog").Once()
	mockContext.EXPECT().Success("View created successfully").Once()

	s.Nil((&ViewMakeCommand{}).Handle(mockContext))
	s.True(file.Contain(filepath.Join("modules", "blog", "resources", "views", "posts", "index.tmpl"), `{{ define "blog/posts/index.tmpl" }}`))
	s.NoDirExists("resources")
}

func (s *ViewMakeCommandTestSuite) TestModuleNotFound() {
	modular.SetDefault(modular.NewRegistry(""))
	s.T().Cleanup(func() { modular.SetDefault(nil) })

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Arguments().Return([]string{"posts/index"}).Once()
	mockContext.EXPECT().Option("module").Return("shop").Once()
	mockContext.EXPECT().Error(errors.ModularModuleNotFound.Args("shop").Error()).Once()

	s.Nil((&ViewMakeCommand{}).Handle(mockContext))
	s.NoDirExists("resources")
	s.NoDirExists("modules")
}
