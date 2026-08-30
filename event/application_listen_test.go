package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksqueue "github.com/goravel/framework/mocks/queue"
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

func (r *recordingListener) Handle(evt any, args ...any) error {
	if r.panics {
		panic("listener panicked")
	}

	r.events = append(r.events, evt)
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
			name: "QueueJobIsRegisteredOncePerSignature",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				mockQueue.EXPECT().Register(mock.Anything).Once()

				if err := app.Listen("user.created", &recordingListener{signature: "recording"}); err != nil {
					return err
				}

				return app.Listen("user.updated", &recordingListener{signature: "recording"})
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.listeners["user.created"], 1)
				assert.Len(t, app.listeners["user.updated"], 1)
			},
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
			name: "TypedClosureWithExplicitEvent",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen("user.created", func(evt *userCreated) error { return nil })
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.listeners["user.created"], 1)
			},
		},
		{
			name: "Wildcard",
			setup: func(app *Application, mockQueue *mocksqueue.Queue) error {
				return app.Listen("user.*", func(evt any, args ...any) error { return nil })
			},
			assert: func(t *testing.T, app *Application) {
				assert.Len(t, app.wildcards["user.*"], 1)
				assert.True(t, app.wildcards["user.*"][0].wildcard)
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

	assert.ErrorContains(t, err, "invalid listener for event user.created")
	assert.ErrorContains(t, err, "invalid listener for event user.updated")
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
