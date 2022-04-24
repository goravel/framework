package schedule

import (
	"github.com/goravel/framework/schedule/support"
	"github.com/goravel/framework/support/facades"
	"github.com/robfig/cron/v3"
)

type Application struct {
	cron *cron.Cron
}

func (app *Application) Call(callback func()) *support.Event {
	return &support.Event{Callback: callback}
}

func (app *Application) Command(command string) *support.Event {
	return &support.Event{Command: command}
}

func (app *Application) Register(events []*support.Event) {
	if app.cron == nil {
		app.cron = cron.New(cron.WithLogger(&Logger{}))
	}

	app.addEvents(events)
}

func (app *Application) Run() {
	app.cron.Start()
}

func (app *Application) addEvents(events []*support.Event) {
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

func (app *Application) getJob(event *support.Event) cron.Job {
	return cron.FuncJob(func() {
		if event.Command != "" {
			facades.Artisan.Call(event.Command)
		} else {
			event.Callback()
		}
	})
}

type Logger struct{}

func (log *Logger) Info(msg string, keysAndValues ...interface{}) {
	facades.Log.Info(msg, keysAndValues)
}

func (log *Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	facades.Log.Error(msg, keysAndValues)
}
