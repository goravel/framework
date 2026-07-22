package broadcasting

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sync"

	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/config"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/facades"
)

type authEntry struct {
	callback broadcasting.ChannelAuthFunc
	regex    *regexp.Regexp
	params   []string
}

type Application struct {
	config      *Config
	log         log.Log
	queue       queue.Queue
	channelAuth map[string]authEntry
	mu          sync.RWMutex
}

func NewApplication(cfg config.Config, log log.Log, queue queue.Queue) *Application {
	return &Application{
		config:      NewConfig(cfg),
		log:         log,
		queue:       queue,
		channelAuth: make(map[string]authEntry),
	}
}

func (a *Application) Channel(pattern string, callback broadcasting.ChannelAuthFunc) {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry := authEntry{
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

	a.channelAuth[pattern] = entry
}

func (a *Application) Dispatch(event broadcasting.ShouldBroadcast) error {
	if !event.BroadcastWhen() {
		return nil
	}

	channels := event.BroadcastOn()
	if len(channels) == 0 {
		return nil
	}

	eventName := event.BroadcastAs()
	if eventName == "" {
		eventName = reflect.TypeOf(event).Elem().Name()
	}

	conn := a.config.DefaultConnection()
	if withConn, ok := event.(broadcasting.ShouldBroadcastWithConnection); ok && withConn.BroadcastConnection() != "" {
		conn = withConn.BroadcastConnection()
	}

	item := broadcastItem{
		Channels:   channelNames(channels),
		Event:      eventName,
		Payload:    event.BroadcastWith(),
		Connection: conn,
	}

	encoded, err := json.Marshal(item)
	if err != nil {
		return err
	}

	job := a.queue.Job(&BroadcastJob{}, []queue.Arg{{Type: "string", Value: string(encoded)}})

	if withQueue, ok := event.(broadcasting.ShouldBroadcastWithQueue); ok && withQueue.BroadcastQueue() != "" {
		job = job.OnQueue(withQueue.BroadcastQueue())
	}

	return job.Dispatch()
}

func (a *Application) Authenticate(ctx contractshttp.Context) contractshttp.Response {
	socketID := ctx.Request().Query("socket_id")
	channelName := ctx.Request().Query("channel_name")

	if socketID == "" || channelName == "" {
		return ctx.Response().Json(http.StatusBadRequest, contractshttp.Json{
			"error": errors.BroadcastAuthMissingParams.Error(),
		})
	}

	auth := facades.Auth(ctx)
	if !auth.Check() {
		return ctx.Response().Json(http.StatusForbidden, contractshttp.Json{"error": errors.BroadcastAuthUnauthenticated.Error()})
	}

	ch := broadcasting.Channel{Name: channelName}
	if !IsPrivateChannel(ch) && !IsPresenceChannel(ch) {
		return ctx.Response().Json(http.StatusOK, broadcasting.AuthResponse{})
	}

	userID, err := auth.ID()
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, contractshttp.Json{"error": err.Error()})
	}
	user := map[string]any{"id": userID}

	if !a.resolveAuth(ChannelBaseName(ch), user) {
		return ctx.Response().Json(http.StatusForbidden, contractshttp.Json{
			"error": errors.BroadcastChannelUnauthorized.Args(channelName).Error(),
		})
	}

	channelData := ctx.Request().Input("channel_data")
	cfg := NewConfig(facades.Config())
	conn, err := cfg.Connection(cfg.DefaultConnection())
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, contractshttp.Json{"error": err.Error()})
	}

	signature := computeAuthSignature(conn.Secret, socketID, channelName, channelData)
	resp := broadcasting.AuthResponse{
		Auth: fmt.Sprintf("%s:%s", conn.Key, signature),
	}
	if IsPresenceChannel(ch) && channelData != "" {
		resp.ChannelData = channelData
	}
	return ctx.Response().Json(http.StatusOK, resp)
}

func (a *Application) resolveAuth(channelName string, user any) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for pattern, entry := range a.channelAuth {
		if entry.regex != nil {
			matches := entry.regex.FindStringSubmatch(channelName)
			if matches == nil {
				continue
			}
			params := make(map[string]string)
			for i, name := range entry.params {
				params[name] = matches[i+1]
			}
			return entry.callback(user, channelName, params)
		}
		if pattern == channelName {
			return entry.callback(user, channelName, nil)
		}
	}
	return false
}

var patternRegex = regexp.MustCompile(`\{(\w+)\}`)

type broadcastItem struct {
	Channels   []string       `json:"channels"`
	Event      string         `json:"event"`
	Payload    map[string]any `json:"payload"`
	Connection string         `json:"connection"`
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
