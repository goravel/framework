package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolecontracts "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	consolemocks "github.com/goravel/framework/mocks/console"
	schedulemocks "github.com/goravel/framework/mocks/schedule"
)

func TestRunCommand(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	runCommand := NewRun(mockSchedule)

	_, ok := any(runCommand).(consolecontracts.Shutdownable)
	assert.True(t, ok)

	assert.Equal(t, "schedule:run", runCommand.Signature())
	assert.Equal(t, "Run the scheduled commands", runCommand.Description())
	assert.Equal(t, command.Extend{Category: "schedule"}, runCommand.Extend())
}

func TestRunCommand_Shutdown(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	runCommand := NewRun(mockSchedule)

	mockContext := consolemocks.NewContext(t)
	mockSchedule.EXPECT().Shutdown(mockContext).Return(nil).Once()

	assert.NoError(t, runCommand.Shutdown(mockContext))
}
