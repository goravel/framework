package broadcasting

import (
	"encoding/json"
	"time"

	"github.com/goravel/framework/contracts/broadcasting"
	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type BroadcastJob struct {
	config contractsconfig.Config
	app    foundation.Application
}

func (j *BroadcastJob) Signature() string {
	return "goravel_broadcast"
}

func (j *BroadcastJob) Handle(args ...any) error {
	if len(args) != 1 {
		return errors.BroadcastInvalidQueuePayload
	}

	raw, ok := args[0].(string)
	if !ok {
		return errors.BroadcastInvalidQueuePayload
	}

	var item broadcastItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return errors.BroadcastInvalidQueuePayload
	}

	cfg, err := NewConfig(j.config)
	if err != nil {
		return err
	}

	channels := make([]broadcasting.Channel, len(item.Channels))
	for i, name := range item.Channels {
		channels[i] = broadcasting.Channel{Name: name}
	}

	conns := item.Connections
	if len(conns) == 0 {
		conns = []string{cfg.DefaultConnection()}
	}

	maxTries := item.Tries
	if maxTries <= 0 {
		maxTries = 1
	}
	backoff := time.Duration(item.Backoff) * time.Millisecond
	timeout := time.Duration(item.Timeout) * time.Millisecond

	start := time.Now()
	var lastErr error

	for attempt := 1; attempt <= maxTries; attempt++ {
		lastErr = j.broadcastToConns(cfg, channels, item.Event, item.Payload, conns)
		if lastErr == nil {
			return nil
		}

		if attempt < maxTries {
			if timeout > 0 && time.Since(start)+backoff >= timeout {
				break
			}

			if backoff > 0 {
				time.Sleep(backoff)
			}
		}
	}

	return lastErr
}

func (j *BroadcastJob) broadcastToConns(cfg *Config, channels []broadcasting.Channel, event string, payload map[string]any, conns []string) error {
	for _, conn := range conns {
		cfgConn, err := cfg.Connection(conn)
		if err != nil {
			return err
		}

		driver, err := CreateDriver(cfgConn, j.app)
		if err != nil {
			return err
		}

		if err := driver.Broadcast(channels, event, payload); err != nil {
			return err
		}
	}

	return nil
}
