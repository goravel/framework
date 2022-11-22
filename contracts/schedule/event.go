package schedule

//go:generate mockery --name=Event
type Event interface {
	GetCron() string
	GetCommand() string
	GetCallback() func()
	GetSkipIfStillRunning() bool
	GetDelayIfStillRunning() bool
	Cron(expression string) Event
	EveryMinute() Event
	EveryTwoMinutes() Event
	EveryThreeMinutes() Event
	EveryFourMinutes() Event
	EveryFiveMinutes() Event
	EveryTenMinutes() Event
	EveryFifteenMinutes() Event
	EveryThirtyMinutes() Event
	Hourly() Event
	HourlyAt(offset []string) Event
	EveryTwoHours() Event
	EveryThreeHours() Event
	EveryFourHours() Event
	EverySixHours() Event
	Daily() Event
	At(time string) Event
	DailyAt(time string) Event
	SkipIfStillRunning() Event
	DelayIfStillRunning() Event
}
