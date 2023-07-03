package schedule

//go:generate mockery --name=Event
type Event interface {
	At(time string) Event
	Cron(expression string) Event
	Daily() Event
	DailyAt(time string) Event
	DelayIfStillRunning() Event
	EveryMinute() Event
	EveryTwoMinutes() Event
	EveryThreeMinutes() Event
	EveryFourMinutes() Event
	EveryFiveMinutes() Event
	EveryTenMinutes() Event
	EveryFifteenMinutes() Event
	EveryThirtyMinutes() Event
	EveryTwoHours() Event
	EveryThreeHours() Event
	EveryFourHours() Event
	EverySixHours() Event
	GetCron() string
	GetCommand() string
	GetCallback() func()
	GetName() string
	GetSkipIfStillRunning() bool
	GetDelayIfStillRunning() bool
	Hourly() Event
	HourlyAt(offset []string) Event
	IsOnOneServer() bool
	Name(name string) Event
	OnOneServer() Event
	SkipIfStillRunning() Event
}
