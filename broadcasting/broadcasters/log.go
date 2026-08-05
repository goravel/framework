package broadcasters

import (
	"context"

	"github.com/goravel/framework/contracts/log"
)

type LogDriver struct {
	log log.Log
}

func NewLogDriver(log log.Log) *LogDriver {
	return &LogDriver{log: log}
}

func (d *LogDriver) Broadcast(ctx context.Context, channels []string, event string, payload map[string]any) error {
	d.log.WithContext(ctx).With(map[string]any{
		"event":    event,
		"channels": channels,
		"payload":  payload,
	}).Info("Broadcasting event")
	return nil
}
