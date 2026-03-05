package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsschedule "github.com/goravel/framework/contracts/schedule"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksschedule "github.com/goravel/framework/mocks/schedule"
)

func TestScheduleRunner(t *testing.T) {
	t.Run("signature", func(t *testing.T) {
		runner := &ScheduleRunner{}
		assert.Equal(t, "schedule", runner.Signature())
	})

	t.Run("should run when schedule has events and auto_run enabled", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		schedule := mocksschedule.NewSchedule(t)
		event := mocksschedule.NewEvent(t)

		schedule.EXPECT().Events().Return([]contractsschedule.Event{event}).Once()
		config.EXPECT().GetBool("app.auto_run", true).Return(true).Once()

		runner := NewScheduleRunner(config, schedule)
		assert.True(t, runner.ShouldRun())
	})

	t.Run("should not run when schedule facade not set", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		runner := NewScheduleRunner(config, nil)
		assert.False(t, runner.ShouldRun())
	})

	t.Run("should not run when no events", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		schedule := mocksschedule.NewSchedule(t)
		schedule.EXPECT().Events().Return(nil).Once()

		runner := NewScheduleRunner(config, schedule)
		assert.False(t, runner.ShouldRun())
	})

	t.Run("should not run when auto_run disabled", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		schedule := mocksschedule.NewSchedule(t)
		event := mocksschedule.NewEvent(t)

		schedule.EXPECT().Events().Return([]contractsschedule.Event{event}).Once()
		config.EXPECT().GetBool("app.auto_run", true).Return(false).Once()

		runner := NewScheduleRunner(config, schedule)
		assert.False(t, runner.ShouldRun())
	})

	t.Run("run and shutdown", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		schedule := mocksschedule.NewSchedule(t)
		schedule.EXPECT().Run().Once()
		schedule.EXPECT().Shutdown().Return(nil).Once()

		runner := NewScheduleRunner(config, schedule)
		assert.NoError(t, runner.Run())
		assert.NoError(t, runner.Shutdown())
	})
}
