package broadcasting

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/broadcasting"
	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

type authEntry struct {
	pattern  string
	callback broadcasting.ChannelAuthFunc
	regex    *regexp.Regexp
	params   []string
}

type Application struct {
	app          foundation.Application
	config       *Config
	configFacade contractsconfig.Config
	log          log.Log
	queue        queue.Queue
	channelAuth  []authEntry
	mu           sync.RWMutex
}

func NewApplication(configFacade contractsconfig.Config, log log.Log, queue queue.Queue, app foundation.Application) *Application {
	cfg, err := NewConfig(configFacade)
	if err != nil {
		cfg = &Config{Default: "log"}
	}

	return &Application{
		app:          app,
		config:       cfg,
		configFacade: configFacade,
		log:          log,
		queue:        queue,
		channelAuth:  make([]authEntry, 0),
	}
}

// Channel registers an authorization callback for a channel pattern.
//
// Example:
//
//	// Simple channel
//	facades.Broadcast().Channel("orders", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
//		return true, nil
//	})
//
//	// Channel with route params
//	facades.Broadcast().Channel("orders.{orderId}", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
//		// params["orderId"] contains the matched value
//		order := findOrder(ctx, params["orderId"])
//		return order.UserID == userID, nil
//	})
//
//	// Presence channel returning custom user info
//	facades.Broadcast().Channel("chat", func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
//		return true, map[string]any{"id": userID, "name": "Alice"}
//	})
func (a *Application) Channel(pattern string, callback broadcasting.ChannelAuthFunc) {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry := authEntry{
		pattern:  pattern,
		callback: callback,
	}

	paramNames := patternRegex.FindAllStringSubmatch(pattern, -1)
	if len(paramNames) > 0 {
		params := make([]string, len(paramNames))
		for i, name := range paramNames {
			params[i] = name[1]
		}
		entry.params = params

		segments := patternRegex.Split(pattern, -1)
		regexStr := "^"
		for i, seg := range segments {
			regexStr += regexp.QuoteMeta(seg)
			if i < len(params) {
				regexStr += `([^.]+)`
			}
		}
		regexStr += "$"
		entry.regex = regexp.MustCompile(regexStr)
	}

	a.channelAuth = append(a.channelAuth, entry)
}

// Dispatch broadcasts an event. In the synchronous path (events implementing
// ShouldBroadcastNow with BroadcastNow()==true, or when no queue is
// available) ctx bounds the driver call. In the asynchronous path ctx is not
// propagated across the queue boundary; the worker synthesizes a new context,
// optionally bounded by ShouldBroadcastWithTimeout.
func (a *Application) Dispatch(ctx context.Context, event broadcasting.ShouldBroadcast) error {
	if !event.BroadcastWhen() {
		return nil
	}

	channels := event.BroadcastOn()
	if len(channels) == 0 {
		return nil
	}

	eventName := event.BroadcastAs()
	if eventName == "" {
		eventName = reflect.Indirect(reflect.ValueOf(event)).Type().Name()
	}

	conns := a.resolveConnections(event)

	item := broadcastItem{
		Channels:    channelNames(channels),
		Event:       eventName,
		Payload:     event.BroadcastWith(),
		Connections: conns,
	}
	if withTimeout, ok := event.(broadcasting.ShouldBroadcastWithTimeout); ok && withTimeout.BroadcastTimeout() > 0 {
		item.Timeout = withTimeout.BroadcastTimeout().Milliseconds()
	}

	if now, ok := event.(broadcasting.ShouldBroadcastNow); ok && now.BroadcastNow() {
		return a.dispatchSync(ctx, item)
	}

	return a.dispatchAsync(ctx, event, item)
}

func (a *Application) resolveConnections(event broadcasting.ShouldBroadcast) []string {
	conns := []string{a.config.DefaultConnection()}
	if withConn, ok := event.(broadcasting.ShouldBroadcastWithConnections); ok && len(withConn.BroadcastConnections()) > 0 {
		conns = withConn.BroadcastConnections()
	}
	return conns
}

