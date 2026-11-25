package utils

import (
	"github.com/robfig/cron/v3"

	"github.com/goravel/framework/contracts/schedule"
)

// CronParser parses the cron expressions of the events, with optional seconds.
var CronParser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

// CronSpec returns the cron expression of the event with its timezone, in
// the form CronParser reads.
func CronSpec(event schedule.Event) string {
	if timezone := event.GetTimezone(); timezone != "" {
		return "CRON_TZ=" + timezone + " " + event.GetCron()
	}

	return event.GetCron()
}
