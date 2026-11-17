package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/console/command"
	consolemocks "github.com/goravel/framework/mocks/console"
	schedulemocks "github.com/goravel/framework/mocks/schedule"
)

func TestRunCommand(t *testing.T) {
	runCommand := NewRun(schedulemocks.NewSchedule(t))

	assert.Equal(t, "schedule:run", runCommand.Signature())
	assert.Equal(t, "Run the scheduled commands due in the current minute once and exit", runCommand.Description())
	assert.Equal(t, command.Extend{Category: "schedule"}, runCommand.Extend())
}

func TestRunCommand_Handle(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	runCommand := NewRun(mockSchedule)
	mockContext := consolemocks.NewContext(t)

	mockSchedule.EXPECT().Run(mockContext).Return(nil).Once()
	assert.NoError(t, runCommand.Handle(mockContext))

	mockSchedule.EXPECT().Run(mockContext).Return(assert.AnError).Once()
	assert.Equal(t, assert.AnError, runCommand.Handle(mockContext))
}
