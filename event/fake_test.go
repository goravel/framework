package event

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/event"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

func TestFake_RecordsInsteadOfCallingListeners(t *testing.T) {
	app := newFakeTestApp(t)
	listener := &recordingListener{signature: "listener"}
	assert.NoError(t, app.Listen(&userCreated{}, listener))

	fake := app.Fake()
	evt := &userCreated{}
	result := app.Dispatch(evt, []event.Arg{{Type: "string", Value: "goravel"}})

	assert.False(t, result.Failed())
	assert.Empty(t, listener.args)
	assert.True(t, fake.AssertDispatched(&userCreated{}))
	assert.True(t, fake.AssertDispatched((*userCreated)(nil), func(dispatched event.Dispatched) bool {
		return dispatched.Event() == evt && dispatched.HasArg("goravel")
	}))
	assert.False(t, fake.AssertDispatched(&userCreated{}, func(dispatched event.Dispatched) bool {
		return dispatched.HasArg("other")
	}))
	assert.True(t, fake.AssertNotDispatched(&userUpdated{}))
	assert.False(t, fake.AssertNotDispatched(&userCreated{}))
	assert.True(t, fake.AssertDispatchedTimes(&userCreated{}, 1))
	assert.False(t, fake.AssertDispatchedTimes(&userCreated{}, 0))
	assert.False(t, fake.AssertDispatchedTimes(&userCreated{}, 2))
	assert.False(t, fake.AssertNothingDispatched())

	dispatched := fake.Dispatched(&userCreated{})
	assert.Len(t, dispatched, 1)
	assert.Equal(t, typeName(reflect.TypeOf(evt)), dispatched[0].Name())
	assert.Equal(t, []event.Arg{{Type: "string", Value: "goravel"}}, dispatched[0].Args())
	assert.Equal(t, "goravel", dispatched[0].Arg(0))
	assert.Nil(t, dispatched[0].Arg(1))
}

func TestFake_DoesNotRunEventHandle(t *testing.T) {
	app := newFakeTestApp(t)
	assert.NoError(t, app.Listen(&panickingEvent{}, &recordingListener{signature: "listener"}))

	fake := app.Fake()

	assert.False(t, app.Dispatch(&panickingEvent{}).Failed())
	assert.True(t, fake.AssertDispatched(&panickingEvent{}))
}

func TestFake_RecordsEventsWithoutListeners(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()

	app.Dispatch("app.started")

	assert.True(t, fake.AssertDispatched("app.started"))
}

func TestFake_OnlyTheGivenEvents(t *testing.T) {
	app := newFakeTestApp(t)
	created := &recordingListener{signature: "created"}
	updated := &recordingListener{signature: "updated"}
	assert.NoError(t, app.Listen(&userCreated{}, created))
	assert.NoError(t, app.Listen(&userUpdated{}, updated))
	assert.NoError(t, app.Listen("order.shipped", &recordingListener{signature: "shipped"}))

	fake := app.Fake(&userCreated{}, "order.*")

	app.Dispatch(&userCreated{})
	app.Dispatch(&userUpdated{})
	app.Dispatch("order.shipped")

	assert.Empty(t, created.args)
	assert.Len(t, updated.args, 1)
	assert.True(t, fake.AssertDispatched(&userCreated{}))
	assert.True(t, fake.AssertDispatched("order.shipped"))
	// Events that are not faked reach their listeners and are not recorded.
	assert.True(t, fake.AssertNotDispatched(&userUpdated{}))
}

func TestFake_InvalidEventFakesNothing(t *testing.T) {
	app := newFakeTestApp(t)
	listener := &recordingListener{signature: "listener"}
	assert.NoError(t, app.Listen(&userCreated{}, listener))

	fake := app.Fake(123)
	app.Dispatch(&userCreated{})

	assert.Len(t, listener.args, 1)
	assert.True(t, fake.AssertNothingDispatched())
}

func TestFake_WildcardLookup(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()

	app.Dispatch("user.created")
	app.Dispatch("user.deleted")
	app.Dispatch("order.shipped")

	assert.Len(t, fake.Dispatched("user.*"), 2)
	assert.True(t, fake.AssertDispatchedTimes("*", 3))
	assert.Empty(t, fake.Dispatched(123))
	assert.False(t, fake.AssertDispatched(123))
}

func TestFake_CallingFakeAgainKeepsTheRecords(t *testing.T) {
	app := newFakeTestApp(t)
	listener := &recordingListener{signature: "listener"}
	assert.NoError(t, app.Listen(&userCreated{}, listener))

	fake := app.Fake()
	app.Dispatch(&userCreated{})

	// The faked events are replaced, so userCreated reaches its listener again.
	assert.Same(t, fake, app.Fake(&userUpdated{}))
	app.Dispatch(&userCreated{})

	assert.Len(t, listener.args, 1)
	assert.True(t, fake.AssertDispatchedTimes(&userCreated{}, 1))
}

