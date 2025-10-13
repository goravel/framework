package schedule

import (
	"github.com/goravel/framework/contracts/schedule"
)

type ScheduleRunner struct {
	schedule schedule.Schedule
}

func NewScheduleRunner(schedule schedule.Schedule) *ScheduleRunner {
	return &ScheduleRunner{
		schedule: schedule,
	}
}

func (r *ScheduleRunner) ShouldRun() bool {
	return r.schedule != nil && len(r.schedule.Events()) > 0
}

func (r *ScheduleRunner) Run() error {
	r.schedule.Run()

	return nil
}

func (r *ScheduleRunner) Shutdown() error {
	return r.schedule.Shutdown()
}
