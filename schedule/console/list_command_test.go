package console

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/schedule"
	consolemocks "github.com/goravel/framework/mocks/console"
	schedulemocks "github.com/goravel/framework/mocks/schedule"
	"github.com/goravel/framework/support/carbon"
)

type ListCommandTestSuite struct {
	suite.Suite
}

func (s *ListCommandTestSuite) SetupTest() {
	carbon.SetTestNow(carbon.Now().StartOfMinute())
}

func (s *ListCommandTestSuite) TearDownTest() {
	carbon.UnsetTestNow()
}

func TestListCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListCommandTestSuite))
}

func (s *ListCommandTestSuite) TestListCommand() {
	listCommand := &List{}
	s.Equal("schedule:list", listCommand.Signature())
	s.Equal("List all scheduled tasks", listCommand.Description())
	s.Equal(command.Extend{Category: "schedule"}, listCommand.Extend())

	tests := []struct {
		name     string
		expected [2]string
		setup    func(mockEvent *schedulemocks.Event)
	}{
		{
			name: "schedule artisan command",
			expected: [2]string{
				"<fg=yellow>* * * * *</>  artisan send:emails <fg=yellow>name</>",
				"<fg=7472a3>Next Due: 1 minute after</>",
			},
			setup: func(mockEvent *schedulemocks.Event) {
				mockEvent.EXPECT().GetName().Return("")
				mockEvent.EXPECT().GetCommand().Return("send:emails name")
				mockEvent.EXPECT().GetCron().Return("* * * * *")
			},
		},
		{
			name: "schedule closure command(without name)",
			expected: [2]string{
				fmt.Sprintf("<fg=yellow>* * * * *</>  Closure at: %s:65", filepath.Join("schedule", "console", "list_command_test.go")),
				"<fg=7472a3>Next Due: 1 minute after</>",
			},
			setup: func(mockEvent *schedulemocks.Event) {
				mockEvent.EXPECT().GetName().Return("")
				mockEvent.EXPECT().GetCommand().Return("")
				mockEvent.EXPECT().GetCallback().Return(func() {})
				mockEvent.EXPECT().GetCron().Return("* * * * *")
			},
		},
		{
			name: "schedule closure command(with name)",
			expected: [2]string{
				"<fg=yellow>* * * * *</>  test-command",
				"<fg=7472a3>Next Due: 1 minute after</>",
			},
			setup: func(mockEvent *schedulemocks.Event) {
				mockEvent.EXPECT().GetName().Return("test-command")
				mockEvent.EXPECT().GetCommand().Return("")
				mockEvent.EXPECT().GetCron().Return("* * * * *")
			},
		},
		{
			name: "schedule invalid cron expression",
			expected: [2]string{
				"<fg=yellow>* *</>  invalid-cron",
				"",
			},
			setup: func(mockEvent *schedulemocks.Event) {
				mockEvent.EXPECT().GetName().Return("invalid-cron")
				mockEvent.EXPECT().GetCommand().Return("")
				mockEvent.EXPECT().GetCron().Return("* *")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockContext := consolemocks.NewContext(s.T())
			mockSchedule := schedulemocks.NewSchedule(s.T())
			mockEvent := schedulemocks.NewEvent(s.T())

			mockSchedule.EXPECT().Events().Return([]schedule.Event{mockEvent}).Once()
			mockContext.EXPECT().NewLine().Return().Once()
			mockContext.EXPECT().TwoColumnDetail(tt.expected[0], tt.expected[1]).Once()
			tt.setup(mockEvent)

			s.NoError(NewListCommand(mockSchedule).Handle(mockContext))
		})
	}
}
