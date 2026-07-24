package broadcasting

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/queue"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

type mockBroadcastEvent struct {
	broadcastOn               []broadcasting.Channel
	broadcastAs               string
	broadcastWith             map[string]any
	broadcastWhen             bool
	broadcastQueue            string
	broadcastConnections      []string
	broadcastQueueConnection  string
	broadcastDelay            time.Time
}

func (e *mockBroadcastEvent) BroadcastOn() []broadcasting.Channel     { return e.broadcastOn }
func (e *mockBroadcastEvent) BroadcastAs() string                     { return e.broadcastAs }
func (e *mockBroadcastEvent) BroadcastWith() map[string]any           { return e.broadcastWith }
func (e *mockBroadcastEvent) BroadcastWhen() bool                     { return e.broadcastWhen }
func (e *mockBroadcastEvent) BroadcastQueue() string                  { return e.broadcastQueue }
func (e *mockBroadcastEvent) BroadcastConnections() []string          { return e.broadcastConnections }
func (e *mockBroadcastEvent) BroadcastQueueConnection() string        { return e.broadcastQueueConnection }
func (e *mockBroadcastEvent) BroadcastDelay() time.Time               { return e.broadcastDelay }

type mockBroadcastNowEvent struct {
	*broadcasting.Channel
	broadcastOn          []broadcasting.Channel
	broadcastAs          string
	broadcastWith        map[string]any
	broadcastWhen        bool
	broadcastConnections []string
}

func (e *mockBroadcastNowEvent) BroadcastOn() []broadcasting.Channel    { return e.broadcastOn }
func (e *mockBroadcastNowEvent) BroadcastAs() string                    { return e.broadcastAs }
func (e *mockBroadcastNowEvent) BroadcastWith() map[string]any          { return e.broadcastWith }
func (e *mockBroadcastNowEvent) BroadcastWhen() bool                    { return e.broadcastWhen }
func (e *mockBroadcastNowEvent) BroadcastConnections() []string         { return e.broadcastConnections }
func (e *mockBroadcastNowEvent) BroadcastNow() bool                     { return true }

func setupMockConfig(t *testing.T, defaultConn string) *mocksconfig.Config {
	mockConfig := mocksconfig.NewConfig(t)
	if defaultConn == "" {
		defaultConn = "log"
	}
	c := &Config{Default: defaultConn, Connections: map[string]broadcasting.ConnectionConfig{
		"log":  {Driver: "log"},
		"null": {Driver: "null"},
	}}
	mockConfig.EXPECT().UnmarshalKey("broadcasting", mock.Anything).Run(func(key string, rawVal interface{}) {
		cfg := rawVal.(*Config)
		*cfg = *c
	}).Return(nil).Once()
	return mockConfig
}

func TestApplication_Channel(t *testing.T) {
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)

	newApp := func() *Application {
		mockConfig := setupMockConfig(t, "")
		return NewApplication(mockConfig, nil, mockLog, mockQueue, nil)
	}

	t.Run("regex wildcard matches parameters", func(t *testing.T) {
		app := newApp()
		app.Channel("orders.{orderId}", func(user any, channel string, params map[string]string) bool {
			return true
		})

		assert.True(t, app.resolveAuth("orders.123", nil))
		assert.False(t, app.resolveAuth("ordersA123", nil))
		assert.False(t, app.resolveAuth("unknown", nil))
	})

	t.Run("auth callback receives extracted params", func(t *testing.T) {
		app := newApp()
		app.Channel("orders.{orderId}", func(user any, channel string, params map[string]string) bool {
			return params["orderId"] == "123"
		})

		assert.True(t, app.resolveAuth("orders.123", map[string]any{"id": "1"}))
		assert.False(t, app.resolveAuth("orders.456", map[string]any{"id": "1"}))
	})

	t.Run("no match returns false", func(t *testing.T) {
		app := newApp()
		app.Channel("orders.{orderId}", func(user any, channel string, params map[string]string) bool {
			return false
		})

		assert.False(t, app.resolveAuth("unknown-channel", map[string]any{"id": "1"}))
	})

	t.Run("exact match without params", func(t *testing.T) {
		app := newApp()
		app.Channel("public-channel", func(user any, channel string, params map[string]string) bool {
			return true
		})

		assert.True(t, app.resolveAuth("public-channel", nil))
		assert.False(t, app.resolveAuth("other-channel", nil))
	})

	t.Run("exact match takes precedence over regex", func(t *testing.T) {
		app := newApp()
		app.Channel("orders.{orderId}", func(user any, channel string, params map[string]string) bool {
			return false
		})
		app.Channel("orders.special", func(user any, channel string, params map[string]string) bool {
			return true
		})

		assert.True(t, app.resolveAuth("orders.special", nil))
	})
}

