package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/support/str"
)

type userCreated struct{}

func (r *userCreated) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type userUpdated struct{}

func (r *userUpdated) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

// recordingListener is an event.QueueListener that records what it received.
type recordingListener struct {
	signature string
	options   event.Queue
	err       error
	panics    bool
	events    []any
	args      [][]any
}

func (r *recordingListener) Signature() string {
	return r.signature
}

func (r *recordingListener) Queue(args ...any) event.Queue {
	return r.options
}

func (r *recordingListener) Handle(eventName string, args ...any) error {
	if r.panics {
		panic("listener panicked")
	}

	r.events = append(r.events, eventName)
	r.args = append(r.args, args)

	return r.err
}

func TestApplication_Listen(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(app *Application, mockQueue *mocksqueue.Queue) error
		expectedErr error
		assert      func(t *testing.T, app *Application)
	}{
		{
			name: "StringEventWithQueueListener",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				mockQueue.EXPECT().Register(mock.Anything).Once()

				return app.Listen("user.created", &recordingListener{signature: "recording"})
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.listeners["user.created"], 1)
			},
		},
		{
			name: "QueueJobIsRegisteredOncePerListener",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				mockQueue.EXPECT().Register(mock.Anything).Once()
				listener := &recordingListener{signature: "recording"}

				return app.Listen([]string{"user.created", "user.updated"}, listener)
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.listeners["user.created"], 1)
				assert.Len(t, app.listeners["user.updated"], 1)
			},
		},
		{
			name: "DuplicateSignatureIsRejected",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				mockQueue.EXPECT().Register(mock.Anything).Once()

				// The queue resolves a job by its signature alone, so a second
				// listener claiming it would run the first listener's code.
				if err := app.Listen("user.created", &recordingListener{signature: "recording"}); err != nil {
					return err
				}

				return app.Listen("user.updated", &recordingListener{signature: "recording"})
			},
			expectedErr: errors.EventQueueDuplicateSignature.Args("recording"),
		},
		{
			name: "MultipleStringEventsAndMultipleListeners",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen([]string{"user.created", "user.updated"},
					func(evt any, args ...any) error { return nil },
					func(evt any, args ...any) error { return nil },
				)
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.listeners["user.created"], 2)
				assert.Len(t, app.listeners["user.updated"], 2)
			},
		},
		{
			name: "EventInterface",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen(&userCreated{}, func(evt any, args ...any) error { return nil })
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.listeners["event.userCreated"], 1)
			},
		},
		{
			name: "EventInterfaceSlice",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen([]event.Event{&userCreated{}, &userUpdated{}},
					func(evt any, args ...any) error { return nil })
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.listeners["event.userCreated"], 1)
				assert.Len(t, app.listeners["event.userUpdated"], 1)
			},
		},
		{
			name: "AnySlice",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen([]any{"user.created", &userUpdated{}},
					func(evt any, args ...any) error { return nil })
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.listeners["user.created"], 1)
				assert.Len(t, app.listeners["event.userUpdated"], 1)
			},
		},
		{
			name: "TypedClosureInfersTheEvent",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen(func(evt *userCreated) error { return nil })
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.listeners["event.userCreated"], 1)
			},
		},
		{
			name: "TypedClosureWithMatchingExplicitEvent",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen(&userCreated{}, func(evt *userCreated) error { return nil })
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.listeners["event.userCreated"], 1)
			},
		},
		{
			name: "TypedClosureOnAnotherEvent",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen("user.created", func(evt *userCreated) error { return nil })
			},
			expectedErr: errors.EventListenerEventMismatch.Args("event.userCreated", "user.created"),
		},
		{
			name: "TypedClosureOnWildcard",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen("user.*", func(evt *userCreated) error { return nil })
			},
			expectedErr: errors.EventListenerEventMismatch.Args("event.userCreated", "user.*"),
		},
		{
			name: "Wildcard",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen("user.*", func(evt any, args ...any) error { return nil })
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.wildcards, 1)
				assert.Equal(t, "user.*", app.wildcards[0].pattern)
				assert.Len(t, app.wildcards[0].listeners, 1)
				assert.True(t, app.wildcards[0].listeners[0].wildcard)
				assert.Empty(t, app.listeners["user.*"])
			},
		},
		{
			name: "EmptyStringEvent",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen("", func(evt any, args ...any) error { return nil })
			},
			expectedErr: errors.EventInvalidEvent.Args(""),
		},
		{
			name: "NilEvent",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen(nil, func(evt any, args ...any) error { return nil })
			},
			expectedErr: errors.EventInvalidEvent.Args(nil),
		},
		{
			name: "InvalidListener",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen("user.created", "not a listener")
			},
			expectedErr: errors.EventInvalidListener.Args("user.created"),
		},
		{
			name: "LegacyListenerIsRejected",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen("user.created", &TestListener{})
			},
			expectedErr: errors.EventInvalidListener.Args("user.created"),
		},
		{
			name: "ClosureWithTooManyParameters",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen("user.created", func(evt *userCreated, extra string) error { return nil })
			},
			expectedErr: errors.EventInvalidListener.Args("user.created"),
		},
		{
			name: "BareClosureWithoutErrorReturn",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen(func(evt *userCreated) {})
			},
			expectedErr: errors.EventInvalidListener.Args("func(*event.userCreated)"),
		},
		{
			name: "BareNonClosure",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen("user.created")
			},
			expectedErr: errors.EventInvalidListener.Args("string"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockQueue := mocksqueue.NewQueue(t)
			app := NewApplication(mockQueue)

			err := test.setup(app, mockQueue)

			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			test.assert(t, app)
		})
	}
}

