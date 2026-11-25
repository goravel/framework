package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/schedule"
)

type Work struct {
	schedule schedule.Schedule
}

func NewWork(schedule schedule.Schedule) *Work {
	return &Work{
		schedule: schedule,
	}
}

// Signature The name and signature of the console command.
func (r *Work) Signature() string {
	return "schedule:work"
}

// Description The console command description.
func (r *Work) Description() string {
	return "Start the schedule worker, which runs the scheduled commands until stopped"
}

// Extend The console command extend.
func (r *Work) Extend() command.Extend {
	return command.Extend{
		Category: "schedule",
	}
}

// Handle Execute the console command.
func (r *Work) Handle(ctx console.Context) error {
	r.schedule.Work()
	return nil
}

func (r *Work) Shutdown(ctx console.Context) error {
	return r.schedule.Shutdown(ctx)
}