func TestFake_Reset(t *testing.T) {
	app := newFakeTestApp(t)
	listener := &recordingListener{signature: "listener"}
	assert.NoError(t, app.Listen(&userCreated{}, listener))

	fake := app.Fake()
	app.Dispatch(&userCreated{})
	fake.Reset()

	assert.True(t, fake.AssertNothingDispatched())

	app.Dispatch(&userCreated{})

	assert.Len(t, listener.args, 1)
	assert.True(t, fake.AssertNothingDispatched())
}

func TestFake_Job(t *testing.T) {
	app := newFakeTestApp(t)
	listener := &recordingListener{signature: "listener"}
	assert.NoError(t, app.Listen(&userCreated{}, listener))

	task := app.Job(&userCreated{}, []event.Arg{{Type: "string", Value: "goravel"}})

	// Whether the event is faked is decided when the task is dispatched.
	fake := app.Fake()
	assert.NoError(t, task.Dispatch())

	assert.Empty(t, listener.args)
	assert.True(t, fake.AssertDispatched(&userCreated{}, func(dispatched event.Dispatched) bool {
		return dispatched.HasArg("goravel")
	}))

	fake.Reset()
	assert.NoError(t, task.Dispatch())
	assert.Len(t, listener.args, 1)
}

func TestFake_NewTaskIsNotFaked(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()
	listener := &recordingListener{signature: "listener"}

	// A task built by hand belongs to no Application, so no fake sees it.
	assert.NoError(t, NewTask(nil, nil, &userCreated{}, []event.Listener{listener}).Dispatch())

	assert.Len(t, listener.args, 1)
	assert.True(t, fake.AssertNothingDispatched())
}

func TestFake_AssertListening(t *testing.T) {
	app := newFakeTestApp(t)
	listener := &recordingListener{signature: "created"}
	assert.NoError(t, app.Listen(&userCreated{}, listener))
	assert.NoError(t, app.Listen("user.*", &recordingListener{signature: "wildcard"}))
	assert.NoError(t, app.Listen("user.deleted", func(evt any, args ...any) error { return nil }))

	fake := app.Fake()

	assert.True(t, fake.AssertListening(&userCreated{}, listener))
	assert.True(t, fake.AssertListening((*userCreated)(nil), "created"))
	assert.True(t, fake.AssertListening("user.deleted", "wildcard"))
	assert.False(t, fake.AssertListening(&userUpdated{}, listener))
	assert.False(t, fake.AssertListening("user.deleted", ""))
	assert.False(t, fake.AssertListening("user.deleted", func(evt any, args ...any) error { return nil }))
	assert.False(t, fake.AssertListening(123, listener))
	assert.False(t, fake.AssertListening(&userCreated{}, (*recordingListener)(nil)))
	assert.False(t, fake.AssertListening(&userCreated{}, nil))
}

func TestFake_RecordsAreCopies(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()

	args := []event.Arg{{Type: "string", Value: "goravel"}}
	app.Dispatch("app.started", args)
	args[0].Value = "changed"

	dispatched := fake.Dispatched("app.started")
	dispatched[0].Args()[0].Value = "changed"

	assert.True(t, fake.AssertDispatched("app.started", func(dispatched event.Dispatched) bool {
		return dispatched.HasArg("goravel")
	}))
}

func TestFake_FiltersMayCallTheFake(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()
	app.Dispatch("app.started")

	assert.True(t, fake.AssertDispatched("app.started", func(event.Dispatched) bool {
		return fake.AssertDispatchedTimes("app.started", 1)
	}))
}

func TestFake_Concurrent(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			app.Dispatch("app.started")
		}()
		go func() {
			defer wg.Done()
			fake.Dispatched("app.started")
		}()
	}
	wg.Wait()

	assert.True(t, fake.AssertDispatchedTimes("app.started", 50))
}

func TestDispatched_HasArgComparesUncomparableValues(t *testing.T) {
	dispatched := &dispatched{args: []event.Arg{{Value: []string{"a"}}}}

	assert.True(t, dispatched.HasArg([]string{"a"}))
	assert.False(t, dispatched.HasArg([]string{"b"}))
	assert.Nil(t, dispatched.Arg(-1))
}

// newFakeTestApp lets the listeners register their queue jobs, which every
// Listen with a Listener does.
func newFakeTestApp(t *testing.T) *Application {
	mockQueue := mocksqueue.NewQueue(t)
	mockQueue.EXPECT().Register(mock.Anything).Maybe()

	return NewApplication(mockQueue)
}
