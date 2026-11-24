package schedule

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/errors"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mockslog "github.com/goravel/framework/mocks/log"
)

type ApplicationTestSuite struct {
	suite.Suite
}

func TestApplicationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {
}

func (s *ApplicationTestSuite) TestCallAndCommand() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	var commandCall atomic.Int64
	mockArtisan.EXPECT().Call("test --name Goravel argument0 argument1").RunAndReturn(func(string) error {
		commandCall.Add(1)
		return nil
	}).Times(3)

	mockLog := mockslog.NewLog(s.T())
	var panicCall atomic.Int64
	mockLog.EXPECT().Error("panic", mock.Anything).Run(func(args ...any) {
		panicCall.Add(1)
	}).Return().Times(3)

	var immediatelyCall atomic.Int64
	var delayIfStillRunningCall atomic.Int64
	var skipIfStillRunningCall atomic.Int64
	shutdownErr := make(chan error, 1)

	app := NewApplication(mockArtisan, nil, mockLog, false)
	app.Register([]schedule.Event{
		app.Call(func() {
			panic(1)
		}).Cron("* * * * * *"),
		app.Call(func() {
			if immediatelyCall.Add(1) == 3 {
				go func() {
					shutdownErr <- app.Shutdown()
				}()
			}
		}).Cron("* * * * * *"),
		app.Call(func() {
			time.Sleep(1100 * time.Millisecond)
			delayIfStillRunningCall.Add(1)
		}).Cron("* * * * * *").DelayIfStillRunning(),
		app.Call(func() {
			time.Sleep(2500 * time.Millisecond)
			skipIfStillRunningCall.Add(1)
		}).Cron("* * * * * *").SkipIfStillRunning(),
		app.Command("test --name Goravel argument0 argument1").Cron("* * * * * *"),
	})

	go app.Work()

	select {
	case err := <-shutdownErr:
		s.NoError(err)
	case <-time.After(10 * time.Second):
		s.NoError(app.Shutdown())
		s.FailNow("timed out waiting for scheduler shutdown")
	}

	s.Equal(int64(3), immediatelyCall.Load())
	s.Equal(int64(3), delayIfStillRunningCall.Load())
	s.Equal(int64(1), skipIfStillRunningCall.Load())
	s.Equal(int64(3), commandCall.Load())
	s.Equal(int64(3), panicCall.Load())
}

func (s *ApplicationTestSuite) TestOnOneServer() {
	mockCache := mockscache.NewCache(s.T())
	mockLock := mockscache.NewLock(s.T())

	// The execution order is not stable in the Windows system, so we don't use Once() here.
	mockLock.EXPECT().Get().Return(true)
	mockCache.EXPECT().Lock(mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "immediately") && len(key) == 17
	}), 1*time.Hour).Return(mockLock)

	mockCache1 := mockscache.NewCache(s.T())
	mockLock1 := mockscache.NewLock(s.T())
	mockLock1.EXPECT().Get().Return(false)
	mockCache1.EXPECT().Lock(mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "immediately") && len(key) == 17
	}), 1*time.Hour).Return(mockLock1)

	immediatelyCall1 := 0
	immediatelyCall2 := 0

	app := NewApplication(nil, mockCache, nil, false)
	app.Register([]schedule.Event{
		app.Call(func() {
			immediatelyCall1++
		}).Cron("* * * * * *").OnOneServer().Name("immediately"),
	})

	app1 := NewApplication(nil, mockCache1, nil, false)
	app1.Register([]schedule.Event{
		app1.Call(func() {
			immediatelyCall2++
		}).Cron("* * * * * *").OnOneServer().Name("immediately"),
	})

	go app.Work()
	go app1.Work()

	time.Sleep(2 * time.Second)

	s.True(immediatelyCall1 > 0)
	s.True(immediatelyCall2 == 0)

	s.NoError(app.Shutdown())
	s.NoError(app1.Shutdown())
}

func (s *ApplicationTestSuite) TestShutdown() {
	immediatelyCall := 0

	app := NewApplication(nil, nil, nil, false)
	app.Register([]schedule.Event{
		app.Call(func() {
			time.Sleep(4 * time.Second)
			immediatelyCall++
		}).Cron("*/2 * * * * *"),
	})

	go app.Work()

	time.Sleep(3 * time.Second)

	s.NoError(app.Shutdown())

	// Due to the millisecond precision, the immediatelyCall may be 1 or 2.
	s.True(immediatelyCall >= 1 && immediatelyCall <= 2)
}

