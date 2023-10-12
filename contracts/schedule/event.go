package schedule

//go:generate mockery --name=Event
type Event interface {
	// At schedule the event to run at the specified time.
	At(time string) Event
	// Cron schedule the event using the given Cron expression.
	Cron(expression string) Event
	// Daily schedule the event to run daily.
	Daily() Event
	// DailyAt schedule the event to run daily at a given time (10:00, 19:30, etc).
	DailyAt(time string) Event
	// DelayIfStillRunning if the event is still running, the event will be delayed.
	DelayIfStillRunning() Event
	// EveryMinute schedule the event to run every minute.
	EveryMinute() Event
	// EveryTwoMinutes schedule the event to run every two minutes.
	EveryTwoMinutes() Event
	// EveryThreeMinutes schedule the event to run every three minutes.
	EveryThreeMinutes() Event
	// EveryFourMinutes schedule the event to run every four minutes.
	EveryFourMinutes() Event
	// EveryFiveMinutes schedule the event to run every five minutes.
	EveryFiveMinutes() Event
	// EveryTenMinutes schedule the event to run every ten minutes.
	EveryTenMinutes() Event
	// EveryFifteenMinutes schedule the event to run every fifteen minutes.
	EveryFifteenMinutes() Event
	// EveryThirtyMinutes schedule the event to run every thirty minutes.
	EveryThirtyMinutes() Event
	// EveryTwoHours schedule the event to run every two hours.
	EveryTwoHours() Event
	// EveryThreeHours schedule the event to run every three hours.
	EveryThreeHours() Event
	// EveryFourHours schedule the event to run every four hours.
	EveryFourHours() Event
	// EverySixHours schedule the event to run every six hours.
	EverySixHours() Event
	// GetCron get cron expression.
	GetCron() string
	// GetCommand get the command.
	GetCommand() string
	// GetCallback get callback.
	GetCallback() func()
	// GetName get name.
	GetName() string
	// GetSkipIfStillRunning get skipIfStillRunning bool.
	GetSkipIfStillRunning() bool
	// GetDelayIfStillRunning get delayIfStillRunning bool.
	GetDelayIfStillRunning() bool
	// Hourly schedule the event to run hourly.
	Hourly() Event
	// HourlyAt schedule the event to run hourly at a given offset in the hour.
	HourlyAt(offset []string) Event
	// IsOnOneServer get isOnOneServer bool.
	IsOnOneServer() bool
	// Name set the event name.
	Name(name string) Event
	// OnOneServer only allow the event to run on one server for each cron expression.
	OnOneServer() Event
	// SkipIfStillRunning if the event is still running, the event will be skipped.
	SkipIfStillRunning() Event
}
