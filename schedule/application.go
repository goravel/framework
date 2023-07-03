package schedule

import (
	"time"

	"github.com/gookit/color"
	"github.com/robfig/cron/v3"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/support/carbon"
)

type Application struct {
	artisan console.Artisan
	cache   cache.Cache
	cron    *cron.Cron
	log     log.Log
}

func NewApplication(artisan console.Artisan, cache cache.Cache, log log.Log) *Application {
	return &Application{
		artisan: artisan,
		cache:   cache,
		log:     log,
	}
}

func (app *Application) Call(callback func()) schedule.Event {
	return NewCallbackEvent(callback)
}

func (app *Application) Command(command string) schedule.Event {
	return NewCommandEvent(command)
}

func (app *Application) Register(events []schedule.Event) {
	if app.cron == nil {
		app.cron = cron.New(cron.WithLogger(NewLogger(app.log)))
	}

	app.addEvents(events)
}

func (app *Application) Run() {
	app.cron.Run()
}

func (app *Application) addEvents(events []schedule.Event) {
	for _, event := range events {
		chain := cron.NewChain(cron.Recover(NewLogger(app.log)))
		if event.GetDelayIfStillRunning() {
			chain = cron.NewChain(cron.DelayIfStillRunning(NewLogger(app.log)), cron.Recover(NewLogger(app.log)))
		} else if event.GetSkipIfStillRunning() {
			chain = cron.NewChain(cron.SkipIfStillRunning(NewLogger(app.log)), cron.Recover(NewLogger(app.log)))
		}
		_, err := app.cron.AddJob(event.GetCron(), chain.Then(app.getJob(event)))

		if err != nil {
			app.log.Errorf("add schedule error: %v", err)
		}
	}
}

func (app *Application) getJob(event schedule.Event) cron.Job {
	return cron.FuncJob(func() {
		if event.IsOnOneServer() && event.GetName() != "" {
			if app.cache.Lock(event.GetName()+carbon.Now().Format("Hi"), 1*time.Hour).Get() {
				app.runJob(event)
			}
		} else {
			app.runJob(event)
		}
	})
}

func (app *Application) runJob(event schedule.Event) {
	if event.GetCommand() != "" {
		app.artisan.Call(event.GetCommand())
	} else {
		event.GetCallback()()
	}
}

type Logger struct {
	log log.Log
}

func NewLogger(log log.Log) *Logger {
	return &Logger{
		log: log,
	}
}

func (log *Logger) Info(msg string, keysAndValues ...any) {
	color.Green.Printf("%s %v\n", msg, keysAndValues)
}

func (log *Logger) Error(err error, msg string, keysAndValues ...any) {
	log.log.Error(msg, keysAndValues)
}
