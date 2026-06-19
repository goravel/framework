package console

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	consolecontracts "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	consolemocks "github.com/goravel/framework/mocks/console"
	schedulemocks "github.com/goravel/framework/mocks/schedule"
)

func TestRunCommand(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	mockContext := consolemocks.NewContext(t)
	runCommand := NewRun(mockSchedule)

	_, ok := any(runCommand).(consolecontracts.Shutdownable)
	assert.True(t, ok)

	ctx, cancel := context.WithCancel(context.Background())

	runCh := make(chan struct{})
	mockContext.EXPECT().Done().Return(ctx.Done()).Once()
	mockSchedule.EXPECT().Run().Run(func() {
		close(runCh)
	}).Once()

	assert.Equal(t, "schedule:run", runCommand.Signature())
	assert.Equal(t, "Run the scheduled commands", runCommand.Description())
	assert.Equal(t, command.Extend{Category: "schedule"}, runCommand.Extend())

	errCh := make(chan error, 1)
	go func() {
		errCh <- runCommand.Handle(mockContext)
	}()

	<-runCh
	cancel()

	assert.NoError(t, <-errCh)
}

func TestRunCommand_ShutDown(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	runCommand := NewRun(mockSchedule)

	mockContext := consolemocks.NewContext(t)
	mockSchedule.EXPECT().Shutdown(mockContext).Return(nil).Once()

	assert.NoError(t, runCommand.ShutDown(mockContext))
}
