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
	listCommand := &List{}
	assert.Equal(t, "schedule:list", listCommand.Signature())
	assert.Equal(t, "List all scheduled tasks", listCommand.Description())
	assert.Equal(t, command.Extend{Category: "schedule"}, listCommand.Extend())

	tests := []struct {
		name     string
		expected []string
		setup    func(t *testing.T) []schedule.Event
	}{
		{
			name: "no schedule",
			expected: []string{
				"No scheduled tasks have been defined.",
			},
			setup: func(t *testing.T) []schedule.Event {
				return nil
			},
		},
		{
			name: "schedule artisan command",
			expected: []string{
				"<fg=yellow>* * * * *</>  artisan send:emails <fg=yellow>name</>",
				"<fg=7472a3>Next Due: 1 minute after</>",
			},
			setup: func(t *testing.T) []schedule.Event {
				mockEvent := mocksschedule.NewEvent(t)
				mockEvent.EXPECT().GetName().Return("").Once()
				mockEvent.EXPECT().GetCommand().Return("send:emails name").Once()
				mockEvent.EXPECT().GetCron().Return("* * * * *").Times(3)

				return []schedule.Event{mockEvent}
			},
		},
		{
			name: "schedule closure command(without name)",
			expected: []string{
				fmt.Sprintf("<fg=yellow>* * * * *</>  Closure at: %s:62", filepath.Join("schedule", "console", "list_command_test.go")),
				"<fg=7472a3>Next Due: 1 minute after</>",
			},
			setup: func(t *testing.T) []schedule.Event {
				mockEvent := mocksschedule.NewEvent(t)
				mockEvent.EXPECT().GetName().Return("").Once()
				mockEvent.EXPECT().GetCommand().Return("").Once()
				mockEvent.EXPECT().GetCallback().Return(func() {}).Twice()
				mockEvent.EXPECT().GetCron().Return("* * * * *").Times(3)

				return []schedule.Event{mockEvent}
			},
		},
		{
			name: "schedule closure command(with name)",
			expected: []string{
				"<fg=yellow>* * * * *</>  test-command",
				"<fg=7472a3>Next Due: 1 minute after</>",
			},
			setup: func(t *testing.T) []schedule.Event {
				mockEvent := mocksschedule.NewEvent(t)
				mockEvent.EXPECT().GetName().Return("test-command").Once()
				mockEvent.EXPECT().GetCommand().Return("").Once()
				mockEvent.EXPECT().GetCron().Return("* * * * *").Times(3)

				return []schedule.Event{mockEvent}
			},
		},
		{
			name: "schedule invalid cron expression",
			expected: []string{
				"<fg=yellow>* *</>  invalid-cron",
				"",
			},
			setup: func(t *testing.T) []schedule.Event {
				mockEvent := mocksschedule.NewEvent(t)
				mockEvent.EXPECT().GetName().Return("invalid-cron").Once()
				mockEvent.EXPECT().GetCommand().Return("").Once()
				mockEvent.EXPECT().GetCron().Return("* *").Times(3)

				return []schedule.Event{mockEvent}
			},
		},
	}

	carbon.SetTestNow(carbon.Now().StartOfMinute())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockContext := consolemocks.NewContext(t)
			mockSchedule := mocksschedule.NewSchedule(t)
			mockSchedule.EXPECT().Events().Return(tt.setup(t)).Once()
			mockContext.EXPECT().NewLine().Return().Once()

			if len(tt.expected) > 1 {
				mockContext.EXPECT().TwoColumnDetail(tt.expected[0], tt.expected[1]).Once()
			} else {
				mockContext.EXPECT().Warning(tt.expected[0]).Once()
			}

			assert.NoError(t, NewList(mockSchedule).Handle(mockContext))
		})
	}
	carbon.UnsetTestNow()
}
