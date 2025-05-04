package console

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/schedule"
	consolemocks "github.com/goravel/framework/mocks/console"
	mocksschedule "github.com/goravel/framework/mocks/schedule"
	"github.com/goravel/framework/support/carbon"
)

func TestListCommand(t *testing.T) {
	mockSchedule := mocksschedule.NewSchedule(t)
	listCommand := NewList(mockSchedule)
	mockContext := consolemocks.NewContext(t)

	assert.Equal(t, "schedule:list", listCommand.Signature())
	assert.Equal(t, "List all scheduled tasks", listCommand.Description())
	assert.Equal(t, command.Extend{Category: "schedule"}, listCommand.Extend())

	// no schedule
	mockSchedule.EXPECT().Events().Return(nil).Once()
	mockContext.EXPECT().NewLine().Once()
	mockContext.EXPECT().Warning("No scheduled tasks have been defined.").Once()
	assert.NoError(t, listCommand.Handle(mockContext))

	// schedule artisan command
	cmd := mocksschedule.NewEvent(t)
	cmd.EXPECT().GetCommand().Return("send:emails name").Once()
	cmd.EXPECT().GetCron().Return("* * * * *").Twice()
	mockContext.EXPECT().TwoColumnDetail(
		"<fg=yellow>  *    *  * * *</>  artisan send:emails <fg=yellow>name</>",
		"<fg=7472a3>Next Due: 1 minute after</>",
	).Once()

	// schedule closure command(without name)
	closure := mocksschedule.NewEvent(t)
	closure.EXPECT().GetName().Return("").Once()
	closure.EXPECT().GetCommand().Return("").Once()
	closure.EXPECT().GetCallback().Return(func() {}).Once()
	closure.EXPECT().GetCron().Return("*/30 * * * *").Twice()
	mockContext.EXPECT().TwoColumnDetail(
		fmt.Sprintf("<fg=yellow>  */30 *  * * *</>  Closure at: %s:45", filepath.Join("schedule", "console", "list_command_test.go")),
		"<fg=7472a3>Next Due: 30 minutes after</>",
	).Once()

	// schedule closure command(with name)
	namedClosure := mocksschedule.NewEvent(t)
	namedClosure.EXPECT().GetName().Return("test-command").Once()
	namedClosure.EXPECT().GetCommand().Return("").Once()
	namedClosure.EXPECT().GetCron().Return("00 10 * * *").Twice()
	mockContext.EXPECT().TwoColumnDetail(
		"<fg=yellow>  00   10 * * *</>  test-command",
		"<fg=7472a3>Next Due: 10 hours after</>",
	).Once()

	// schedule command(second-level cron)
	secondlyCommand := mocksschedule.NewEvent(t)
	secondlyCommand.EXPECT().GetCommand().Return("do:something --every --second").Once()
	secondlyCommand.EXPECT().GetCron().Return("* * * * * *").Twice()
	mockContext.EXPECT().TwoColumnDetail(
		"<fg=yellow>* *    *  * * *</>  artisan do:something <fg=yellow>--every --second</>",
		"<fg=7472a3>Next Due: 1 second after</>",
	).Once()

	mockSchedule.EXPECT().Events().Return([]schedule.Event{
		cmd,
		closure,
		namedClosure,
		secondlyCommand,
	}).Once()
	mockContext.EXPECT().NewLine().Once()

	carbon.SetTestNow(carbon.Now().StartOfDay())
	assert.NoError(t, NewList(mockSchedule).Handle(mockContext))
	carbon.UnsetTestNow()
}
