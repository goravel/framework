package console

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
	assert.Equal(t, command.Extend{
		Category: "schedule",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:  "once",
				Usage: "run the events due in the current minute and exit, for an external trigger such as cron",
			},
		},
	}, runCommand.Extend())
}

func TestRunCommand_Shutdown(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	runCommand := NewRun(mockSchedule)

	mockContext := consolemocks.NewContext(t)
	mockSchedule.EXPECT().Shutdown(mockContext).Return(nil).Once()

	assert.NoError(t, runCommand.Shutdown(mockContext))
}

func TestRunCommand_Handle(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	runCommand := NewRun(mockSchedule)

	mockContext := consolemocks.NewContext(t)
	mockContext.EXPECT().OptionBool("once").Return(false).Once()
	mockSchedule.EXPECT().Run().Once()
	assert.NoError(t, runCommand.Handle(mockContext))

	mockContext.EXPECT().OptionBool("once").Return(true).Twice()
	mockSchedule.EXPECT().RunDue(mock.Anything, mock.MatchedBy(func(at time.Time) bool {
		return time.Since(at) < time.Minute
	})).Return(nil).Once()
	assert.NoError(t, runCommand.Handle(mockContext))

	mockSchedule.EXPECT().RunDue(mock.Anything, mock.Anything).Return(assert.AnError).Once()
	assert.Equal(t, assert.AnError, runCommand.Handle(mockContext))
}
