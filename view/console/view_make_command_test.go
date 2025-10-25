package console

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestViewMakeCommandSignature(t *testing.T) {
	cmd := &ViewMakeCommand{}
	assert.Equal(t, "make:view", cmd.Signature())
}

func TestViewMakeCommandDescription(t *testing.T) {
	cmd := &ViewMakeCommand{}
	assert.Equal(t, "Create a new view file", cmd.Description())
}

func TestViewMakeCommandExtend(t *testing.T) {
	cmd := &ViewMakeCommand{}
	extend := cmd.Extend()

	assert.Equal(t, "make", extend.Category)
	assert.Len(t, extend.Flags, 2)

	// Test StringFlag
	stringFlag, ok := extend.Flags[0].(*command.StringFlag)
	assert.True(t, ok)
	assert.Equal(t, "path", stringFlag.Name)
	assert.Equal(t, "resources/views", stringFlag.Value)
	assert.Equal(t, "The path where the view file should be created", stringFlag.Usage)

	// Test BoolFlag
	boolFlag, ok := extend.Flags[1].(*command.BoolFlag)
	assert.True(t, ok)
	assert.Equal(t, "force", boolFlag.Name)
	assert.Equal(t, []string{"f"}, boolFlag.Aliases)
	assert.Equal(t, "Create the view even if it already exists", boolFlag.Usage)
}

func TestViewMakeCommand_EmptyName(t *testing.T) {
	viewMakeCommand := &ViewMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Error("the view name name cannot be empty").Once()
	assert.Nil(t, viewMakeCommand.Handle(mockContext))
}

func TestViewMakeCommand_CreateSuccess(t *testing.T) {
	viewMakeCommand := &ViewMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	mockContext.EXPECT().Argument(0).Return("welcome.tmpl").Once()
	mockContext.EXPECT().Option("path").Return("resources/views").Once()

	expectedPath := filepath.Join("resources", "views", "welcome.tmpl")
	mockContext.EXPECT().Success("View created successfully: " + expectedPath).Once()

	assert.Nil(t, viewMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists(expectedPath))
	file.Remove("resources")
}

func TestViewMakeCommandGetStub(t *testing.T) {
	cmd := &ViewMakeCommand{}
	stub := cmd.getStub()

	assert.NotEmpty(t, stub)
	assert.Contains(t, stub, "DummyPathName")
	assert.Contains(t, stub, "DummyViewName")
	assert.Contains(t, stub, "DummyPathDefinition")
	assert.Contains(t, stub, "{{ define")
	assert.Contains(t, stub, "{{ end }}")
}

func TestViewMakeCommandPopulateStub(t *testing.T) {
	cmd := &ViewMakeCommand{}

	tests := []struct {
		name           string
		stub           string
		viewName       string
		viewPath       string
		expectedResult string
	}{
		{
			name:           "basic replacement",
			stub:           "// DummyPathName\n{{ define \"DummyPathDefinition\" }}\n<h1>Welcome to DummyViewName</h1>\n{{ end }}",
			viewName:       "welcome.tmpl",
			viewPath:       "resources/views/welcome.tmpl",
			expectedResult: "// resources/views/welcome.tmpl\n{{ define \"welcome.tmpl\" }}\n<h1>Welcome to welcome.tmpl</h1>\n{{ end }}",
		},
		{
			name:           "subdirectory path",
			stub:           "// DummyPathName\n{{ define \"DummyPathDefinition\" }}\n<h1>Welcome to DummyViewName</h1>\n{{ end }}",
			viewName:       "admin/dashboard.tmpl",
			viewPath:       "resources/views/admin/dashboard.tmpl",
			expectedResult: "// resources/views/admin/dashboard.tmpl\n{{ define \"admin/dashboard.tmpl\" }}\n<h1>Welcome to admin/dashboard.tmpl</h1>\n{{ end }}",
		},
		{
			name:           "custom path - should not remove custom path prefix",
			stub:           "// DummyPathName\n{{ define \"DummyPathDefinition\" }}\n<h1>Welcome to DummyViewName</h1>\n{{ end }}",
			viewName:       "home.tmpl",
			viewPath:       "custom/views/home.tmpl",
			expectedResult: "// custom/views/home.tmpl\n{{ define \"custom/views/home.tmpl\" }}\n<h1>Welcome to home.tmpl</h1>\n{{ end }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.populateStub(tt.stub, tt.viewName, tt.viewPath)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestViewMakeCommandPopulateStubWithResourcesPath(t *testing.T) {
	cmd := &ViewMakeCommand{}

	// Test that resources/views/ prefix is removed from path definition
	stub := "// DummyPathName\n{{ define \"DummyPathDefinition\" }}\n<h1>Welcome to DummyViewName</h1>\n{{ end }}"
	viewName := "welcome.tmpl"
	viewPath := "resources/views/welcome.tmpl"

	result := cmd.populateStub(stub, viewName, viewPath)

	// The path definition should have resources/views/ removed
	assert.Contains(t, result, "{{ define \"welcome.tmpl\" }}")
	assert.NotContains(t, result, "{{ define \"resources/views/welcome.tmpl\" }}")
}

func TestViewMakeCommandPopulateStubWithNonResourcesPath(t *testing.T) {
	cmd := &ViewMakeCommand{}

	// Test that non-resources paths are not modified
	stub := "// DummyPathName\n{{ define \"DummyPathDefinition\" }}\n<h1>Welcome to DummyViewName</h1>\n{{ end }}"
	viewName := "custom.tmpl"
	viewPath := "custom/views/custom.tmpl"

	result := cmd.populateStub(stub, viewName, viewPath)

	// The path definition should remain unchanged for non-resources paths
	assert.Contains(t, result, "{{ define \"custom/views/custom.tmpl\" }}")
}