func TestApplication_ListenCollectsEveryError(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	err := app.Listen([]string{"user.created", "user.updated"}, "not a listener")

	// The whole request is rejected before anything is registered.
	assert.EqualError(t, err, errors.EventInvalidListener.Args("user.created, user.updated").Error())
	assert.Empty(t, app.listeners)
}

func TestQueueJob(t *testing.T) {
	listener := &recordingListener{signature: "recording"}
	job := &queueJob{listener: listener}

	assert.Equal(t, "recording", job.Signature())
	assert.EqualError(t, job.Handle(), errors.EventQueueMissingEvent.Args("recording").Error())

	assert.NoError(t, job.Handle("user.created", "first", "second"))
	assert.Equal(t, []any{"user.created"}, listener.events)
	assert.Equal(t, [][]any{{"first", "second"}}, listener.args)
}

func TestGetEventName(t *testing.T) {
	tests := []struct {
		name         string
		event        any
		expectedName string
		expectedErr  error
	}{
		{name: "String", event: "user.created", expectedName: "user.created"},
		{name: "Pointer", event: &userCreated{}, expectedName: "event.userCreated"},
		{name: "Value", event: userCreated{}, expectedName: "event.userCreated"},
		{name: "EmptyString", event: "", expectedErr: errors.EventInvalidEvent.Args("")},
		{name: "Nil", event: nil, expectedErr: errors.EventInvalidEvent.Args(nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			name, err := getEventName(test.event)

			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expectedName, name)
		})
	}
}

var _ queue.Job = (*queueJob)(nil)

type transformingEvent struct{}

func (r *transformingEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	transformed := make([]event.Arg, 0, len(args))
	for _, arg := range args {
		transformed = append(transformed, event.Arg{Type: arg.Type, Value: arg.Value.(string) + "!"})
	}

	return transformed, nil
}

// matchWildcard must agree with the str helper the framework already ships,
// which is what the previous implementation used.
func TestMatchWildcard(t *testing.T) {
	cases := []struct{ pattern, name string }{
		{"user.*", "user.created"},
		{"user.*", "user."},
		{"user.*", "user"},
		{"user.*", "order.created"},
		{"*.created", "user.created"},
		{"*", "anything"},
		{"*", ""},
		{"", ""},
		{"", "x"},
		{"user.created", "user.created"},
		{"user.created", "user.updated"},
		{"user.*.created", "user.admin.created"},
		{"user.*.created", "user.created"},
		{"**", "user.created"},
		{"a*b*c", "axxbyyc"},
		{"a*b*c", "abc"},
		{"a*b*c", "acb"},
		{"user.(created)", "user.(created)"},
		{"user.[a-z]+", "user.abc"},
		{"用户.*", "用户.创建"},
		{"用户.*", "订单.创建"},
	}

	for _, c := range cases {
		t.Run(c.pattern+"|"+c.name, func(t *testing.T) {
			assert.Equal(t, str.Of(c.name).Is(c.pattern), matchWildcard(c.pattern, c.name))
		})
	}
}

type valueListener struct{}

func (r valueListener) Signature() string             { return "value" }
func (r valueListener) Queue(args ...any) event.Queue { return event.Queue{} }
func (r valueListener) Handle(string, ...any) error   { return nil }

func TestApplication_ListenRejectsAValueListener(t *testing.T) {
	app := NewApplication(mocksqueue.NewQueue(t))

	// Two values of the same type share an identity, which would let one be
	// queued and the other executed under the same signature.
	err := app.Listen("user.created", valueListener{})

	assert.EqualError(t, err, errors.EventListenerNotPointer.Args("value").Error())
	assert.Empty(t, app.listeners)
}
