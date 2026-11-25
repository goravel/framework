package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolecontracts "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	consolemocks "github.com/goravel/framework/mocks/console"
	schedulemocks "github.com/goravel/framework/mocks/schedule"
)

func TestWorkCommand(t *testing.T) {
	workCommand := NewWork(schedulemocks.NewSchedule(t))

	_, ok := any(workCommand).(consolecontracts.Shutdownable)
	assert.True(t, ok)

	assert.Equal(t, "schedule:work", workCommand.Signature())
	assert.Equal(t, "Start the schedule worker, which runs the scheduled commands until stopped", workCommand.Description())
	assert.Equal(t, command.Extend{Category: "schedule"}, workCommand.Extend())
}

func TestWorkCommand_Handle(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	workCommand := NewWork(mockSchedule)

	mockSchedule.EXPECT().Work().Once()
	assert.NoError(t, workCommand.Handle(consolemocks.NewContext(t)))
}

func TestWorkCommand_Shutdown(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	workCommand := NewWork(mockSchedule)

	mockContext := consolemocks.NewContext(t)
	mockSchedule.EXPECT().Shutdown(mockContext).Return(nil).Once()

	assert.NoError(t, workCommand.Shutdown(mockContext))
}
