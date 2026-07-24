package broadcasting

import (
	"encoding/json"

	"github.com/goravel/framework/contracts/broadcasting"
	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

// BroadcastJob handles dispatching broadcasts asynchronously.
// All broadcast events share one job signature because the job data is
// self-contained in args (serialized to JSON), avoiding the need for
// separate job types per event.
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

	for _, conn := range conns {
		cfgConn, err := cfg.Connection(conn)
		if err != nil {
			return err
		}

		driver, err := CreateDriver(cfgConn, j.app)
		if err != nil {
			return err
		}

		if err := driver.Broadcast(channels, item.Event, item.Payload); err != nil {
			return err
		}
	}

	return nil
}
