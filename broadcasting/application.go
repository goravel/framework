package broadcasting

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"

	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
)

type Application struct {
	config      *Config
	log         log.Log
	queue       queue.Queue
	channelAuth map[string]broadcasting.ChannelAuthFunc
}

func NewApplication(cfg config.Config, log log.Log, queue queue.Queue) *Application {
	return &Application{
		config:      NewConfig(cfg),
		log:         log,
		queue:       queue,
		channelAuth: make(map[string]broadcasting.ChannelAuthFunc),
	}
}

func (a *Application) Channel(pattern string, callback broadcasting.ChannelAuthFunc) {
	a.channelAuth[pattern] = callback
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

	job := a.queue.Job(&BroadcastJob{
		Channels: channels,
		Event:    eventName,
		Payload:  event.BroadcastWith(),
	})

	if conn := event.BroadcastConnection(); conn != "" {
		job.OnConnection(conn)
	}
	if q := event.BroadcastQueue(); q != "" {
		job.OnQueue(q)
	}

	return job.Dispatch()
}

func (a *Application) resolveAuth(channelName string, user any) bool {
	for pattern, callback := range a.channelAuth {
		params, matched := matchPattern(pattern, channelName)
		if matched {
			return callback(user, channelName, params)
		}
	}
	return false
}

var patternRegex = regexp.MustCompile(`\{(\w+)\}`)

func matchPattern(pattern, subject string) (map[string]string, bool) {
	paramNames := patternRegex.FindAllStringSubmatch(pattern, -1)
	if len(paramNames) == 0 {
		return nil, pattern == subject
	}

	regexStr := "^" + patternRegex.ReplaceAllString(pattern, `([^.]+)`) + "$"
	re := regexp.MustCompile(regexStr)
	matches := re.FindStringSubmatch(subject)
	if matches == nil {
		return nil, false
	}

	params := make(map[string]string)
	for i, name := range paramNames {
		params[name[1]] = matches[i+1]
	}
	return params, true
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
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
