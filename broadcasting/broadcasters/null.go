package broadcasters

import "github.com/goravel/framework/contracts/broadcasting"

type NullDriver struct{}

func NewNullDriver() *NullDriver {
	return &NullDriver{}
}

func (d *NullDriver) Broadcast(channels []broadcasting.Channel, event string, payload map[string]any) error {
	return nil
}
