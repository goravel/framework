package event

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

func TestApplication_Dispatch(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	var received []any
	assert.NoError(t, app.Listen(&userCreated{}, func(evt any, args ...any) error {
		received = append(received, evt, args)

		return nil
	}))

	evt := &userCreated{}
	result := app.Dispatch(evt, []event.Arg{{Type: "string", Value: "goravel"}})

	assert.False(t, result.Failed())
	assert.NoError(t, result.Error())
	assert.Equal(t, []any{evt, []any{"goravel"}}, received)
}

func TestApplication_DispatchWithoutPayload(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	var received []any
	assert.NoError(t, app.Listen("app.started", func(evt any, args ...any) error {
		received = append(received, evt, args)

		return nil
	}))

	assert.False(t, app.Dispatch("app.started").Failed())
	assert.Equal(t, []any{"app.started", []any{}}, received)
}

func TestApplication_DispatchWildcard(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	var received []any
	assert.NoError(t, app.Listen("user.*", func(evt any, args ...any) error {
		received = append(received, evt)

		return nil
	}))

	assert.False(t, app.Dispatch("user.created").Failed())
	assert.False(t, app.Dispatch("user.updated").Failed())
	// The second dispatch of the same event is served from the wildcard cache.
	assert.False(t, app.Dispatch("user.created").Failed())
	assert.False(t, app.Dispatch("order.placed").Failed())

	// Wildcard listeners receive the name of the event that matched.
	assert.Equal(t, []any{"user.created", "user.updated", "user.created"}, received)
}

func TestApplication_DispatchWildcardCacheIsInvalidated(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	var calls int
	assert.NoError(t, app.Listen("user.*", func(evt any, args ...any) error {
		calls++

		return nil
	}))
	assert.False(t, app.Dispatch("user.created").Failed())

	assert.NoError(t, app.Listen("user.cre*", func(evt any, args ...any) error {
		calls++

		return nil
	}))
	assert.False(t, app.Dispatch("user.created").Failed())

	assert.Equal(t, 3, calls)
}

func TestApplication_DispatchTypedClosure(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	var received *userCreated
	assert.NoError(t, app.Listen(func(evt *userCreated) error {
		received = evt

		return nil
	}))

	evt := &userCreated{}
	assert.False(t, app.Dispatch(evt).Failed())
	assert.Same(t, evt, received)
}

func TestApplication_DispatchTypedClosureWithMismatchedEvent(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	// The closure listens on event.userCreated, but a value is dispatched where
	// the closure expects a pointer.
	assert.NoError(t, app.Listen(func(evt *userCreated) error { return nil }))

	result := app.Dispatch(userCreated{})

	assert.True(t, result.Failed())
	assert.EqualError(t, result.Error(), errors.EventInvalidEvent.Args(userCreated{}).Error())
}

func TestApplication_DispatchWithoutListeners(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	result := app.Dispatch("user.created")

	assert.False(t, result.Failed())
	assert.Nil(t, result.Errors())
	assert.NoError(t, result.Error())
}

func TestApplication_DispatchInvalidEvent(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	result := app.Dispatch(nil)

	assert.True(t, result.Failed())
	assert.EqualError(t, result.Error(), errors.EventInvalidEvent.Args(nil).Error())
}

func TestApplication_DispatchCollectsEveryListenerError(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	var calls int
	assert.NoError(t, app.Listen("user.created",
		func(evt any, args ...any) error { calls++; return errors.New("first") },
		func(evt any, args ...any) error { calls++; return nil },
		func(evt any, args ...any) error { calls++; return errors.New("second") },
	))

	result := app.Dispatch("user.created")

	assert.Equal(t, 3, calls)
	assert.True(t, result.Failed())
	assert.Len(t, result.Errors(), 2)
	assert.ErrorContains(t, result.Error(), "first")
	assert.ErrorContains(t, result.Error(), "second")
}

