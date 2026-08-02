package broadcasting

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/broadcasting"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksauth "github.com/goravel/framework/mocks/auth"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockshttp "github.com/goravel/framework/mocks/http"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

type mockBroadcastEvent struct {
	broadcastOn              []broadcasting.Channel
	broadcastAs              string
	broadcastWith            map[string]any
	broadcastWhen            bool
	broadcastQueue           string
	broadcastConnections     []string
	broadcastQueueConnection string
	broadcastDelay           time.Time
	broadcastTimeout         time.Duration
	broadcastTries           int
	broadcastBackoff         []time.Duration
}

func (e *mockBroadcastEvent) BroadcastOn() []broadcasting.Channel { return e.broadcastOn }
func (e *mockBroadcastEvent) BroadcastAs() string                 { return e.broadcastAs }
func (e *mockBroadcastEvent) BroadcastWith() map[string]any       { return e.broadcastWith }
func (e *mockBroadcastEvent) BroadcastWhen() bool                 { return e.broadcastWhen }
func (e *mockBroadcastEvent) BroadcastQueue() string              { return e.broadcastQueue }
func (e *mockBroadcastEvent) BroadcastConnections() []string      { return e.broadcastConnections }
func (e *mockBroadcastEvent) BroadcastQueueConnection() string    { return e.broadcastQueueConnection }
func (e *mockBroadcastEvent) BroadcastDelay() time.Time           { return e.broadcastDelay }
func (e *mockBroadcastEvent) BroadcastTimeout() time.Duration     { return e.broadcastTimeout }
func (e *mockBroadcastEvent) BroadcastTries() int                 { return e.broadcastTries }
func (e *mockBroadcastEvent) BroadcastBackoff() []time.Duration   { return e.broadcastBackoff }

type mockBroadcastNowEvent struct {
	*broadcasting.Channel
	broadcastOn          []broadcasting.Channel
	broadcastAs          string
	broadcastWith        map[string]any
	broadcastWhen        bool
	broadcastConnections []string
}

func (e *mockBroadcastNowEvent) BroadcastOn() []broadcasting.Channel { return e.broadcastOn }
func (e *mockBroadcastNowEvent) BroadcastAs() string                 { return e.broadcastAs }
func (e *mockBroadcastNowEvent) BroadcastWith() map[string]any       { return e.broadcastWith }
func (e *mockBroadcastNowEvent) BroadcastWhen() bool                 { return e.broadcastWhen }
func (e *mockBroadcastNowEvent) BroadcastConnections() []string      { return e.broadcastConnections }
func (e *mockBroadcastNowEvent) BroadcastNow() bool                  { return true }

// plainBroadcastEvent implements only broadcasting.ShouldBroadcast, so the
// ShouldBroadcastWithTries/ShouldBroadcastWithBackoff type assertions in
// Dispatch take the ok == false branch.
type plainBroadcastEvent struct {
	broadcastOn   []broadcasting.Channel
	broadcastAs   string
	broadcastWith map[string]any
	broadcastWhen bool
}