func (s *ApplicationTestSuite) TestShutdownWithContext() {
	immediatelyCall := 0

	app := NewApplication(nil, nil, nil, false)
	app.Register([]schedule.Event{
		app.Call(func() {
			time.Sleep(10 * time.Second)
			immediatelyCall++
		}).Cron("*/2 * * * * *"),
	})

	go app.Work()

	time.Sleep(3 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	s.EqualError(app.Shutdown(ctx), "context deadline exceeded")
	s.Equal(0, immediatelyCall)
}

func (s *ApplicationTestSuite) TestRunDue() {
	location := time.FixedZone("UTC-3", -3*60*60)
	// 12:05:37 UTC is 09:05 in the schedule's location.
	at := time.Date(2026, 10, 18, 12, 5, 37, 0, time.UTC)

	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("every-five-minutes").Return(nil).Once()
	mockArtisan.EXPECT().Call("daily-at-09:05").Return(nil).Once()

	callbackCall := 0
	app := NewApplication(mockArtisan, nil, nil, false)
	app.location = location
	app.Register([]schedule.Event{
		app.Command("every-five-minutes").EveryFiveMinutes(),
		app.Command("daily-at-09:05").DailyAt("09:05"),
		app.Command("daily-at-12:05").DailyAt("12:05"),
		app.Command("every-ten-minutes").EveryTenMinutes(),
		// Runs once per tick even though it fires every 10 seconds.
		app.Call(func() { callbackCall++ }).Cron("*/10 * * * * *"),
	})

	s.NoError(app.runDue(context.Background(), at))
	s.Equal(1, callbackCall)
}

func (s *ApplicationTestSuite) TestRunDueOnOneServer() {
	// 12:30 UTC is 09:30 in the schedule's location, which the key uses.
	at := time.Date(2026, 10, 18, 12, 30, 0, 0, time.UTC)

	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("reconcile").Return(nil).Once()

	mockCache := mockscache.NewCache(s.T())
	mockLock := mockscache.NewLock(s.T())
	mockCache.EXPECT().Lock("reconcile0930", 1*time.Hour).Return(mockLock).Twice()
	mockLock.EXPECT().Get().Return(true).Once()
	// A duplicate trigger in the same minute doesn't run it again.
	mockLock.EXPECT().Get().Return(false).Once()

	app := NewApplication(mockArtisan, mockCache, nil, false)
	app.location = time.FixedZone("UTC-3", -3*60*60)
	app.Register([]schedule.Event{
		app.Command("reconcile").EveryThirtyMinutes().OnOneServer(),
	})

	s.NoError(app.runDue(context.Background(), at))
	s.NoError(app.runDue(context.Background(), at.Add(20*time.Second)))
}

func (s *ApplicationTestSuite) TestRunDueSkipIfStillRunning() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("long").Return(nil).Once()

	mockCache := mockscache.NewCache(s.T())
	mockLock := mockscache.NewLock(s.T())
	mockCache.EXPECT().Lock("long:running", 24*time.Hour).Return(mockLock).Twice()
	// Another process is still running it: skipped.
	mockLock.EXPECT().Get().Return(false).Once()
	// Then it is free: it runs and releases the lock.
	mockLock.EXPECT().Get().Return(true).Once()
	mockLock.EXPECT().Release().Return(true).Once()

	app := NewApplication(mockArtisan, mockCache, nil, false)
	app.Register([]schedule.Event{
		app.Command("long").EveryMinute().SkipIfStillRunning(),
	})

	s.NoError(app.runDue(context.Background(), time.Now()))
	s.NoError(app.runDue(context.Background(), time.Now()))
}

