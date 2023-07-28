package schedule

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	cachemocks "github.com/goravel/framework/contracts/cache/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	logmocks "github.com/goravel/framework/contracts/log/mocks"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/support/carbon"
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
	mockArtisan := &consolemocks.Artisan{}
	mockArtisan.On("Call", "test --name Goravel argument0 argument1").Return().Times(3)

	mockLog := &logmocks.Log{}
	mockLog.On("Error", "panic", mock.Anything).Return().Times(3)

	immediatelyCall := 0
	delayIfStillRunningCall := 0
	skipIfStillRunningCall := 0

	app := NewApplication(mockArtisan, nil, mockLog)
	app.Register([]schedule.Event{
		app.Call(func() {
			panic(1)
		}).EveryMinute(),
		app.Call(func() {
			immediatelyCall++
		}).EveryMinute(),
		app.Call(func() {
			time.Sleep(61 * time.Second)
			delayIfStillRunningCall++
		}).EveryMinute().DelayIfStillRunning(),
		app.Call(func() {
			time.Sleep(61 * time.Second)
			skipIfStillRunningCall++
		}).EveryMinute().SkipIfStillRunning(),
		app.Command("test --name Goravel argument0 argument1").EveryMinute(),
	})

	second := carbon.Now().Second()
	// Make sure run 3 times
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(120+6+60-second)*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		app.Run()

		for range ctx.Done() {
			return
		}
	}(ctx)

	time.Sleep(time.Duration(120+5+60-second) * time.Second)
	app.cron.Stop()

	s.Equal(3, immediatelyCall)
	s.Equal(2, delayIfStillRunningCall)
	s.Equal(1, skipIfStillRunningCall)
	mockArtisan.AssertExpectations(s.T())
	mockLog.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestOnOneServer() {
	mockArtisan := &consolemocks.Artisan{}
	mockArtisan.On("Call", "test --name Goravel argument0 argument1").Return().Twice()

	now := carbon.Now().AddMinute()
	mockCache := &cachemocks.Cache{}
	mockLock := &cachemocks.Lock{}
	mockCache.On("Lock", "immediately"+now.Format("Hi"), 1*time.Hour).Return(mockLock).Once()
	mockLock.On("Get").Return(true).Once()
	mockLock = &cachemocks.Lock{}
	mockCache.On("Lock", "immediately"+now.AddMinute().Format("Hi"), 1*time.Hour).Return(mockLock).Once()
	mockLock.On("Get").Return(true).Once()
	mockLock = &cachemocks.Lock{}
	mockCache.On("Lock", "test --name Goravel argument0 argument1"+now.Format("Hi"), 1*time.Hour).Return(mockLock).Once()
	mockLock.On("Get").Return(true).Once()
	mockLock = &cachemocks.Lock{}
	mockCache.On("Lock", "test --name Goravel argument0 argument1"+now.AddMinute().Format("Hi"), 1*time.Hour).Return(mockLock).Once()
	mockLock.On("Get").Return(true).Once()

	mockCache1 := &cachemocks.Cache{}
	mockLock1 := &cachemocks.Lock{}
	mockCache1.On("Lock", "immediately"+now.Format("Hi"), 1*time.Hour).Return(mockLock1).Once()
	mockLock1.On("Get").Return(false).Once()
	mockLock1 = &cachemocks.Lock{}
	mockCache1.On("Lock", "immediately"+now.AddMinute().Format("Hi"), 1*time.Hour).Return(mockLock1).Once()
	mockLock1.On("Get").Return(false).Once()
	mockLock1 = &cachemocks.Lock{}
	mockCache1.On("Lock", "test --name Goravel argument0 argument1"+now.Format("Hi"), 1*time.Hour).Return(mockLock1).Once()
	mockLock1.On("Get").Return(false).Once()
	mockLock1 = &cachemocks.Lock{}
	mockCache1.On("Lock", "test --name Goravel argument0 argument1"+now.AddMinute().Format("Hi"), 1*time.Hour).Return(mockLock1).Once()
	mockLock1.On("Get").Return(false).Once()

	immediatelyCall := 0

	app := NewApplication(mockArtisan, mockCache, nil)
	app.Register([]schedule.Event{
		app.Call(func() {
			immediatelyCall++
		}).EveryMinute().OnOneServer().Name("immediately"),
		app.Command("test --name Goravel argument0 argument1").EveryMinute().OnOneServer(),
	})

	app1 := NewApplication(nil, mockCache1, nil)
	app1.Register([]schedule.Event{
		app.Call(func() {
			immediatelyCall++
		}).EveryMinute().OnOneServer().Name("immediately"),
		app.Command("test --name Goravel argument0 argument1").EveryMinute().OnOneServer(),
	})

	second := carbon.Now().Second()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60+2+60-second)*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		app.Run()
		app1.Run()

		for range ctx.Done() {
			return
		}
	}(ctx)

	time.Sleep(time.Duration(60+1+60-second) * time.Second)
	app.cron.Stop()
	app1.cron.Stop()

	s.Equal(2, immediatelyCall)
	mockArtisan.AssertExpectations(s.T())
	mockCache.AssertExpectations(s.T())
}