func TestApplication_Dispatch(t *testing.T) {
	mockLog := mockslog.NewLog(t)

	t.Run("dispatch with event name", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)
		mockPJ := mocksqueue.NewPendingJob(t)

		mockQ.EXPECT().Job(
			mock.MatchedBy(func(j *BroadcastJob) bool { return j.Signature() == "goravel_broadcast" }),
			mock.MatchedBy(func(args []queue.Arg) bool { return len(args) == 1 && args[0].Type == "string" }),
		).Return(mockPJ).Once()
		mockPJ.EXPECT().Dispatch().Return(nil).Once()

		app := NewApplication(mockCf, nil, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:   []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:   "test.event",
			broadcastWith: map[string]any{"key": "value"},
			broadcastWhen: true,
		}

		err := app.Dispatch(event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with BroadcastWhen false", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)

		app := NewApplication(mockCf, nil, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:   []broadcasting.Channel{{Name: "test-channel"}},
			broadcastWhen: false,
		}

		err := app.Dispatch(event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with no channels", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)

		app := NewApplication(mockCf, nil, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:   []broadcasting.Channel{},
			broadcastWhen: true,
		}

		err := app.Dispatch(event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with queue and connection", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)
		mockPJ := mocksqueue.NewPendingJob(t)

		mockQ.EXPECT().Job(
			mock.MatchedBy(func(j *BroadcastJob) bool { return j.Signature() == "goravel_broadcast" }),
			mock.MatchedBy(func(args []queue.Arg) bool { return len(args) == 1 && args[0].Type == "string" }),
		).Return(mockPJ).Once()
		mockPJ.EXPECT().OnConnection("redis").Return(mockPJ).Once()
		mockPJ.EXPECT().OnQueue("high-priority").Return(mockPJ).Once()
		mockPJ.EXPECT().Dispatch().Return(nil).Once()

		app := NewApplication(mockCf, nil, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:              []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:              "test.event",
			broadcastWith:            map[string]any{"key": "value"},
			broadcastWhen:            true,
			broadcastConnections:     []string{"pusher"},
			broadcastQueue:           "high-priority",
			broadcastQueueConnection: "redis",
		}

		err := app.Dispatch(event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with only queue connection", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)
		mockPJ := mocksqueue.NewPendingJob(t)

		mockQ.EXPECT().Job(
			mock.MatchedBy(func(j *BroadcastJob) bool { return j.Signature() == "goravel_broadcast" }),
			mock.MatchedBy(func(args []queue.Arg) bool { return len(args) == 1 && args[0].Type == "string" }),
		).Return(mockPJ).Once()
		mockPJ.EXPECT().OnConnection("redis").Return(mockPJ).Once()
		mockPJ.EXPECT().Dispatch().Return(nil).Once()

		app := NewApplication(mockCf, nil, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:              []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:              "test.event",
			broadcastWith:            map[string]any{"key": "value"},
			broadcastWhen:            true,
			broadcastQueueConnection: "redis",
		}

		err := app.Dispatch(event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with delay", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)
		mockPJ := mocksqueue.NewPendingJob(t)

		mockQ.EXPECT().Job(
			mock.MatchedBy(func(j *BroadcastJob) bool { return j.Signature() == "goravel_broadcast" }),
			mock.MatchedBy(func(args []queue.Arg) bool { return len(args) == 1 && args[0].Type == "string" }),
		).Return(mockPJ).Once()
		delay := time.Now().Add(10 * time.Second)
		mockPJ.EXPECT().Delay(delay).Return(mockPJ).Once()
		mockPJ.EXPECT().Dispatch().Return(nil).Once()

		app := NewApplication(mockCf, nil, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:    []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:    "test.event",
			broadcastWith:  map[string]any{"key": "value"},
			broadcastWhen:  true,
			broadcastDelay: delay,
		}

		err := app.Dispatch(event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with fallback event name", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)
		mockPJ := mocksqueue.NewPendingJob(t)

		mockQ.EXPECT().Job(
			mock.MatchedBy(func(j *BroadcastJob) bool { return j.Signature() == "goravel_broadcast" }),
			mock.MatchedBy(func(args []queue.Arg) bool { return len(args) == 1 && args[0].Type == "string" }),
		).Return(mockPJ).Once()
		mockPJ.EXPECT().Dispatch().Return(nil).Once()

		app := NewApplication(mockCf, nil, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:   []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:   "",
			broadcastWith: map[string]any{"key": "value"},
			broadcastWhen: true,
		}

		err := app.Dispatch(event)
		assert.NoError(t, err)
	})

	t.Run("dispatch sync with ShouldBroadcastNow", func(t *testing.T) {
		mockCf := setupMockConfig(t, "null")
		mockQ := mocksqueue.NewQueue(t)

		app := NewApplication(mockCf, nil, mockLog, mockQ, nil)

		event := &mockBroadcastNowEvent{
			broadcastOn:          []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:          "test.event",
			broadcastWith:        map[string]any{"key": "value"},
			broadcastWhen:        true,
			broadcastConnections: []string{"null"},
		}

		err := app.Dispatch(event)
		assert.NoError(t, err)
	})

	t.Run("dispatch sync with multiple connections", func(t *testing.T) {
		mockCf := setupMockConfig(t, "null")
		mockQ := mocksqueue.NewQueue(t)

		app := NewApplication(mockCf, nil, mockLog, mockQ, nil)

		event := &mockBroadcastNowEvent{
			broadcastOn:          []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:          "test.event",
			broadcastWith:        map[string]any{"key": "value"},
			broadcastWhen:        true,
			broadcastConnections: []string{"null", "null"},
		}

		err := app.Dispatch(event)
		assert.NoError(t, err)
	})

	t.Run("dispatch sync with ShouldBroadcastNow uses fallback event name", func(t *testing.T) {
		mockCf := setupMockConfig(t, "null")
		mockQ := mocksqueue.NewQueue(t)

		app := NewApplication(mockCf, nil, mockLog, mockQ, nil)

		event := &mockBroadcastNowEvent{
			broadcastOn:          []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:          "",
			broadcastWith:        map[string]any{},
			broadcastWhen:        true,
			broadcastConnections: []string{"null"},
		}

		err := app.Dispatch(event)
		assert.NoError(t, err)
	})
}

func TestComputeAuthSignature(t *testing.T) {
	sig := computeAuthSignature("secret", "1234.5678", "private-orders.123", "")
	assert.NotEmpty(t, sig)

	sig2 := computeAuthSignature("secret", "1234.5678", "presence-chat", `{"user_id":"1","user_info":{"name":"Alice"}}`)
	assert.NotEmpty(t, sig2)
	assert.NotEqual(t, sig, sig2)
}

func TestChannelHelpers(t *testing.T) {
	public := PublicChannel("my-channel")
	assert.Equal(t, "my-channel", public.Name)
	assert.False(t, IsPrivateChannel(public))
	assert.False(t, IsPresenceChannel(public))
	assert.Equal(t, "my-channel", ChannelBaseName(public))

	private := PrivateChannel("orders.123")
	assert.Equal(t, "private-orders.123", private.Name)
	assert.True(t, IsPrivateChannel(private))
	assert.False(t, IsPresenceChannel(private))
	assert.Equal(t, "orders.123", ChannelBaseName(private))

	presence := PresenceChannel("chat")
	assert.Equal(t, "presence-chat", presence.Name)
	assert.False(t, IsPrivateChannel(presence))
	assert.True(t, IsPresenceChannel(presence))
	assert.Equal(t, "chat", ChannelBaseName(presence))
}
