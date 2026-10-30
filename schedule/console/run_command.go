package console

import (
	"context"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/support/carbon"
)

type Run struct {
	schedule schedule.Schedule
}

func NewRun(schedule schedule.Schedule) *Run {
	return &Run{
		schedule: schedule,
	}
}

// Signature The name and signature of the console command.
func (r *Run) Signature() string {
	return "schedule:run"
}

// Description The console command description.
func (r *Run) Description() string {
	return "Run the scheduled commands"
}

// Extend The console command extend.
func (r *Run) Extend() command.Extend {
	return command.Extend{
		Category: "schedule",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:  "once",
				Usage: "run the events due in the current minute and exit, for an external trigger such as cron",
			},
		},
	}
}

// Handle Execute the console command.
func (r *Run) Handle(ctx console.Context) error {
	if ctx.OptionBool("once") {
		return r.schedule.RunDue(context.Background(), carbon.Now().StdTime())
	}

	r.schedule.Run()
	return nil
}

func (r *Run) Shutdown(ctx console.Context) error {
	return r.schedule.Shutdown(ctx)
}
