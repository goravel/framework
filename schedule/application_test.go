package schedule

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/schedule"
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

	go app.Run()

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

	go app.Run()
	go app1.Run()

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

	go app.Run()

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

	go app.Run()

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

	s.NoError(app.RunDue(context.Background(), at))
	s.Equal(1, callbackCall)
}

func (s *ApplicationTestSuite) TestRunDueOnOneServer() {
	at := time.Date(2026, 10, 18, 12, 30, 0, 0, time.UTC)

	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("reconcile").Return(nil).Once()

	mockCache := mockscache.NewCache(s.T())
	mockLock := mockscache.NewLock(s.T())
	mockCache.EXPECT().Lock("reconcile1230", 1*time.Hour).Return(mockLock).Twice()
	mockLock.EXPECT().Get().Return(true).Once()
	// A duplicate trigger in the same minute doesn't run it again.
	mockLock.EXPECT().Get().Return(false).Once()

	app := NewApplication(mockArtisan, mockCache, nil, false)
	app.location = time.UTC
	app.Register([]schedule.Event{
		app.Command("reconcile").EveryThirtyMinutes().OnOneServer(),
	})

	s.NoError(app.RunDue(context.Background(), at))
	s.NoError(app.RunDue(context.Background(), at.Add(20*time.Second)))
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

	s.NoError(app.RunDue(context.Background(), time.Now()))
	s.NoError(app.RunDue(context.Background(), time.Now()))
}

func (s *ApplicationTestSuite) TestRunDueDelayIfStillRunning() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("long").Return(nil).Once()

	mockCache := mockscache.NewCache(s.T())
	mockLock := mockscache.NewLock(s.T())
	mockCache.EXPECT().Lock("long:running", 24*time.Hour).Return(mockLock).Twice()
	// Waits for the previous run, bounded by the context deadline.
	mockLock.EXPECT().Block(mock.MatchedBy(func(wait time.Duration) bool {
		return wait > 0 && wait <= 10*time.Second
	})).Return(true).Once()
	mockLock.EXPECT().Release().Return(true).Once()
	// Without a deadline it waits at most a minute, then skips.
	mockLock.EXPECT().Block(maxDelayIfStillRunning).Return(false).Once()

	app := NewApplication(mockArtisan, mockCache, nil, false)
	app.Register([]schedule.Event{
		app.Command("long").EveryMinute().DelayIfStillRunning(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.NoError(app.RunDue(ctx, time.Now()))
	s.NoError(app.RunDue(context.Background(), time.Now()))
}

func (s *ApplicationTestSuite) TestRunDueErrors() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("broken").Return(assert.AnError).Once()
	mockArtisan.EXPECT().Call("fine").Return(nil).Once()

	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Error(mock.Anything).Return().Twice()

	app := NewApplication(mockArtisan, nil, mockLog, false)
	app.Register([]schedule.Event{
		app.Command("broken").EveryMinute(),
		app.Call(func() { panic("boom") }).EveryMinute().Name("panicking"),
		app.Command("fine").EveryMinute(),
	})

	// A failing event doesn't stop the others, and every failure is returned.
	err := app.RunDue(context.Background(), time.Now())
	s.ErrorContains(err, "schedule event broken failed: "+assert.AnError.Error())
	s.ErrorContains(err, "schedule event panicking panicked: boom")
}
