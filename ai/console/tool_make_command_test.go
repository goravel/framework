package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestToolMakeCommand(t *testing.T) {
	toolMakeCommand := &ToolMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	assert.Equal(t, "make:tool", toolMakeCommand.Signature())
	assert.Equal(t, "Create a new agent tool", toolMakeCommand.Description())

	extend := toolMakeCommand.Extend()
	assert.Equal(t, "make", extend.Category)
	if assert.Len(t, extend.Flags, 1) {
		flag, ok := extend.Flags[0].(*command.BoolFlag)
		assert.True(t, ok)
		if ok {
			assert.Equal(t, "force", flag.Name)
			assert.Equal(t, []string{"f"}, flag.Aliases)
			assert.Equal(t, "Create the tool even if it already exists", flag.Usage)
		}
	}

	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the tool name", mock.Anything).Return("", errors.New("the tool name cannot be empty")).Once()
	mockContext.EXPECT().Error("the tool name cannot be empty").Once()
	assert.NoError(t, toolMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("WeatherTool").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Tool created successfully").Once()
	assert.NoError(t, toolMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/tools/weather_tool.go"))

	mockContext.EXPECT().Argument(0).Return("WeatherTool").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the tool already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, toolMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("user/WeatherTool").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Tool created successfully").Once()
	assert.NoError(t, toolMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/tools/user/weather_tool.go"))
	assert.True(t, file.Contain("app/tools/user/weather_tool.go", "package user"))
	assert.True(t, file.Contain("app/tools/user/weather_tool.go", "type WeatherTool struct"))
	assert.True(t, file.Contain("app/tools/user/weather_tool.go", "func (r *WeatherTool) Name() string"))
	assert.True(t, file.Contain("app/tools/user/weather_tool.go", "return \"weather_tool\""))
	assert.True(t, file.Contain("app/tools/user/weather_tool.go", "func (r *WeatherTool) Description() string"))
	assert.True(t, file.Contain("app/tools/user/weather_tool.go", "func (r *WeatherTool) Parameters() map[string]any"))
	assert.True(t, file.Contain("app/tools/user/weather_tool.go", "func (r *WeatherTool) Execute(ctx context.Context, args map[string]any) (string, error)"))
	assert.True(t, file.Contain("app/tools/user/weather_tool.go", "var _ ai.Tool = (*WeatherTool)(nil)"))
	assert.NoError(t, file.Remove("app"))
}
