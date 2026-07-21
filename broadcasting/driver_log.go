package broadcasting

import (
	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/log"
)

type LogDriver struct {
	log log.Log
}

func NewLogDriver(log log.Log) *LogDriver {
	return &LogDriver{log: log}
}

func (d *LogDriver) Broadcast(channels []broadcasting.Channel, event string, payload map[string]any) error {
	chanNames := make([]string, len(channels))
	for i, ch := range channels {
		chanNames[i] = ch.Name
	}
	d.log.With(map[string]any{
		"event":    event,
		"channels": chanNames,
		"payload":  payload,
	}).Info("Broadcasting event")
	return nil
}
