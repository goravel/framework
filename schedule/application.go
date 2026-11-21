package schedule

import (
	"context"
	"runtime/debug"
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

// cronParser parses the cron expressions of the events, with optional seconds.
var cronParser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

// maxDelayIfStillRunning bounds how long Run waits for a previous run of
// a DelayIfStillRunning event when the context has no deadline. It is well
// inside a minute, so the events behind it still run in the same tick.
const maxDelayIfStillRunning = 10 * time.Second

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
		cron:     cron.New(cron.WithParser(cronParser), cron.WithLogger(NewLogger(log, debug)), cron.WithLocation(location)),
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

// Work runs the schedules in a long-lived process, blocking until Shutdown.
func (app *Application) Work() {
	app.cron.Run()
}

// Run runs the events due in the current minute, in the schedule's
// location. Events that fire more than once a minute run once. OnOneServer
// uses the same cache lock as Work; SkipIfStillRunning and
// DelayIfStillRunning use a cache lock held while the event runs, since the
// previous run may be in another process. Once ctx is done, no further event
// starts and its error is returned with the errors of the events.
func (app *Application) Run(ctx context.Context) error {
	return app.runDue(ctx, carbon.Now().StdTime())
}

func (app *Application) runDue(ctx context.Context, at time.Time) error {
	tick := at.In(app.location).Truncate(time.Minute)

	var errs []error
	for _, event := range app.events {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(errs, err)...)
		}

		if !app.isDue(event, tick) {
			continue
		}

		if event.GetName() == "" && (event.IsOnOneServer() || event.GetSkipIfStillRunning() || event.GetDelayIfStillRunning()) {
			app.log.Warningf("schedule event %s has no name, so OnOneServer, SkipIfStillRunning and DelayIfStillRunning are not applied", event.GetCron())
		}

		if event.IsOnOneServer() && event.GetName() != "" &&
			!app.cache.Lock(app.onOneServerKey(event, tick), 1*time.Hour).Get() {
			continue
		}

		release, ok := app.acquireOverlapLock(ctx, event)
		if !ok {
			continue
		}

		if err := app.runEvent(event); err != nil {
			errs = append(errs, err)
		}
		release()
	}

	return errors.Join(errs...)
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
		_, err := app.cron.AddJob(cronSpec(event), chain.Then(app.getJob(event)))

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
	if err := app.execute(event); err != nil {
		app.log.Errorf("run %s command error: %v", event.GetCommand(), err)
	}
}

// execute runs the command or the callback of the event between its hooks:
// the Before callbacks, then the event, then the OnSuccess or the OnFailure
// callbacks and the After callbacks. Work and Run each report its failure
// their own way, so a panic of the event is passed on once the hooks ran.
func (app *Application) execute(event schedule.Event) (err error) {
	for _, callback := range event.GetBeforeCallbacks() {
		callback()
	}

	defer func() {
		recovered := recover()

		failure := err
		if recovered != nil {
			failure = errors.ScheduleEventPanicked.Args(eventName(event), recovered)
		}

		if failure != nil {
			for _, callback := range event.GetFailureCallbacks() {
				callback(failure)
			}
		} else {
			for _, callback := range event.GetSuccessCallbacks() {
				callback()
			}
		}

		for _, callback := range event.GetAfterCallbacks() {
			callback()
		}

		if recovered != nil {
			panic(recovered)
		}
	}()

	if event.GetCommand() != "" {
		return app.artisan.Call(event.GetCommand())
	}

	event.GetCallback()()

	return nil
}

// isDue reports whether the event fires within the minute starting at tick.
func (app *Application) isDue(event schedule.Event, tick time.Time) bool {
	spec, err := cronParser.Parse(cronSpec(event))
	if err != nil {
		// Register only keeps events with a valid expression and timezone.
		return false
	}

	// An @every expression has no fixed point in time, so only the long-lived
	// worker can follow it once its delay is longer than a minute.
	if delay, ok := spec.(cron.ConstantDelaySchedule); ok && delay.Delay > time.Minute {
		app.log.Warningf("schedule event %s uses %s, which only schedule:work supports", event.GetName(), event.GetCron())
		return false
	}

	// An expression that can never match, like February 30, has no next time.
	next := spec.Next(tick.Add(-time.Nanosecond))

	return !next.IsZero() && next.Before(tick.Add(time.Minute))
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

// runEvent runs the event and logs its failure, with the stack for a panic.
func (app *Application) runEvent(event schedule.Event) (err error) {
	name := eventName(event)

	defer func() {
		if recovered := recover(); recovered != nil {
			err = errors.ScheduleEventPanicked.Args(name, recovered)
			app.log.Errorf("%v\n%s", err, debug.Stack())
		}
	}()

	if execErr := app.execute(event); execErr != nil {
		err = errors.ScheduleEventFailed.Args(name, execErr)
		app.log.Error(err)
	}

	return err
}

// cronSpec returns the cron expression of the event with its timezone, in
// the form the cron parser reads.
func cronSpec(event schedule.Event) string {
	if timezone := event.GetTimezone(); timezone != "" {
		return "CRON_TZ=" + timezone + " " + event.GetCron()
	}

	return event.GetCron()
}

// eventName returns the name of the event, or its cron expression when it
// has none.
func eventName(event schedule.Event) string {
	if name := event.GetName(); name != "" {
		return name
	}

	return event.GetCron()
}
