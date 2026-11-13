package schedule

import (
	"context"
	stderrors "errors"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
)

// parser parses the cron expressions of the events, with optional seconds.
var parser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

// maxDelayIfStillRunning bounds how long RunDue waits for a previous run of
// a DelayIfStillRunning event when the context has no deadline.
const maxDelayIfStillRunning = time.Minute

type Application struct {
	artisan  console.Artisan
	cache    cache.Cache
	log      log.Log
	cron     *cron.Cron
	location *time.Location
	events   []schedule.Event
	debug    bool
}

func NewApplication(artisan console.Artisan, cache cache.Cache, log log.Log, debug bool) *Application {
	location := carbon.Now().StdTime().Location()

	return &Application{
		artisan:  artisan,
		cache:    cache,
		cron:     cron.New(cron.WithParser(parser), cron.WithLogger(NewLogger(log, debug)), cron.WithLocation(location)),
		location: location,
		log:      log,
		debug:    debug,
	}
}

func (app *Application) Call(callback func()) schedule.Event {
	return NewCallbackEvent(callback)
}

func (app *Application) Command(command string) schedule.Event {
	return NewCommandEvent(command)
}

func (app *Application) Events() []schedule.Event {
	return app.events
}

func (app *Application) Register(events []schedule.Event) {
	app.addEvents(events)
}

func (app *Application) Run() {
	app.cron.Run()
}

// RunDue runs the events due in the minute of `at`, in the schedule's
// location. Events that fire more than once a minute run once. OnOneServer
// uses the same cache lock as Run; SkipIfStillRunning and
// DelayIfStillRunning use a cache lock held while the event runs, since the
// previous run may be in another process.
func (app *Application) RunDue(ctx context.Context, at time.Time) error {
	tick := at.In(app.location).Truncate(time.Minute)

	var errs []error
	for _, event := range app.events {
		if !app.isDue(event, tick) {
			continue
		}

		if event.IsOnOneServer() && event.GetName() != "" &&
			!app.cache.Lock(app.onOneServerKey(event, tick), 1*time.Hour).Get() {
			continue
		}

		release, ok := app.acquireOverlapLock(ctx, event)
		if !ok {
			continue
		}

		err := app.runEvent(event)
		release()

		if err != nil {
			app.log.Error(err)
			errs = append(errs, err)
		}
	}

	return stderrors.Join(errs...)
}

func (app *Application) Shutdown(ctx ...context.Context) error {
	if len(ctx) == 0 {
		ctx = append(ctx, context.Background())
	}

	cronCtx := app.cron.Stop()

	for {
		select {
		case <-cronCtx.Done():
			return nil
		case <-ctx[0].Done():
			return ctx[0].Err()
		}
	}
}

func (app *Application) addEvents(events []schedule.Event) {
	for _, event := range events {
		chain := cron.NewChain(cron.Recover(NewLogger(app.log, app.debug)))
		if event.GetDelayIfStillRunning() {
			chain = cron.NewChain(cron.DelayIfStillRunning(NewLogger(app.log, app.debug)), cron.Recover(NewLogger(app.log, app.debug)))
		} else if event.GetSkipIfStillRunning() {
			chain = cron.NewChain(cron.SkipIfStillRunning(NewLogger(app.log, app.debug)), cron.Recover(NewLogger(app.log, app.debug)))
		}
		_, err := app.cron.AddJob(event.GetCron(), chain.Then(app.getJob(event)))

		if err != nil {
			app.log.Errorf("add schedule error: %v", err)
			continue
		}
		app.events = append(app.events, event)
	}
}

func (app *Application) getJob(event schedule.Event) cron.Job {
	return cron.FuncJob(func() {
		if event.IsOnOneServer() && event.GetName() != "" {
			if app.cache.Lock(app.onOneServerKey(event, carbon.Now().StdTime()), 1*time.Hour).Get() {
				app.runJob(event)
			}
		} else {
			app.runJob(event)
		}
	})
}

// onOneServerKey is the cache lock key that makes an OnOneServer event run
// once per minute (or per second for cron expressions with seconds).
func (app *Application) onOneServerKey(event schedule.Event, at time.Time) string {
	moment := carbon.FromStdTime(at)
	keySuffix := moment.Format("Hi")
	if segments := strings.Split(event.GetCron(), " "); len(segments) == 6 {
		keySuffix = moment.Format("His")
	}

	return event.GetName() + keySuffix
}

func (app *Application) runJob(event schedule.Event) {
	if event.GetCommand() != "" {
		if err := app.artisan.Call(event.GetCommand()); err != nil {
			app.log.Errorf("run %s command error: %v", event.GetCommand(), err)
		}
	} else {
		event.GetCallback()()
	}
}

// isDue reports whether the event fires within the minute starting at tick.
func (app *Application) isDue(event schedule.Event, tick time.Time) bool {
	spec, err := parser.Parse(event.GetCron())
	if err != nil {
		// Register only keeps events with a valid expression.
		return false
	}

	return spec.Next(tick.Add(-time.Nanosecond)).Before(tick.Add(time.Minute))
}

// acquireOverlapLock prevents a SkipIfStillRunning or DelayIfStillRunning
// event from overlapping a run started by another process. It returns false
// when the event has to be skipped.
func (app *Application) acquireOverlapLock(ctx context.Context, event schedule.Event) (func(), bool) {
	noop := func() {}
	if app.cache == nil || event.GetName() == "" || (!event.GetSkipIfStillRunning() && !event.GetDelayIfStillRunning()) {
		return noop, true
	}

	lock := app.cache.Lock(event.GetName()+":running", 24*time.Hour)

	var acquired bool
	if event.GetDelayIfStillRunning() {
		wait := maxDelayIfStillRunning
		if deadline, ok := ctx.Deadline(); ok {
			wait = min(wait, time.Until(deadline))
		}
		acquired = lock.Block(wait)
	} else {
		acquired = lock.Get()
	}

	if !acquired {
		return noop, false
	}

	return func() { lock.Release() }, true
}

func (app *Application) runEvent(event schedule.Event) (err error) {
	name := event.GetName()
	if name == "" {
		name = event.GetCron()
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			err = errors.ScheduleEventPanicked.Args(name, recovered)
		}
	}()

	if event.GetCommand() != "" {
		if callErr := app.artisan.Call(event.GetCommand()); callErr != nil {
			return errors.ScheduleEventFailed.Args(name, callErr)
		}

		return nil
	}

	event.GetCallback()()

	return nil
}
