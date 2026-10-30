package schedule

import (
	"context"
	"time"
)

type Schedule interface {
	// Call add a new callback event to the schedule.
	Call(callback func()) Event
	// Command adds a new Artisan command event to the schedule.
	Command(command string) Event
	// Events returns all registered schedule events.
	Events() []Event
	// Register schedules.
	Register(events []Event)
	// Run schedules.
	Run()
	// RunDue runs the events due in the minute of the given time once, one
	// after the other, and returns when they have finished. It is meant for
	// external triggers that start a short-lived process every minute (system
	// cron, Kubernetes CronJob, serverless schedulers) instead of keeping Run
	// alive. The returned error joins the errors of the events that failed.
	RunDue(ctx context.Context, at time.Time) error
	// Shutdown schedules.
	Shutdown(ctx ...context.Context) error
}
