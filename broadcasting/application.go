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

	"github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/broadcasting"
	contractsconfig "github.com/goravel/framework/contracts/config"
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
	config      *Config
	auth        auth.Auth
	log         log.Log
	queue       queue.Queue
	channelAuth []authEntry
	mu          sync.RWMutex
	defaultConn string
}

func NewApplication(cfg contractsconfig.Config, auth auth.Auth, log log.Log, queue queue.Queue) *Application {
	bc, err := NewConfig(cfg)
	if err != nil {
		bc = &Config{Default: "log"}
	}

	return &Application{
		config:      bc,
		auth:        auth,
		log:         log,
		queue:       queue,
		channelAuth: make([]authEntry, 0),
	}
}

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

func (a *Application) Connection(connection string) broadcasting.Broadcast {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.defaultConn = connection
	return a
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
		eventName = reflect.Indirect(reflect.ValueOf(event)).Type().Name()
	}

	conn := a.config.Default
	if a.defaultConn != "" {
		conn = a.defaultConn
	}
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

	if withConn, ok := event.(broadcasting.ShouldBroadcastWithQueueConnection); ok && withConn.BroadcastQueueConnection() != "" {
		job = job.OnConnection(withConn.BroadcastQueueConnection())
	}

	if withQueue, ok := event.(broadcasting.ShouldBroadcastWithQueue); ok && withQueue.BroadcastQueue() != "" {
		job = job.OnQueue(withQueue.BroadcastQueue())
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

	userID, err := a.auth.ID()
	if err != nil {
		return ctx.Response().Json(http.StatusUnauthorized, contractshttp.Json{
			"error": errors.BroadcastAuthUnauthenticated.Error(),
		})
	}
	user := map[string]any{"id": userID}

	if !a.resolveAuth(ChannelBaseName(ch), user) {
		return ctx.Response().Json(http.StatusForbidden, contractshttp.Json{
			"error": errors.BroadcastChannelUnauthorized.Args(channelName).Error(),
		})
	}

	conn, err := a.config.Connection(a.config.DefaultConnection())
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, contractshttp.Json{"error": err.Error()})
	}

	channelData := ""
	if IsPresenceChannel(ch) {
		channelData = a.buildPresenceChannelData(userID)
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

func (a *Application) buildPresenceChannelData(userID string) string {
	channelData, _ := json.Marshal(map[string]any{
		"user_id":   userID,
		"user_info": map[string]any{"id": userID},
	})
	return string(channelData)
}

func (a *Application) resolveAuth(channelName string, user any) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, entry := range a.channelAuth {
		if entry.regex == nil && entry.pattern == channelName {
			return entry.callback(user, channelName, nil)
		}
	}

	for _, entry := range a.channelAuth {
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
