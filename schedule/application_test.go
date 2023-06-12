package schedule

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/carbon"
	consolemock "github.com/goravel/framework/contracts/console/mocks"
	logmock "github.com/goravel/framework/contracts/log/mocks"
	"github.com/goravel/framework/contracts/schedule"
)

func TestApplication(t *testing.T) {
	mockArtisan := &consolemock.Artisan{}
	mockArtisan.On("Call", "test --name Goravel argument0 argument1").Return().Times(3)

	mockLog := &logmock.Log{}
	mockLog.On("Error", "panic", mock.Anything).Return().Times(3)

	immediatelyCall := 0
	delayIfStillRunningCall := 0
	skipIfStillRunningCall := 0

	app := NewApplication(mockArtisan, mockLog)
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

	assert.Equal(t, 3, immediatelyCall)
	assert.Equal(t, 2, delayIfStillRunningCall)
	assert.Equal(t, 1, skipIfStillRunningCall)
	mockArtisan.AssertExpectations(t)
}
