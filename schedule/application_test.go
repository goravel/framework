package schedule

import (
	"context"
	"strings"
	"testing"
	"time"

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
	mockArtisan.EXPECT().Call("test --name Goravel argument0 argument1").Return(nil).Times(2)

	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Error("panic", mock.Anything).Return().Times(4)

	immediatelyCall := 0
	delayIfStillRunningCall := 0
	skipIfStillRunningCall := 0

	app := NewApplication(mockArtisan, nil, mockLog, false)
	app.Register([]schedule.Event{
		app.Call(func() {
			panic(1)
		}).Cron("* * * * * *"),
		app.Call(func() {
			immediatelyCall++
		}).Cron("* * * * * *"),
		app.Call(func() {
			time.Sleep(2 * time.Second)
			delayIfStillRunningCall++
		}).Cron("* * * * * *").DelayIfStillRunning(),
		app.Call(func() {
			time.Sleep(2 * time.Second)
			skipIfStillRunningCall++
		}).Cron("* * * * * *").SkipIfStillRunning(),
		app.Command("test --name Goravel argument0 argument1").Cron("*/2 * * * * *"),
	})

	go app.Run()

	time.Sleep(4 * time.Second)

	s.Equal(4, immediatelyCall)
	s.Equal(4, delayIfStillRunningCall)
	s.Equal(2, skipIfStillRunningCall)

	s.NoError(app.Shutdown())
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

	time.Sleep(2 * time.Second)

	s.NoError(app.Shutdown())
	s.Equal(1, immediatelyCall)
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

	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	s.EqualError(app.Shutdown(ctx), "context deadline exceeded")
	s.Equal(0, immediatelyCall)
}
