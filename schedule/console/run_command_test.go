package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/console/command"
	consolemocks "github.com/goravel/framework/mocks/console"
	schedulemocks "github.com/goravel/framework/mocks/schedule"
)

func TestRunCommand(t *testing.T) {
	mockSchedule := schedulemocks.NewSchedule(t)
	runCommand := NewRun(mockSchedule)

	mockSchedule.EXPECT().Run().Once()

	assert.Equal(t, "schedule:run", runCommand.Signature())
	assert.Equal(t, "Run the scheduled commands", runCommand.Description())
	assert.Equal(t, command.Extend{Category: "schedule"}, runCommand.Extend())
	assert.NoError(t, runCommand.Handle(consolemocks.NewContext(t)))
}
