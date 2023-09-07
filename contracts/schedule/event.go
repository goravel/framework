package schedule

//go:generate mockery --name=Event
type Event interface {
	// At Schedule the event to run at the specified time.
	At(time string) Event
	// Cron Schedule the event using the given Cron expression.
	Cron(expression string) Event
	// Daily Schedule the event to run daily.
	Daily() Event
	// DailyAt Schedule the event to run daily at a given time (10:00, 19:30, etc).
	DailyAt(time string) Event
	// DelayIfStillRunning If the event is still running, the event will be delayed.
	DelayIfStillRunning() Event
	// EveryMinute Schedule the event to run every minute.
	EveryMinute() Event
	// EveryTwoMinutes Schedule the event to run every two minutes.
	EveryTwoMinutes() Event
	// EveryThreeMinutes Schedule the event to run every three minutes.
	EveryThreeMinutes() Event
	// EveryFourMinutes Schedule the event to run every four minutes.
	EveryFourMinutes() Event
	// EveryFiveMinutes Schedule the event to run every five minutes.
	EveryFiveMinutes() Event
	// EveryTenMinutes Schedule the event to run every ten minutes.
	EveryTenMinutes() Event
	// EveryFifteenMinutes Schedule the event to run every fifteen minutes.
	EveryFifteenMinutes() Event
	// EveryThirtyMinutes Schedule the event to run every thirty minutes.
	EveryThirtyMinutes() Event
	// EveryTwoHours Schedule the event to run every two hours.
	EveryTwoHours() Event
	// EveryThreeHours Schedule the event to run every three hours.
	EveryThreeHours() Event
	// EveryFourHours Schedule the event to run every four hours.
	EveryFourHours() Event
	// EverySixHours Schedule the event to run every six hours.
	EverySixHours() Event
	// GetCron Get cron expression.
	GetCron() string
	// GetCommand Get command.
	GetCommand() string
	// GetCallback Get callback.
	GetCallback() func()
	// GetName Get name.
	GetName() string
	// GetSkipIfStillRunning Get skipIfStillRunning bool.
	GetSkipIfStillRunning() bool
	// GetDelayIfStillRunning Get delayIfStillRunning bool.
	GetDelayIfStillRunning() bool
	// Hourly Schedule the event to run hourly.
	Hourly() Event
	// HourlyAt Schedule the event to run hourly at a given offset in the hour.
	HourlyAt(offset []string) Event
	// IsOnOneServer Get isOnOneServer bool.
	IsOnOneServer() bool
	// Name Set the event name.
	Name(name string) Event
	// OnOneServer Only allow the event to run on one server for each cron expression.
	OnOneServer() Event
	// SkipIfStillRunning If the event is still running, the event will be skipped.
	SkipIfStillRunning() Event
}