func TestApplication_DispatchRecoversFromPanic(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	mockQueue.EXPECT().Register(mock.Anything).Once()

	app := NewApplication(mockQueue)

	var called bool
	assert.NoError(t, app.Listen("user.created",
		&recordingListener{signature: "panicking", panics: true},
		func(evt any, args ...any) error { called = true; return nil },
	))

	result := app.Dispatch("user.created")

	// The panicking listener fails on its own, the next one still runs.
	assert.True(t, called)
	assert.True(t, result.Failed())
	assert.EqualError(t, result.Error(), errors.EventListenerPanic.Args("user.created", "listener panicked").Error())
}

func TestApplication_DispatchQueuedListener(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	mockPendingJob := mocksqueue.NewPendingJob(t)
	listener := &recordingListener{
		signature: "queued",
		options:   event.Queue{Enable: true, Connection: "redis", Queue: "events"},
	}

	mockQueue.EXPECT().Register(mock.Anything).Once()
	// The event name travels as the first queue argument, only scalars survive
	// the queue boundary.
	mockQueue.EXPECT().Job(mock.Anything, []queue.Arg{
		{Type: "string", Value: "user.created"},
		{Type: "string", Value: "goravel"},
	}).Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().OnConnection("redis").Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().OnQueue("events").Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().Dispatch().Return(nil).Once()

	app := NewApplication(mockQueue)
	assert.NoError(t, app.Listen("user.created", listener))

	result := app.Dispatch("user.created", []event.Arg{{Type: "string", Value: "goravel"}})

	assert.False(t, result.Failed())
	// The listener was queued, not called in process.
	assert.Nil(t, listener.events)
}

func TestApplication_DispatchQueuedListenerError(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	mockPendingJob := mocksqueue.NewPendingJob(t)

	mockQueue.EXPECT().Register(mock.Anything).Once()
	mockQueue.EXPECT().Job(mock.Anything, []queue.Arg{
		{Type: "string", Value: "user.created"},
	}).Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().Dispatch().Return(errors.New("queue error")).Once()

	app := NewApplication(mockQueue)
	assert.NoError(t, app.Listen("user.created", &recordingListener{
		signature: "queued",
		options:   event.Queue{Enable: true},
	}))

	assert.EqualError(t, app.Dispatch("user.created").Error(), "queue error")
}

func TestApplication_DispatchSyncListenerSkipsTheQueue(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	mockQueue.EXPECT().Register(mock.Anything).Once()

	listener := &recordingListener{signature: "sync"}
	app := NewApplication(mockQueue)
	assert.NoError(t, app.Listen("user.created", listener))

	assert.False(t, app.Dispatch("user.created", []event.Arg{{Type: "string", Value: "goravel"}}).Failed())
	assert.Equal(t, []any{"user.created"}, listener.events)
	assert.Equal(t, [][]any{{"goravel"}}, listener.args)
}

func TestApplication_DispatchListenerRegisteredThroughRegister(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	mockPendingJob := mocksqueue.NewPendingJob(t)
	listener := &TestListener{}

	mockQueue.EXPECT().Register([]queue.Job{listener}).Once()
	// Deprecated listeners keep going through the queue facade, synchronously
	// because their Queue().Enable is false.
	mockQueue.EXPECT().Job(listener, []queue.Arg{{Type: "string", Value: "goravel"}}).Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().DispatchSync().Return(nil).Once()

	app := NewApplication(mockQueue)
	app.Register(map[event.Event][]event.Listener{
		&userCreated{}: {listener},
	})

	assert.False(t, app.Dispatch(&userCreated{}, []event.Arg{{Type: "string", Value: "goravel"}}).Failed())
}

func TestApplication_DispatchConcurrently(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()

			assert.NoError(t, app.Listen("user.*", func(evt any, args ...any) error { return nil }))
		}()
		go func() {
			defer wg.Done()

			app.Dispatch("user.created")
		}()
	}

	wg.Wait()
}
