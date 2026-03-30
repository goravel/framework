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

func TestAgentMakeCommand(t *testing.T) {
	agentMakeCommand := &AgentMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	assert.Equal(t, "make:agent", agentMakeCommand.Signature())
	assert.Equal(t, "Create a new agent", agentMakeCommand.Description())

	extend := agentMakeCommand.Extend()
	assert.Equal(t, "make", extend.Category)
	if assert.Len(t, extend.Flags, 1) {
		flag, ok := extend.Flags[0].(*command.BoolFlag)
		assert.True(t, ok)
		if ok {
			assert.Equal(t, "force", flag.Name)
			assert.Equal(t, []string{"f"}, flag.Aliases)
			assert.Equal(t, "Create the agent even if it already exists", flag.Usage)
		}
	}

	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the agent name", mock.Anything).Return("", errors.New("the agent name cannot be empty")).Once()
	mockContext.EXPECT().Error("the agent name cannot be empty").Once()
	assert.NoError(t, agentMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("UserAgent").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Agent created successfully").Once()
	assert.NoError(t, agentMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/agents/user_agent.go"))

	mockContext.EXPECT().Argument(0).Return("UserAgent").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the agent already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, agentMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("user/SupportAgent").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Agent created successfully").Once()
	assert.NoError(t, agentMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/agents/user/support_agent.go"))
	assert.True(t, file.Contain("app/agents/user/support_agent.go", "package user"))
	assert.True(t, file.Contain("app/agents/user/support_agent.go", "type SupportAgent struct"))
	assert.True(t, file.Contain("app/agents/user/support_agent.go", "func (r *SupportAgent) Instructions() string"))
	assert.True(t, file.Contain("app/agents/user/support_agent.go", "func (r *SupportAgent) Messages() []ai.Message"))
	assert.NoError(t, file.Remove("app"))
}
