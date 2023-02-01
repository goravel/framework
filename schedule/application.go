package schedule

import (
	"github.com/gookit/color"
	"github.com/robfig/cron/v3"

	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/schedule/support"
)

type Application struct {
	cron *cron.Cron
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Call(callback func()) schedule.Event {
	return &support.Event{Callback: callback}
}

func (app *Application) Command(command string) schedule.Event {
	return &support.Event{Command: command}
}

func (app *Application) Register(events []schedule.Event) {
	if app.cron == nil {
		app.cron = cron.New(cron.WithLogger(&Logger{}))
	}

	app.addEvents(events)
}

func (app *Application) Run() {
	app.cron.Start()
}

func (app *Application) addEvents(events []schedule.Event) {
	for _, event := range events {
		chain := cron.NewChain()
		if event.GetDelayIfStillRunning() {
			chain = cron.NewChain(cron.DelayIfStillRunning(&Logger{}))
		} else if event.GetSkipIfStillRunning() {
			chain = cron.NewChain(cron.SkipIfStillRunning(&Logger{}))
		}
		_, err := app.cron.AddJob(event.GetCron(), chain.Then(app.getJob(event)))

		if err != nil {
			facades.Log.Errorf("add schedule error: %v", err)
		}
	}
}

func (app *Application) getJob(event schedule.Event) cron.Job {
	return cron.FuncJob(func() {
		if event.GetCommand() != "" {
			facades.Artisan.Call(event.GetCommand())
		} else {
			event.GetCallback()()
		}
	})
}

type Logger struct{}

func (log *Logger) Info(msg string, keysAndValues ...any) {
	color.Green.Printf("%s %v\n", msg, keysAndValues)
}

func (log *Logger) Error(err error, msg string, keysAndValues ...any) {
	facades.Log.Error(msg, keysAndValues)
}