// dispatchSync broadcasts immediately. Synchronous broadcasts never retry;
// ctx bounds the driver call (honoured by the Pusher HTTP client via
// WithContext). When item.Timeout > 0 (event implements ShouldBroadcastWithTimeout)
// the deadline is tightened further.
func (a *Application) dispatchSync(ctx context.Context, item broadcastItem) error {
	ctx, cancel := withTimeout(ctx, item.Timeout)
	defer cancel()

	for _, conn := range item.Connections {
		cfg, err := a.config.Connection(conn)
		if err != nil {
			return err
		}

		driver, err := CreateDriver(cfg, a.app)
		if err != nil {
			return err
		}

		channels := make([]broadcasting.Channel, len(item.Channels))
		for i, name := range item.Channels {
			channels[i] = broadcasting.Channel{Name: name}
		}

		if err := driver.Broadcast(ctx, channels, item.Event, item.Payload); err != nil {
			return err
		}
	}
	return nil
}

// dispatchAsync enqueues the broadcast for worker delivery. The provided ctx
// is not propagated: queue jobs are serialized and dispatched later, possibly
// in a different process, so the worker synthesizes a fresh context
// (optionally bounded by ShouldBroadcastWithTimeout).
func (a *Application) dispatchAsync(_ context.Context, event broadcasting.ShouldBroadcast, item broadcastItem) error {
	encoded, err := json.Marshal(item)
	if err != nil {
		return err
	}

	job := a.queue.Job(&BroadcastJob{config: a.configFacade, app: a.app}, []queue.Arg{{Type: "string", Value: string(encoded)}})

	if withConn, ok := event.(broadcasting.ShouldBroadcastWithQueueConnection); ok && withConn.BroadcastQueueConnection() != "" {
		job = job.OnConnection(withConn.BroadcastQueueConnection())
	}

	if withQueue, ok := event.(broadcasting.ShouldBroadcastWithQueue); ok && withQueue.BroadcastQueue() != "" {
		job = job.OnQueue(withQueue.BroadcastQueue())
	}

	if withDelay, ok := event.(broadcasting.ShouldBroadcastWithDelay); ok && !withDelay.BroadcastDelay().IsZero() {
		job = job.Delay(withDelay.BroadcastDelay())
	}

	return job.Dispatch()
}

