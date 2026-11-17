package schedule

import (
	"context"
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
	// Run runs the events due in the current minute once, one after the other,
	// and returns when they have finished. It is meant for external triggers
	// that start a short-lived process every minute (system cron, Kubernetes
	// CronJob, serverless schedulers). The returned error joins the errors of
	// the events that failed. Once ctx is done, no further event starts and
	// ctx's error is returned too.
	Run(ctx context.Context) error
	// Work runs the schedules in a long-lived process, blocking until Shutdown.
	Work()
	// Shutdown schedules.
	Shutdown(ctx ...context.Context) error
}