func (s *ApplicationTestSuite) TestRunDueDelayIfStillRunning() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("long").Return(nil).Once()

	mockCache := mockscache.NewCache(s.T())
	mockLock := mockscache.NewLock(s.T())
	mockCache.EXPECT().Lock("long:running", 24*time.Hour).Return(mockLock).Twice()
	// Waits for the previous run, bounded by the context deadline.
	mockLock.EXPECT().Block(mock.MatchedBy(func(wait time.Duration) bool {
		return wait > 4*time.Second && wait <= 5*time.Second
	})).Return(true).Once()
	mockLock.EXPECT().Release().Return(true).Once()
	// Without a deadline it waits at most maxDelayIfStillRunning, then skips.
	mockLock.EXPECT().Block(maxDelayIfStillRunning).Return(false).Once()

	app := NewApplication(mockArtisan, mockCache, nil, false)
	app.Register([]schedule.Event{
		app.Command("long").EveryMinute().DelayIfStillRunning(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.NoError(app.runDue(ctx, time.Now()))
	s.NoError(app.runDue(context.Background(), time.Now()))
}

func (s *ApplicationTestSuite) TestRunDueReleasesLockOfFailedEvent() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("broken").Return(assert.AnError).Once()

	mockCache := mockscache.NewCache(s.T())
	mockLock := mockscache.NewLock(s.T())
	mockCache.EXPECT().Lock("broken:running", 24*time.Hour).Return(mockLock).Twice()
	mockLock.EXPECT().Get().Return(true).Twice()
	mockLock.EXPECT().Release().Return(true).Twice()

	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Error(mock.Anything).Return().Once()
	mockLog.EXPECT().Errorf("%v\n%s", mock.Anything, mock.Anything).Return().Once()

	app := NewApplication(mockArtisan, mockCache, mockLog, false)
	app.Register([]schedule.Event{
		app.Command("broken").EveryMinute().SkipIfStillRunning(),
		app.Call(func() { panic("boom") }).EveryMinute().Name("broken").SkipIfStillRunning(),
	})

	// The lock is released after a failed and after a panicking event.
	s.Error(app.runDue(context.Background(), time.Now()))
}

func (s *ApplicationTestSuite) TestRunDueSkipsNeverMatchingCron() {
	app := NewApplication(nil, nil, nil, false)
	app.Register([]schedule.Event{
		app.Call(func() { s.Fail("an expression that never matches must not run") }).Cron("0 0 30 2 *"),
	})

	s.NoError(app.runDue(context.Background(), time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)))
}

func (s *ApplicationTestSuite) TestRunDueErrors() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("broken").Return(assert.AnError).Once()
	mockArtisan.EXPECT().Call("fine").Return(nil).Once()

	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Error(mock.Anything).Return().Once()
	// A panic is logged with its stack.
	mockLog.EXPECT().Errorf("%v\n%s", mock.MatchedBy(func(err error) bool {
		return strings.Contains(err.Error(), "[schedule] schedule event panicking panicked: boom")
	}), mock.MatchedBy(func(stack []byte) bool {
		return strings.Contains(string(stack), "runtime/debug.Stack")
	})).Return().Once()

	app := NewApplication(mockArtisan, nil, mockLog, false)
	app.Register([]schedule.Event{
		app.Command("broken").EveryMinute(),
		app.Call(func() { panic("boom") }).EveryMinute().Name("panicking"),
		app.Command("fine").EveryMinute(),
	})

	// A failing event doesn't stop the others, and every failure is returned.
	err := app.runDue(context.Background(), time.Now())
	s.ErrorContains(err, "schedule event broken failed: "+assert.AnError.Error())
	s.ErrorContains(err, "[schedule] schedule event panicking panicked: boom")
}

func (s *ApplicationTestSuite) TestRunDueSkipsInvalidCron() {
	// Register drops events with an invalid expression; runDue ignores them
	// too if one gets into the list another way.
	app := NewApplication(mocksconsole.NewArtisan(s.T()), nil, nil, false)
	app.events = append(app.events, app.Command("invalid").Cron("not a cron"))

	s.NoError(app.runDue(context.Background(), time.Now()))
}

func (s *ApplicationTestSuite) TestRun() {
	callbackCall := 0
	app := NewApplication(nil, nil, nil, false)
	app.Register([]schedule.Event{
		app.Call(func() { callbackCall++ }).EveryMinute(),
	})

	s.NoError(app.Run(context.Background()))
	s.Equal(1, callbackCall)
}

func (s *ApplicationTestSuite) TestRunDueCancelledContext() {
	ctx, cancel := context.WithCancel(context.Background())

	callbackCall := 0
	app := NewApplication(nil, nil, nil, false)
	app.Register([]schedule.Event{
		app.Call(func() {
			callbackCall++
			cancel()
		}).EveryMinute(),
		// Not started once the context is cancelled.
		app.Call(func() { callbackCall++ }).EveryMinute(),
	})

	s.ErrorIs(app.runDue(ctx, time.Now()), context.Canceled)
	s.Equal(1, callbackCall)
}

