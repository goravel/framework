package broadcasting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/queue"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

type mockBroadcastEvent struct {
	broadcastOn         []broadcasting.Channel
	broadcastAs         string
	broadcastWith       map[string]any
	broadcastWhen       bool
	broadcastQueue      string
	broadcastConnection string
}

func (e *mockBroadcastEvent) BroadcastOn() []broadcasting.Channel { return e.broadcastOn }
func (e *mockBroadcastEvent) BroadcastAs() string                 { return e.broadcastAs }
func (e *mockBroadcastEvent) BroadcastWith() map[string]any       { return e.broadcastWith }
func (e *mockBroadcastEvent) BroadcastWhen() bool                 { return e.broadcastWhen }
func (e *mockBroadcastEvent) BroadcastQueue() string              { return e.broadcastQueue }
func (e *mockBroadcastEvent) BroadcastConnection() string         { return e.broadcastConnection }

func setupMockConfig(t *testing.T) *mocksconfig.Config {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("broadcasting.default", "log").Return("log").Maybe()
	mockConfig.EXPECT().GetBool("broadcasting.auth.enabled", true).Return(false).Maybe()
	mockConfig.EXPECT().GetString("broadcasting.auth.path", "/broadcasting/auth").Return("/broadcasting/auth").Maybe()
	return mockConfig
}

func TestApplication_Channel(t *testing.T) {
	mockConfig := setupMockConfig(t)
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)

	app := NewApplication(mockConfig, mockLog, mockQueue)
	app.Channel("orders.{orderId}", func(user any, channel string, params map[string]string) bool {
		return params["orderId"] == "123"
	})

	assert.True(t, app.resolveAuth("orders.123", map[string]any{"id": "1"}))
	assert.False(t, app.resolveAuth("orders.456", map[string]any{"id": "1"}))
}

func TestApplication_Channel_RegexWildcard(t *testing.T) {
	mockConfig := setupMockConfig(t)
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)

	app := NewApplication(mockConfig, mockLog, mockQueue)
	app.Channel("orders.{orderId}", func(user any, channel string, params map[string]string) bool {
		return true
	})

	assert.Equal(t, true, app.resolveAuth("orders.123", nil))
	assert.Equal(t, false, app.resolveAuth("ordersA123", nil))
	assert.Equal(t, false, app.resolveAuth("unknown", nil))
}

func TestApplication_Channel_NoMatch(t *testing.T) {
	mockConfig := setupMockConfig(t)
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)

	app := NewApplication(mockConfig, mockLog, mockQueue)
	app.Channel("orders.{orderId}", func(user any, channel string, params map[string]string) bool {
		return false
	})

	assert.False(t, app.resolveAuth("unknown-channel", map[string]any{"id": "1"}))
}

func TestApplication_Channel_NoParams(t *testing.T) {
	mockConfig := setupMockConfig(t)
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)

	app := NewApplication(mockConfig, mockLog, mockQueue)
	app.Channel("public-channel", func(user any, channel string, params map[string]string) bool {
		return true
	})

	assert.True(t, app.resolveAuth("public-channel", nil))
	assert.False(t, app.resolveAuth("other-channel", nil))
}

func TestApplication_Dispatch(t *testing.T) {
	mockConfig := setupMockConfig(t)
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockPendingJob := mocksqueue.NewPendingJob(t)

	mockQueue.EXPECT().Job(mock.MatchedBy(func(j *BroadcastJob) bool {
		return j.Signature() == "goravel_broadcast"
	}), mock.MatchedBy(func(args []queue.Arg) bool {
		return len(args) == 1 && args[0].Type == "string"
	})).Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().Dispatch().Return(nil).Once()

	app := NewApplication(mockConfig, mockLog, mockQueue)

	event := &mockBroadcastEvent{
		broadcastOn:   []broadcasting.Channel{{Name: "test-channel"}},
		broadcastAs:   "test.event",
		broadcastWith: map[string]any{"key": "value"},
		broadcastWhen: true,
	}

	err := app.Dispatch(event)
	assert.NoError(t, err)
}

func TestApplication_Dispatch_BroadcastWhen_False(t *testing.T) {
	mockConfig := setupMockConfig(t)
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)

	app := NewApplication(mockConfig, mockLog, mockQueue)

	event := &mockBroadcastEvent{
		broadcastOn:   []broadcasting.Channel{{Name: "test-channel"}},
		broadcastWhen: false,
	}

	err := app.Dispatch(event)
	assert.NoError(t, err)
}

func TestApplication_Dispatch_NoChannels(t *testing.T) {
	mockConfig := setupMockConfig(t)
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)

	app := NewApplication(mockConfig, mockLog, mockQueue)

	event := &mockBroadcastEvent{
		broadcastOn:   []broadcasting.Channel{},
		broadcastWhen: true,
	}

	err := app.Dispatch(event)
	assert.NoError(t, err)
}

func TestApplication_Dispatch_WithQueueConnection(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("broadcasting.default", "log").Return("log").Maybe()
	mockConfig.EXPECT().GetBool("broadcasting.auth.enabled", true).Return(false).Maybe()
	mockConfig.EXPECT().GetString("broadcasting.auth.path", "/broadcasting/auth").Return("/broadcasting/auth").Maybe()
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockPendingJob := mocksqueue.NewPendingJob(t)

	mockQueue.EXPECT().Job(mock.Anything, mock.Anything).Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().OnQueue("high-priority").Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().Dispatch().Return(nil).Once()

	app := NewApplication(mockConfig, mockLog, mockQueue)

	event := &mockBroadcastEvent{
		broadcastOn:         []broadcasting.Channel{{Name: "test-channel"}},
		broadcastAs:         "test.event",
		broadcastWith:       map[string]any{"key": "value"},
		broadcastWhen:       true,
		broadcastConnection: "pusher",
		broadcastQueue:      "high-priority",
	}

	err := app.Dispatch(event)
	assert.NoError(t, err)
}

func TestApplication_Dispatch_EventName_Fallback(t *testing.T) {
	mockConfig := setupMockConfig(t)
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockPendingJob := mocksqueue.NewPendingJob(t)

	mockQueue.EXPECT().Job(mock.Anything, mock.MatchedBy(func(args []queue.Arg) bool {
		return len(args) == 1 && args[0].Type == "string"
	})).Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().Dispatch().Return(nil).Once()

	app := NewApplication(mockConfig, mockLog, mockQueue)

	event := &mockBroadcastEvent{
		broadcastOn:   []broadcasting.Channel{{Name: "test-channel"}},
		broadcastAs:   "",
		broadcastWith: map[string]any{"key": "value"},
		broadcastWhen: true,
	}

	err := app.Dispatch(event)
	assert.NoError(t, err)
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
