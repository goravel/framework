package broadcasting

import (
	"context"
	"encoding/json"

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

// Handle executes the queued broadcast. It is single-shot: retries are
// governed by the queue worker's Args.Tries, not by this job. A fresh context
// is synthesized (the dispatch-time ctx cannot cross the queue boundary); if
// the originating event implemented ShouldBroadcastWithTimeout the worker
// context is bounded accordingly, which the Pusher driver honours via
// WithContext.
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

	ctx, cancel := withTimeout(context.Background(), item.Timeout)
	defer cancel()

	return j.broadcastToConns(ctx, cfg, channels, item.Event, item.Payload, conns)
}

func (j *BroadcastJob) broadcastToConns(ctx context.Context, cfg *Config, channels []broadcasting.Channel, event string, payload map[string]any, conns []string) error {
	for _, conn := range conns {
		cfgConn, err := cfg.Connection(conn)
		if err != nil {
			return err
		}

		driver, err := CreateDriver(cfgConn, j.app)
		if err != nil {
			return err
		}

		if err := driver.Broadcast(ctx, channels, event, payload); err != nil {
			return err
		}
	}

	return nil
}