func (s *ApplicationTestSuite) TestRunDueEvery() {
	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Warningf("schedule event %s uses %s, which only schedule:work supports", "hourly", "@every 1h").Return().Once()

	callbackCall := 0
	app := NewApplication(nil, nil, mockLog, false)
	app.Register([]schedule.Event{
		app.Call(func() { callbackCall++ }).Cron("@every 30s"),
		app.Call(func() { callbackCall++ }).Cron("@every 1m"),
		app.Call(func() { callbackCall += 10 }).Cron("@every 1h").Name("hourly"),
	})

	s.NoError(app.runDue(context.Background(), time.Now()))
	s.Equal(2, callbackCall)
}

func (s *ApplicationTestSuite) TestRunDueWarnsUnnamedEvent() {
	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Warningf("schedule event %s has no name, so OnOneServer, SkipIfStillRunning and DelayIfStillRunning are not applied", "* * * * *").Return().Once()

	callbackCall := 0
	app := NewApplication(nil, nil, mockLog, false)
	app.Register([]schedule.Event{
		app.Call(func() { callbackCall++ }).EveryMinute().SkipIfStillRunning(),
	})

	s.NoError(app.runDue(context.Background(), time.Now()))
	s.Equal(1, callbackCall)
}

func (s *ApplicationTestSuite) TestRunDueHooks() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("fine").Return(nil).Once()
	mockArtisan.EXPECT().Call("broken").Return(assert.AnError).Once()

	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Error(mock.Anything).Return().Once()
	mockLog.EXPECT().Errorf("%v\n%s", mock.Anything, mock.Anything).Return().Once()

	var calls []string
	var failures []error
	hooks := func(event schedule.Event, name string) schedule.Event {
		return event.
			Before(func() { calls = append(calls, name+":before") }).
			OnSuccess(func() { calls = append(calls, name+":success") }).
			OnFailure(func(err error) {
				calls = append(calls, name+":failure")
				failures = append(failures, err)
			}).
			After(func() { calls = append(calls, name+":after") })
	}

	app := NewApplication(mockArtisan, nil, mockLog, false)
	app.Register([]schedule.Event{
		hooks(app.Command("fine").EveryMinute(), "fine"),
		hooks(app.Command("broken").EveryMinute(), "broken"),
		hooks(app.Call(func() {
			calls = append(calls, "panicking:run")
			panic("boom")
		}).EveryMinute().Name("panicking"), "panicking"),
	})

	s.Error(app.runDue(context.Background(), time.Now()))
	s.Equal([]string{
		"fine:before", "fine:success", "fine:after",
		"broken:before", "broken:failure", "broken:after",
		"panicking:before", "panicking:run", "panicking:failure", "panicking:after",
	}, calls)
	s.Require().Len(failures, 2)
	s.ErrorIs(failures[0], assert.AnError)
	s.EqualError(failures[1], "[schedule] schedule event panicking panicked: boom")
}

func (s *ApplicationTestSuite) TestRunDueHooksRunInRegistrationOrder() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("fine").Return(nil).Once()
	mockArtisan.EXPECT().Call("broken").Return(assert.AnError).Once()

	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Error(mock.Anything).Return().Once()

	var calls []string
	hooks := func(event schedule.Event, name string) schedule.Event {
		return event.
			Before(func() { calls = append(calls, name+":before:1") }).
			Before(func() { calls = append(calls, name+":before:2") }).
			After(func() { calls = append(calls, name+":after:1") }).
			OnFailure(func(error) { calls = append(calls, name+":failure:1") }).
			OnSuccess(func() { calls = append(calls, name+":success:1") }).
			After(func() { calls = append(calls, name+":after:2") }).
			OnFailure(func(error) { calls = append(calls, name+":failure:2") }).
			OnSuccess(func() { calls = append(calls, name+":success:2") })
	}

	app := NewApplication(mockArtisan, nil, mockLog, false)
	app.Register([]schedule.Event{
		hooks(app.Command("fine").EveryMinute(), "fine"),
		hooks(app.Command("broken").EveryMinute(), "broken"),
	})

	s.Error(app.runDue(context.Background(), time.Now()))
	s.Equal([]string{
		"fine:before:1", "fine:before:2",
		"fine:after:1", "fine:success:1", "fine:after:2", "fine:success:2",
		"broken:before:1", "broken:before:2",
		"broken:after:1", "broken:failure:1", "broken:after:2", "broken:failure:2",
	}, calls)
}