func (a *Application) Authenticate(ctx contractshttp.Context) contractshttp.Response {
	socketID := ctx.Request().Input("socket_id")
	channelName := ctx.Request().Input("channel_name")

	if socketID == "" || channelName == "" {
		return ctx.Response().Json(http.StatusBadRequest, contractshttp.Json{
			"error": errors.BroadcastAuthMissingParams.Error(),
		})
	}

	ch := broadcasting.Channel{Name: channelName}
	if !IsPrivateChannel(ch) && !IsPresenceChannel(ch) {
		return ctx.Response().Json(http.StatusOK, broadcasting.AuthResponse{})
	}

	if a.app == nil {
		return ctx.Response().Json(http.StatusUnauthorized, contractshttp.Json{
			"error": errors.BroadcastAuthUnauthenticated.Error(),
		})
	}

	auth := a.app.MakeAuth(ctx)
	if auth == nil {
		return ctx.Response().Json(http.StatusUnauthorized, contractshttp.Json{
			"error": errors.BroadcastAuthUnauthenticated.Error(),
		})
	}

	token := ctx.Request().Header("Authorization", "")
	if token != "" {
		if _, parseErr := auth.Parse(token); parseErr != nil {
			if !errors.Is(parseErr, errors.AuthUnsupportedDriverMethod) {
				return ctx.Response().Json(http.StatusUnauthorized, contractshttp.Json{
					"error": errors.BroadcastAuthUnauthenticated.Error(),
				})
			}
		}
	}

	userID, err := auth.ID()
	if err != nil {
		return ctx.Response().Json(http.StatusUnauthorized, contractshttp.Json{
			"error": errors.BroadcastAuthUnauthenticated.Error(),
		})
	}

	authorized, userInfo := a.resolveAuth(ctx, ChannelBaseName(ch), userID)
	if !authorized {
		return ctx.Response().Json(http.StatusForbidden, contractshttp.Json{
			"error": errors.BroadcastChannelUnauthorized.Args(channelName).Error(),
		})
	}

	conn, err := a.config.Connection(a.config.DefaultConnection())
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, contractshttp.Json{"error": err.Error()})
	}

	if conn.Key == "" || conn.Secret == "" {
		return ctx.Response().Json(http.StatusInternalServerError, contractshttp.Json{
			"error": errors.BroadcastAuthConfigMissing.Error(),
		})
	}

	var resp broadcasting.AuthResponse
	if IsPresenceChannel(ch) {
		channelData, err := a.buildPresenceChannelData(userID, userInfo)
		if err != nil {
			return ctx.Response().Json(http.StatusInternalServerError, contractshttp.Json{"error": errors.BroadcastChannelDataMarshalFailed.Error()})
		}
		signature := computeAuthSignature(conn.Secret, socketID, channelName, channelData)
		resp = broadcasting.AuthResponse{
			Auth:        fmt.Sprintf("%s:%s", conn.Key, signature),
			ChannelData: channelData,
		}
	} else {
		signature := computeAuthSignature(conn.Secret, socketID, channelName, "")
		resp = broadcasting.AuthResponse{
			Auth: fmt.Sprintf("%s:%s", conn.Key, signature),
		}
	}
	return ctx.Response().Json(http.StatusOK, resp)
}

func (a *Application) buildPresenceChannelData(userID string, userInfo any) (string, error) {
	if userInfo == nil {
		userInfo = userID
	}
	channelData, err := json.Marshal(map[string]any{
		"user_id":   userID,
		"user_info": userInfo,
	})
	if err != nil {
		return "", err
	}
	return string(channelData), nil
}

func (a *Application) resolveAuth(ctx context.Context, channelName string, userID any) (bool, any) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Exact matches take precedence over regex patterns regardless of
	// registration order, so exact matches are resolved in a first pass.
	for _, entry := range a.channelAuth {
		if entry.regex == nil && entry.pattern == channelName {
			return entry.callback(ctx, userID, channelName, map[string]string{})
		}
	}

	for _, entry := range a.channelAuth {
		if entry.regex == nil {
			continue
		}

		matches := entry.regex.FindStringSubmatch(channelName)
		if matches == nil {
			continue
		}
		params := make(map[string]string)
		for i, name := range entry.params {
			params[name] = matches[i+1]
		}
		return entry.callback(ctx, userID, channelName, params)
	}
	return false, nil
}

var patternRegex = regexp.MustCompile(`\{(\w+)\}`)

type broadcastItem struct {
	Channels    []string       `json:"channels"`
	Event       string         `json:"event"`
	Payload     map[string]any `json:"payload"`
	Connections []string       `json:"connections"`
	Timeout     int64          `json:"timeout,omitempty"`
}

func channelNames(channels []broadcasting.Channel) []string {
	names := make([]string, len(channels))
	for i, ch := range channels {
		names[i] = ch.Name
	}
	return names
}

func computeAuthSignature(secret, socketID, channelName, channelData string) string {
	message := fmt.Sprintf("%s:%s", socketID, channelName)
	if channelData != "" {
		message += ":" + channelData
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// withTimeout returns parent unchanged when timeoutMs <= 0, otherwise it
// derives a child context bounded by timeoutMs. A nil parent falls back to
// context.Background. The caller MUST defer cancel().
func withTimeout(parent context.Context, timeoutMs int64) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if timeoutMs <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, time.Duration(timeoutMs)*time.Millisecond)
}
