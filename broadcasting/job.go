package broadcasting

import (
	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/facades"
)

type BroadcastJob struct {
	Channels []broadcasting.Channel
	Event    string
	Payload  map[string]any
}

func (j *BroadcastJob) Signature() string {
	return "goravel_broadcast"
}

func (j *BroadcastJob) Handle(args ...any) error {
	cfg := NewConfig(facades.Config())
	conn, err := cfg.Connection(cfg.DefaultConnection())
	if err != nil {
		return err
	}

	driver, err := CreateDriver(conn, facades.App())
	if err != nil {
		return err
	}

	return driver.Broadcast(j.Channels, j.Event, j.Payload)
}