func (s *ApplicationTestSuite) TestRunDueAfterHookPanics() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("broken").Return(assert.AnError).Once()

	mockLog := mockslog.NewLog(s.T())
	// The panicking hook is logged, and the event keeps its own error.
	mockLog.EXPECT().Error(errors.ScheduleEventHookPanicked.Args("broken", "boom")).Return().Once()
	mockLog.EXPECT().Error(mock.Anything).Return().Once()

	var calls []string
	app := NewApplication(mockArtisan, nil, mockLog, false)
	app.Register([]schedule.Event{
		app.Command("broken").EveryMinute().
			After(func() { panic("boom") }).
			OnFailure(func(err error) { calls = append(calls, "failure:"+err.Error()) }).
			After(func() { calls = append(calls, "after") }),
	})

	s.EqualError(app.runDue(context.Background(), time.Now()), errors.ScheduleEventFailed.Args("broken", assert.AnError).Error())
	s.Equal([]string{"failure:" + assert.AnError.Error(), "after"}, calls)
}

func (s *ApplicationTestSuite) TestRunDueBeforeHookPanics() {
	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Errorf("%v\n%s", mock.Anything, mock.Anything).Return().Once()

	var calls []string
	app := NewApplication(nil, nil, mockLog, false)
	app.Register([]schedule.Event{
		app.Call(func() { calls = append(calls, "run") }).EveryMinute().Name("guarded").
			Before(func() { panic("not now") }).
			OnSuccess(func() { calls = append(calls, "success") }).
			OnFailure(func(error) { calls = append(calls, "failure") }).
			After(func() { calls = append(calls, "after") }),
	})

	// The event doesn't run, and neither do the other hooks.
	s.ErrorContains(app.runDue(context.Background(), time.Now()), "schedule event guarded panicked: not now")
	s.Empty(calls)
}

func (s *ApplicationTestSuite) TestWorkHooks() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("broken").Return(assert.AnError).Once()

	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Errorf("run %s command error: %v", "broken", assert.AnError).Return().Once()

	var calls []string
	app := NewApplication(mockArtisan, nil, mockLog, false)
	broken := app.Command("broken").EveryMinute().
		OnFailure(func(err error) { calls = append(calls, "broken:failure:"+err.Error()) }).
		After(func() { calls = append(calls, "broken:after") })
	panicking := app.Call(func() { panic("boom") }).EveryMinute().Name("panicking").
		OnFailure(func(err error) { calls = append(calls, "panicking:failure:"+err.Error()) }).
		After(func() { calls = append(calls, "panicking:after") })

	app.getJob(broken).Run()
	// The panic is passed on once the hooks ran, so the cron chain recovers
	// and logs it like before.
	s.PanicsWithValue("boom", func() { app.getJob(panicking).Run() })

	s.Equal([]string{
		"broken:failure:" + assert.AnError.Error(), "broken:after",
		"panicking:failure:[schedule] schedule event panicking panicked: boom", "panicking:after",
	}, calls)
}

func (s *ApplicationTestSuite) TestRunDueTimezone() {
	santiago, err := time.LoadLocation("America/Santiago")
	s.Require().NoError(err)

	callbackCall := 0
	app := NewApplication(nil, nil, nil, false)
	app.Register([]schedule.Event{
		app.Call(func() { callbackCall++ }).DailyAt("09:00").Timezone("America/Santiago"),
	})

	// schedule:work evaluates the expression in the event's timezone too.
	s.Require().Len(app.cron.Entries(), 1)
	entry, ok := app.cron.Entries()[0].Schedule.(*cron.SpecSchedule)
	s.Require().True(ok)
	s.Equal("America/Santiago", entry.Location.String())

	// 09:00 in Santiago is a different minute in the schedule's timezone.
	s.NoError(app.runDue(context.Background(), time.Date(2026, 11, 20, 9, 0, 0, 0, santiago)))
	s.Equal(1, callbackCall)

	s.NoError(app.runDue(context.Background(), time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)))
	s.Equal(1, callbackCall)
}

func (s *ApplicationTestSuite) TestRegisterUnknownTimezone() {
	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Errorf("add schedule error: %v", mock.Anything).Return().Once()

	app := NewApplication(nil, nil, mockLog, false)
	app.Register([]schedule.Event{
		app.Call(func() {}).EveryMinute().Timezone("Mars/Olympus_Mons"),
	})

	s.Empty(app.Events())
}
