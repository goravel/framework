package schedule

import (
	"context"
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
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {
}

func (s *ApplicationTestSuite) TestCallAndCommand() {
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Call("test --name Goravel argument0 argument1").Return(nil).Once()

	mockLog := mockslog.NewLog(s.T())
	mockLog.EXPECT().Error("panic", mock.Anything).Return().Once()

	immediatelyCall := 0
	delayIfStillRunningCall := 0
	skipIfStillRunningCall := 0

	app := NewApplication(mockArtisan, nil, mockLog, false)
	app.Register([]schedule.Event{
		app.Call(func() {
			panic(1)
		}).EveryMinute(),
		app.Call(func() {
			immediatelyCall++
		}).EveryMinute(),
		app.Call(func() {
			time.Sleep(3 * time.Second)
			delayIfStillRunningCall++
		}).Cron("*/2 * * * * *").DelayIfStillRunning(),
		app.Call(func() {
			time.Sleep(3 * time.Second)
			skipIfStillRunningCall++
		}).Cron("*/2 * * * * *").SkipIfStillRunning(),
		app.Command("test --name Goravel argument0 argument1").EveryMinute(),
	})

	go app.Run()

	time.Sleep(60 * time.Second)

	s.NoError(app.Stop())
	s.Equal(1, immediatelyCall)
	s.Equal(30, delayIfStillRunningCall)
	s.Equal(15, skipIfStillRunningCall)
}

func (s *ApplicationTestSuite) TestOnOneServer() {
	mockCache := mockscache.NewCache(s.T())
	mockLock := mockscache.NewLock(s.T())
	mockLock.EXPECT().Get().Return(true).Once()
	mockCache.EXPECT().Lock(mock.Anything, 1*time.Hour).Return(mockLock).Once()

	mockCache1 := mockscache.NewCache(s.T())
	mockLock1 := mockscache.NewLock(s.T())
	mockLock1.EXPECT().Get().Return(false).Once()
	mockCache1.EXPECT().Lock(mock.Anything, 1*time.Hour).Return(mockLock1).Once()

	immediatelyCall := 0

	app := NewApplication(nil, mockCache, nil, false)
	app.Register([]schedule.Event{
		app.Call(func() {
			immediatelyCall++
		}).Cron("*/2 * * * * *").OnOneServer().Name("immediately"),
	})

	app1 := NewApplication(nil, mockCache1, nil, false)
	app1.Register([]schedule.Event{
		app1.Call(func() {
			immediatelyCall++
		}).Cron("*/2 * * * * *").OnOneServer().Name("immediately"),
	})

	go app.Run()
	go app1.Run()

	time.Sleep(2 * time.Second)

	s.NoError(app.Stop())
	s.NoError(app1.Stop())

	s.Equal(1, immediatelyCall)
}

func (s *ApplicationTestSuite) TestStop() {
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

	s.NoError(app.Stop())
	s.Equal(1, immediatelyCall)
}

func (s *ApplicationTestSuite) TestStopWithContext() {
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	s.EqualError(app.Stop(ctx), "context deadline exceeded")
	s.Equal(0, immediatelyCall)
}
