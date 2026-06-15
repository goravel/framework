package console

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/console/command"
	consolemocks "github.com/goravel/framework/mocks/console"
	schedulemocks "github.com/goravel/framework/mocks/schedule"
)

func TestRunCommand(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	mockContext := consolemocks.NewContext(t)
	runCommand := NewRun(mockSchedule)

	ctx, cancel := context.WithCancel(context.Background())

	runCh := make(chan struct{})
	mockContext.EXPECT().Context().Return(ctx)
	mockSchedule.EXPECT().Run().Run(func() {
		close(runCh)
	})
	mockSchedule.EXPECT().Shutdown(ctx).Return(nil)

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
