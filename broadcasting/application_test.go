package broadcasting

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/broadcasting"
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

func (e *mockBroadcastEvent) BroadcastOn() []broadcasting.Channel  { return e.broadcastOn }
func (e *mockBroadcastEvent) BroadcastAs() string                  { return e.broadcastAs }
func (e *mockBroadcastEvent) BroadcastWith() map[string]any        { return e.broadcastWith }
func (e *mockBroadcastEvent) BroadcastWhen() bool                  { return e.broadcastWhen }
func (e *mockBroadcastEvent) BroadcastQueue() string               { return e.broadcastQueue }
func (e *mockBroadcastEvent) BroadcastConnection() string          { return e.broadcastConnection }

func setupMockConfig(t *testing.T) *mocksconfig.Config {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("broadcasting.default", "log").Return("log").Maybe()
	mockConfig.EXPECT().GetBool("broadcasting.auth.enabled", true).Return(false).Maybe()
	mockConfig.EXPECT().GetString("broadcasting.auth.path", "/broadcasting/auth").Return("/broadcasting/auth").Maybe()
	mockConfig.EXPECT().GetStringSlice("broadcasting.auth.middleware", mock.Anything).Return([]string{"web"}).Maybe()
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

func TestApplication_Dispatch(t *testing.T) {
	mockConfig := setupMockConfig(t)
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockPendingJob := mocksqueue.NewPendingJob(t)

	mockQueue.EXPECT().Job(mock.Anything).Return(mockPendingJob).Once()
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
	mockConfig := setupMockConfig(t)
	mockLog := mockslog.NewLog(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockPendingJob := mocksqueue.NewPendingJob(t)

	mockQueue.EXPECT().Job(mock.Anything).Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().OnConnection("redis").Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().OnQueue("high-priority").Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().Dispatch().Return(nil).Once()

	app := NewApplication(mockConfig, mockLog, mockQueue)

	event := &mockBroadcastEvent{
		broadcastOn:         []broadcasting.Channel{{Name: "test-channel"}},
		broadcastAs:         "test.event",
		broadcastWith:       map[string]any{"key": "value"},
		broadcastWhen:       true,
		broadcastConnection: "redis",
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

	mockQueue.EXPECT().Job(mock.MatchedBy(func(job *BroadcastJob) bool {
		return job.Event == "mockBroadcastEvent"
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

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		match   bool
		params  map[string]string
	}{
		{"orders.{orderId}", "orders.123", true, map[string]string{"orderId": "123"}},
		{"orders.{orderId}", "orders.456", true, map[string]string{"orderId": "456"}},
		{"orders.{orderId}", "unknown", false, nil},
		{"chat.{roomId}", "chat.general", true, map[string]string{"roomId": "general"}},
		{"no-params", "no-params", true, nil},
		{"no-params", "other", false, nil},
		{"{a}.{b}", "foo.bar", true, map[string]string{"a": "foo", "b": "bar"}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s-%s", tt.pattern, tt.subject), func(t *testing.T) {
			params, match := matchPattern(tt.pattern, tt.subject)
			assert.Equal(t, tt.match, match)
			if tt.match {
				assert.Equal(t, tt.params, params)
			}
		})
	}
}

func TestComputeAuthSignature(t *testing.T) {
	sig := computeAuthSignature("secret", "1234.5678", "private-orders.123", "")
	assert.NotEmpty(t, sig)

	sig2 := computeAuthSignature("secret", "1234.5678", "presence-chat", `{"user_id":"1","user_info":{"name":"Alice"}}`)
	assert.NotEmpty(t, sig2)
	assert.NotEqual(t, sig, sig2)
}
