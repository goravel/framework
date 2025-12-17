package console

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
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
	s.Len(extend.Flags, 1)

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

	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Error("the view name name cannot be empty").Once()
	s.Nil(viewMakeCommand.Handle(mockContext))
}

func (s *ViewMakeCommandTestSuite) TestCreateSuccess() {
	viewMakeCommand := &ViewMakeCommand{}
	mockContext := mocksconsole.NewContext(s.T())

	mockContext.EXPECT().Argument(0).Return("welcome").Once()

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
