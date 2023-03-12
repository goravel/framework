package event

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/event"
	eventcontract "github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/queue"
	testingdocker "github.com/goravel/framework/testing/docker"
	"github.com/goravel/framework/testing/mock"
)

var (
	testSyncListener        = 0
	testAsyncListener       = 0
	testCancelListener      = 0
	testCancelAfterListener = 0
)

type EventTestSuite struct {
	suite.Suite
	redisPort int
}

func TestEventTestSuite(t *testing.T) {
	redisPool, redisResource, err := testingdocker.Redis()
	if err != nil {
		log.Fatalf("Get redis error: %s", err)
	}

	facades.Queue = queue.NewApplication()
	facades.Event = NewApplication()

	suite.Run(t, &EventTestSuite{
		redisPort: cast.ToInt(redisResource.GetPort("6379/tcp")),
	})

	if err := redisPool.Purge(redisResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func (s *EventTestSuite) SetupTest() {

}

func (s *EventTestSuite) TestEvent() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	mockConfig.On("GetString", "queue.default").Return("redis").Twice()
	mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Times(3)
	mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	mockConfig.On("GetInt", "database.redis.default.port").Return(s.redisPort).Twice()
	mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()

	facades.Event.Register(map[event.Event][]event.Listener{
		&TestEvent{}: {
			&TestSyncListener{},
			&TestAsyncListener{},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(facades.Queue.Worker(nil).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)

	time.Sleep(3 * time.Second)
	s.Nil(facades.Event.Job(&TestEvent{}, []eventcontract.Arg{
		{Type: "string", Value: "Goravel"},
		{Type: "int", Value: 1},
	}).Dispatch())
	time.Sleep(1 * time.Second)
	s.Equal(1, testSyncListener)
	s.Equal(1, testAsyncListener)

	mockConfig.AssertExpectations(s.T())
}

func (s *EventTestSuite) TestCancelEvent() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "app.name").Return("goravel").Twice()
	mockConfig.On("GetString", "queue.default").Return("redis").Once()
	mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Twice()
	mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Once()
	mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Once()
	mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Once()
	mockConfig.On("GetString", "database.redis.default.password").Return("").Once()
	mockConfig.On("GetInt", "database.redis.default.port").Return(s.redisPort).Once()
	mockConfig.On("GetInt", "database.redis.default.database").Return(0).Once()

	facades.Event.Register(map[event.Event][]event.Listener{
		&TestCancelEvent{}: {
			&TestCancelListener{},
			&TestCancelAfterListener{},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(facades.Queue.Worker(nil).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)

	time.Sleep(3 * time.Second)
	s.EqualError(facades.Event.Job(&TestCancelEvent{}, []eventcontract.Arg{
		{Type: "string", Value: "Goravel"},
		{Type: "int", Value: 1},
	}).Dispatch(), "cancel")
	time.Sleep(1 * time.Second)
	s.Equal(1, testCancelListener)
	s.Equal(0, testCancelAfterListener)

	mockConfig.AssertExpectations(s.T())
}

type TestEvent struct {
}

func (receiver *TestEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type TestCancelEvent struct {
}

func (receiver *TestCancelEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type TestAsyncListener struct {
}

func (receiver *TestAsyncListener) Signature() string {
	return "test_async_listener"
}

func (receiver *TestAsyncListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     true,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestAsyncListener) Handle(args ...any) error {
	testAsyncListener++

	return nil
}

type TestSyncListener struct {
}

func (receiver *TestSyncListener) Signature() string {
	return "test_sync_listener"
}

func (receiver *TestSyncListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestSyncListener) Handle(args ...any) error {
	testSyncListener++

	return nil
}

type TestCancelListener struct {
}

func (receiver *TestCancelListener) Signature() string {
	return "test_cancel_listener"
}

func (receiver *TestCancelListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestCancelListener) Handle(args ...any) error {
	testCancelListener++

	return errors.New("cancel")
}

type TestCancelAfterListener struct {
}

func (receiver *TestCancelAfterListener) Signature() string {
	return "test_cancel_after_listener"
}

func (receiver *TestCancelAfterListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestCancelAfterListener) Handle(args ...any) error {
	testCancelAfterListener++

	return nil
}