func (e *plainBroadcastEvent) BroadcastOn() []broadcasting.Channel { return e.broadcastOn }
func (e *plainBroadcastEvent) BroadcastAs() string                 { return e.broadcastAs }
func (e *plainBroadcastEvent) BroadcastWith() map[string]any       { return e.broadcastWith }
func (e *plainBroadcastEvent) BroadcastWhen() bool                 { return e.broadcastWhen }

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
		return NewApplication(mockConfig, mockLog, mockQueue, nil)
	}

	t.Run("regex wildcard matches parameters", func(t *testing.T) {
		app := newApp()
		app.Channel("orders.{orderId}", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return true, nil
		})

		authorized, _ := app.resolveAuth(context.Background(),"orders.123", nil)
		assert.True(t, authorized)
		authorized, _ = app.resolveAuth(context.Background(),"ordersA123", nil)
		assert.False(t, authorized)
		authorized, _ = app.resolveAuth(context.Background(),"unknown", nil)
		assert.False(t, authorized)
	})

	t.Run("auth callback receives extracted params", func(t *testing.T) {
		app := newApp()
		app.Channel("orders.{orderId}", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return params["orderId"] == "123", nil
		})

		authorized, _ := app.resolveAuth(context.Background(),"orders.123", map[string]any{"id": "1"})
		assert.True(t, authorized)
		authorized, _ = app.resolveAuth(context.Background(),"orders.456", map[string]any{"id": "1"})
		assert.False(t, authorized)
	})

	t.Run("no match returns false", func(t *testing.T) {
		app := newApp()
		app.Channel("orders.{orderId}", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return false, nil
		})

		authorized, _ := app.resolveAuth(context.Background(),"unknown-channel", map[string]any{"id": "1"})
		assert.False(t, authorized)
	})

	t.Run("exact match without params", func(t *testing.T) {
		app := newApp()
		app.Channel("public-channel", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return true, nil
		})

		authorized, _ := app.resolveAuth(context.Background(),"public-channel", nil)
		assert.True(t, authorized)
		authorized, _ = app.resolveAuth(context.Background(),"other-channel", nil)
		assert.False(t, authorized)
	})

	t.Run("exact match takes precedence over regex", func(t *testing.T) {
		app := newApp()
		app.Channel("orders.special", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return true, nil
		})
		app.Channel("orders.{orderId}", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return false, nil
		})

		authorized, _ := app.resolveAuth(context.Background(),"orders.special", nil)
		assert.True(t, authorized)
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

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:   []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:   "test.event",
			broadcastWith: map[string]any{"key": "value"},
			broadcastWhen: true,
		}

		err := app.Dispatch(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with BroadcastWhen false", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:   []broadcasting.Channel{{Name: "test-channel"}},
			broadcastWhen: false,
		}

		err := app.Dispatch(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with no channels", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:   []broadcasting.Channel{},
			broadcastWhen: true,
		}

		err := app.Dispatch(context.Background(), event)
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

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:              []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:              "test.event",
			broadcastWith:            map[string]any{"key": "value"},
			broadcastWhen:            true,
			broadcastConnections:     []string{"pusher"},
			broadcastQueue:           "high-priority",
			broadcastQueueConnection: "redis",
		}

		err := app.Dispatch(context.Background(), event)
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

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:              []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:              "test.event",
			broadcastWith:            map[string]any{"key": "value"},
			broadcastWhen:            true,
			broadcastQueueConnection: "redis",
		}

		err := app.Dispatch(context.Background(), event)
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

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:    []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:    "test.event",
			broadcastWith:  map[string]any{"key": "value"},
			broadcastWhen:  true,
			broadcastDelay: delay,
		}

		err := app.Dispatch(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with timeout", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)
		mockPJ := mocksqueue.NewPendingJob(t)

		mockQ.EXPECT().Job(
			mock.MatchedBy(func(j *BroadcastJob) bool { return j.Signature() == "goravel_broadcast" }),
			mock.MatchedBy(func(args []queue.Arg) bool {
				if len(args) != 1 || args[0].Type != "string" {
					return false
				}
				var item broadcastItem
				if err := json.Unmarshal([]byte(args[0].Value.(string)), &item); err != nil {
					return false
				}
				return item.Timeout == 30000
			}),
		).Return(mockPJ).Once()
		mockPJ.EXPECT().Dispatch().Return(nil).Once()

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:      []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:      "test.event",
			broadcastWith:    map[string]any{"key": "value"},
			broadcastWhen:    true,
			broadcastTimeout: 30 * time.Second,
		}

		err := app.Dispatch(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with tries and backoff", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)
		mockPJ := mocksqueue.NewPendingJob(t)

		mockQ.EXPECT().Job(
			mock.MatchedBy(func(j *BroadcastJob) bool { return j.Signature() == "goravel_broadcast" }),
			mock.MatchedBy(func(args []queue.Arg) bool {
				if len(args) != 1 || args[0].Type != "string" {
					return false
				}
				var item broadcastItem
				if err := json.Unmarshal([]byte(args[0].Value.(string)), &item); err != nil {
					return false
				}
				return item.Tries == 5 && slices.Equal(item.Backoff, []int64{1000, 2000, 5000})
			}),
		).Return(mockPJ).Once()
		mockPJ.EXPECT().Dispatch().Return(nil).Once()

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:      []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:      "test.event",
			broadcastWith:    map[string]any{"key": "value"},
			broadcastWhen:    true,
			broadcastTries:   5,
			broadcastBackoff: []time.Duration{1 * time.Second, 2 * time.Second, 5 * time.Second},
		}

		err := app.Dispatch(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with plain event serializes no retry policy", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)
		mockPJ := mocksqueue.NewPendingJob(t)

		mockQ.EXPECT().Job(
			mock.MatchedBy(func(j *BroadcastJob) bool { return j.Signature() == "goravel_broadcast" }),
			mock.MatchedBy(func(args []queue.Arg) bool {
				if len(args) != 1 || args[0].Type != "string" {
					return false
				}
				var item broadcastItem
				if err := json.Unmarshal([]byte(args[0].Value.(string)), &item); err != nil {
					return false
				}
				return item.Tries == 0 && item.Backoff == nil
			}),
		).Return(mockPJ).Once()
		mockPJ.EXPECT().Dispatch().Return(nil).Once()

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &plainBroadcastEvent{
			broadcastOn:   []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:   "test.event",
			broadcastWith: map[string]any{"key": "value"},
			broadcastWhen: true,
		}

		err := app.Dispatch(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("dispatch with backoff but no tries suppresses backoff", func(t *testing.T) {
		mockCf := setupMockConfig(t, "")
		mockQ := mocksqueue.NewQueue(t)
		mockPJ := mocksqueue.NewPendingJob(t)

		mockQ.EXPECT().Job(
			mock.MatchedBy(func(j *BroadcastJob) bool { return j.Signature() == "goravel_broadcast" }),
			mock.MatchedBy(func(args []queue.Arg) bool {
				if len(args) != 1 || args[0].Type != "string" {
					return false
				}
				var item broadcastItem
				if err := json.Unmarshal([]byte(args[0].Value.(string)), &item); err != nil {
					return false
				}
				return item.Tries == 0 && item.Backoff == nil
			}),
		).Return(mockPJ).Once()
		mockPJ.EXPECT().Dispatch().Return(nil).Once()

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		// Implements ShouldBroadcastWithBackoff with a non-empty backoff but
		// no BroadcastTries: backoff only takes effect with retries, so the
		// serialized payload must not carry it.
		event := &mockBroadcastEvent{
			broadcastOn:      []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:      "test.event",
			broadcastWith:    map[string]any{"key": "value"},
			broadcastWhen:    true,
			broadcastBackoff: []time.Duration{1 * time.Second},
		}

		err := app.Dispatch(context.Background(), event)
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

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastEvent{
			broadcastOn:   []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:   "",
			broadcastWith: map[string]any{"key": "value"},
			broadcastWhen: true,
		}

		err := app.Dispatch(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("dispatch sync with ShouldBroadcastNow", func(t *testing.T) {
		mockCf := setupMockConfig(t, "null")
		mockQ := mocksqueue.NewQueue(t)

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastNowEvent{
			broadcastOn:          []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:          "test.event",
			broadcastWith:        map[string]any{"key": "value"},
			broadcastWhen:        true,
			broadcastConnections: []string{"null"},
		}

		err := app.Dispatch(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("dispatch sync with multiple connections", func(t *testing.T) {
		mockCf := setupMockConfig(t, "null")
		mockQ := mocksqueue.NewQueue(t)

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastNowEvent{
			broadcastOn:          []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:          "test.event",
			broadcastWith:        map[string]any{"key": "value"},
			broadcastWhen:        true,
			broadcastConnections: []string{"null", "null"},
		}

		err := app.Dispatch(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("dispatch sync with ShouldBroadcastNow uses fallback event name", func(t *testing.T) {
		mockCf := setupMockConfig(t, "null")
		mockQ := mocksqueue.NewQueue(t)

		app := NewApplication(mockCf, mockLog, mockQ, nil)

		event := &mockBroadcastNowEvent{
			broadcastOn:          []broadcasting.Channel{{Name: "test-channel"}},
			broadcastAs:          "",
			broadcastWith:        map[string]any{},
			broadcastWhen:        true,
			broadcastConnections: []string{"null"},
		}

		err := app.Dispatch(context.Background(), event)
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

func TestApplication_Authenticate(t *testing.T) {
	newDefaultConfig := func() *Config {
		return &Config{
			Default: "pusher",
			Connections: map[string]broadcasting.ConnectionConfig{
				"pusher": {Driver: "pusher", Key: "app-key", Secret: "app-secret"},
			},
		}
	}

	setupMocksBase := func(t *testing.T, socketID, channelName string) (*mockshttp.Context, *mockshttp.ContextRequest, *mockshttp.ContextResponse, *mockshttp.AbortableResponse) {
		mockReq := mockshttp.NewContextRequest(t)
		mockReq.EXPECT().Input("socket_id").Return(socketID).Once()
		mockReq.EXPECT().Input("channel_name").Return(channelName).Once()

		mockAbortResp := mockshttp.NewAbortableResponse(t)

		mockCtxResp := mockshttp.NewContextResponse(t)

		mockCtx := mockshttp.NewContext(t)
		mockCtx.EXPECT().Response().Return(mockCtxResp).Once()

		return mockCtx, mockReq, mockCtxResp, mockAbortResp
	}

	setupMocksPublic := func(t *testing.T, socketID, channelName string) (*mockshttp.Context, *mockshttp.ContextRequest, *mockshttp.ContextResponse, *mockshttp.AbortableResponse) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksBase(t, socketID, channelName)
		mockCtx.EXPECT().Request().Return(mockReq).Times(2)
		return mockCtx, mockReq, mockCtxResp, mockAbortResp
	}

	setupMocksAuth := func(t *testing.T, socketID, channelName string) (*mockshttp.Context, *mockshttp.ContextRequest, *mockshttp.ContextResponse, *mockshttp.AbortableResponse) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksBase(t, socketID, channelName)
		mockCtx.EXPECT().Request().Return(mockReq).Times(3)
		mockCtx.EXPECT().Context().Return(context.Background()).Maybe()
		return mockCtx, mockReq, mockCtxResp, mockAbortResp
	}

	t.Run("missing socket_id", func(t *testing.T) {
		mockReq := mockshttp.NewContextRequest(t)
		mockReq.EXPECT().Input("socket_id").Return("").Once()
		mockReq.EXPECT().Input("channel_name").Return("channel.1").Once()

		mockAbortResp := mockshttp.NewAbortableResponse(t)

		mockCtxResp := mockshttp.NewContextResponse(t)
		mockCtxResp.EXPECT().Json(http.StatusBadRequest, mock.MatchedBy(func(v any) bool {
			j, ok := v.(contractshttp.Json)
			return ok && j["error"] == errors.BroadcastAuthMissingParams.Error()
		})).Return(mockAbortResp).Once()

		mockCtx := mockshttp.NewContext(t)
		mockCtx.EXPECT().Request().Return(mockReq).Times(2)
		mockCtx.EXPECT().Response().Return(mockCtxResp).Once()

		app := &Application{config: newDefaultConfig()}
		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("missing channel_name", func(t *testing.T) {
		mockReq := mockshttp.NewContextRequest(t)
		mockReq.EXPECT().Input("socket_id").Return("1234.5678").Once()
		mockReq.EXPECT().Input("channel_name").Return("").Once()

		mockAbortResp := mockshttp.NewAbortableResponse(t)

		mockCtxResp := mockshttp.NewContextResponse(t)
		mockCtxResp.EXPECT().Json(http.StatusBadRequest, mock.MatchedBy(func(v any) bool {
			j, ok := v.(contractshttp.Json)
			return ok && j["error"] == errors.BroadcastAuthMissingParams.Error()
		})).Return(mockAbortResp).Once()

		mockCtx := mockshttp.NewContext(t)
		mockCtx.EXPECT().Request().Return(mockReq).Times(2)
		mockCtx.EXPECT().Response().Return(mockCtxResp).Once()

		app := &Application{config: newDefaultConfig()}
		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("public channel returns empty auth", func(t *testing.T) {
		mockCtx, _, mockCtxResp, mockAbortResp := setupMocksPublic(t, "1234.5678", "public-channel")

		mockCtxResp.EXPECT().Json(http.StatusOK, broadcasting.AuthResponse{}).Return(mockAbortResp).Once()

		app := &Application{config: newDefaultConfig()}
		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("private channel with nil auth returns 401", func(t *testing.T) {
		mockCtx, _, mockCtxResp, mockAbortResp := setupMocksPublic(t, "1234.5678", "private-orders.123")

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(nil).Once()

		mockCtxResp.EXPECT().Json(http.StatusUnauthorized, mock.MatchedBy(func(v any) bool {
			j, ok := v.(contractshttp.Json)
			return ok && j["error"] == errors.BroadcastAuthUnauthenticated.Error()
		})).Return(mockAbortResp).Once()

		app := &Application{app: mockFoundApp, config: newDefaultConfig()}
		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("private channel with auth ID error returns 401", func(t *testing.T) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksAuth(t, "1234.5678", "private-orders.123")
		mockReq.EXPECT().Header("Authorization", "").Return("").Once()

		mockAuth := mocksauth.NewAuth(t)
		mockAuth.EXPECT().ID().Return("", errors.BroadcastAuthUnauthenticated).Once()

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(mockAuth).Once()

		mockCtxResp.EXPECT().Json(http.StatusUnauthorized, mock.MatchedBy(func(v any) bool {
			j, ok := v.(contractshttp.Json)
			return ok && j["error"] == errors.BroadcastAuthUnauthenticated.Error()
		})).Return(mockAbortResp).Once()

		app := &Application{app: mockFoundApp, config: newDefaultConfig()}
		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("private channel with auth denied returns 403", func(t *testing.T) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksAuth(t, "1234.5678", "private-orders.123")
		mockReq.EXPECT().Header("Authorization", "").Return("").Once()

		mockAuth := mocksauth.NewAuth(t)
		mockAuth.EXPECT().ID().Return("user-1", nil).Once()

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(mockAuth).Once()

		app := &Application{app: mockFoundApp, config: newDefaultConfig()}
		app.Channel("orders.123", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return false, nil
		})

		mockCtxResp.EXPECT().Json(http.StatusForbidden, mock.MatchedBy(func(v any) bool {
			j, ok := v.(contractshttp.Json)
			return ok && j["error"] == errors.BroadcastChannelUnauthorized.Args("private-orders.123").Error()
		})).Return(mockAbortResp).Once()

		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("private channel with auth granted returns signature", func(t *testing.T) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksAuth(t, "1234.5678", "private-orders.123")
		mockReq.EXPECT().Header("Authorization", "").Return("").Once()

		mockAuth := mocksauth.NewAuth(t)
		mockAuth.EXPECT().ID().Return("user-1", nil).Once()

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(mockAuth).Once()

		app := &Application{app: mockFoundApp, config: newDefaultConfig()}
		app.Channel("orders.123", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return true, nil
		})

		mockCtxResp.EXPECT().Json(http.StatusOK, mock.MatchedBy(func(v any) bool {
			resp, ok := v.(broadcasting.AuthResponse)
			if !ok {
				return false
			}
			assert.NotEmpty(t, resp.Auth)
			sig := computeAuthSignature("app-secret", "1234.5678", "private-orders.123", "")
			assert.Equal(t, "app-key:"+sig, resp.Auth)
			assert.Empty(t, resp.ChannelData)
			return true
		})).Return(mockAbortResp).Once()

		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("presence channel with auth granted returns signature and channel_data", func(t *testing.T) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksAuth(t, "1234.5678", "presence-chat")
		mockReq.EXPECT().Header("Authorization", "").Return("").Once()

		mockAuth := mocksauth.NewAuth(t)
		mockAuth.EXPECT().ID().Return("user-1", nil).Once()

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(mockAuth).Once()

		app := &Application{app: mockFoundApp, config: newDefaultConfig()}
		app.Channel("chat", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return true, nil
		})

		mockCtxResp.EXPECT().Json(http.StatusOK, mock.MatchedBy(func(v any) bool {
			resp, ok := v.(broadcasting.AuthResponse)
			if !ok {
				return false
			}
			assert.NotEmpty(t, resp.Auth)
			assert.NotEmpty(t, resp.ChannelData)

			sig := computeAuthSignature("app-secret", "1234.5678", "presence-chat", resp.ChannelData)
			assert.Equal(t, "app-key:"+sig, resp.Auth)
			return true
		})).Return(mockAbortResp).Once()

		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("presence channel with custom user info in channel_data", func(t *testing.T) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksAuth(t, "1234.5678", "presence-chat")
		mockReq.EXPECT().Header("Authorization", "").Return("").Once()

		mockAuth := mocksauth.NewAuth(t)
		mockAuth.EXPECT().ID().Return("user-1", nil).Once()

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(mockAuth).Once()

		app := &Application{app: mockFoundApp, config: newDefaultConfig()}
		app.Channel("chat", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return true, map[string]any{"id": "user-1", "name": "John"}
		})

		mockCtxResp.EXPECT().Json(http.StatusOK, mock.MatchedBy(func(v any) bool {
			resp, ok := v.(broadcasting.AuthResponse)
			if !ok {
				return false
			}
			assert.NotEmpty(t, resp.Auth)
			assert.NotEmpty(t, resp.ChannelData)

			var data map[string]any
			assert.NoError(t, json.Unmarshal([]byte(resp.ChannelData), &data))
			assert.Equal(t, "user-1", data["user_id"])
			userInfo, ok := data["user_info"].(map[string]any)
			assert.True(t, ok)
			assert.Equal(t, "user-1", userInfo["id"])
			assert.Equal(t, "John", userInfo["name"])

			sig := computeAuthSignature("app-secret", "1234.5678", "presence-chat", resp.ChannelData)
			assert.Equal(t, "app-key:"+sig, resp.Auth)
			return true
		})).Return(mockAbortResp).Once()

		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("config connection error returns 500", func(t *testing.T) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksAuth(t, "1234.5678", "private-orders.123")
		mockReq.EXPECT().Header("Authorization", "").Return("").Once()

		mockAuth := mocksauth.NewAuth(t)
		mockAuth.EXPECT().ID().Return("user-1", nil).Once()

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(mockAuth).Once()

		badCfg := &Config{Default: "nonexistent", Connections: map[string]broadcasting.ConnectionConfig{}}
		app := &Application{app: mockFoundApp, config: badCfg}
		app.Channel("orders.123", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return true, nil
		})

		mockCtxResp.EXPECT().Json(http.StatusInternalServerError, mock.MatchedBy(func(v any) bool {
			j, ok := v.(contractshttp.Json)
			return ok && j["error"] != ""
		})).Return(mockAbortResp).Once()

		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("private channel with JWT parse success then ID success returns 200", func(t *testing.T) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksAuth(t, "1234.5678", "private-orders.123")
		mockReq.EXPECT().Header("Authorization", "").Return("Bearer valid.token").Once()

		mockAuth := mocksauth.NewAuth(t)
		mockAuth.EXPECT().Parse("Bearer valid.token").Return(&contractsauth.Payload{Key: "user-1"}, nil).Once()
		mockAuth.EXPECT().ID().Return("user-1", nil).Once()

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(mockAuth).Once()

		app := &Application{app: mockFoundApp, config: newDefaultConfig()}
		app.Channel("orders.123", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return true, nil
		})

		mockCtxResp.EXPECT().Json(http.StatusOK, mock.MatchedBy(func(v any) bool {
			resp, ok := v.(broadcasting.AuthResponse)
			return ok && resp.Auth != "" && resp.ChannelData == ""
		})).Return(mockAbortResp).Once()

		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("private channel with JWT parse failure returns 401", func(t *testing.T) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksAuth(t, "1234.5678", "private-orders.123")
		mockReq.EXPECT().Header("Authorization", "").Return("Bearer bad.token").Once()

		mockAuth := mocksauth.NewAuth(t)
		mockAuth.EXPECT().Parse("Bearer bad.token").Return(nil, errors.AuthInvalidToken).Once()

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(mockAuth).Once()

		mockCtxResp.EXPECT().Json(http.StatusUnauthorized, mock.MatchedBy(func(v any) bool {
			j, ok := v.(contractshttp.Json)
			return ok && j["error"] == errors.BroadcastAuthUnauthenticated.Error()
		})).Return(mockAbortResp).Once()

		app := &Application{app: mockFoundApp, config: newDefaultConfig()}
		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("private channel with session guard Parse error ignored then ID success returns 200", func(t *testing.T) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksAuth(t, "1234.5678", "private-orders.123")
		mockReq.EXPECT().Header("Authorization", "").Return("Bearer anything").Once()

		mockAuth := mocksauth.NewAuth(t)
		mockAuth.EXPECT().Parse("Bearer anything").Return(nil, errors.AuthUnsupportedDriverMethod.Args("session")).Once()
		mockAuth.EXPECT().ID().Return("user-1", nil).Once()

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(mockAuth).Once()

		app := &Application{app: mockFoundApp, config: newDefaultConfig()}
		app.Channel("orders.123", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
			return true, nil
		})

		mockCtxResp.EXPECT().Json(http.StatusOK, mock.MatchedBy(func(v any) bool {
			resp, ok := v.(broadcasting.AuthResponse)
			return ok && resp.Auth != "" && resp.ChannelData == ""
		})).Return(mockAbortResp).Once()

		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})

	t.Run("private channel with no token and JWT guard ID failure returns 401", func(t *testing.T) {
		mockCtx, mockReq, mockCtxResp, mockAbortResp := setupMocksAuth(t, "1234.5678", "private-orders.123")
		mockReq.EXPECT().Header("Authorization", "").Return("").Once()

		mockAuth := mocksauth.NewAuth(t)
		mockAuth.EXPECT().ID().Return("", errors.AuthParseTokenFirst).Once()

		mockFoundApp := mocksfoundation.NewApplication(t)
		mockFoundApp.EXPECT().MakeAuth(mockCtx).Return(mockAuth).Once()

		mockCtxResp.EXPECT().Json(http.StatusUnauthorized, mock.MatchedBy(func(v any) bool {
			j, ok := v.(contractshttp.Json)
			return ok && j["error"] == errors.BroadcastAuthUnauthenticated.Error()
		})).Return(mockAbortResp).Once()

		app := &Application{app: mockFoundApp, config: newDefaultConfig()}
		resp := app.Authenticate(mockCtx)
		assert.NotNil(t, resp)
	})
}
