package broadcasters

import (
	"context"

	"github.com/goravel/framework/contracts/broadcasting"
)

type NullDriver struct{}

func NewNullDriver() *NullDriver {
	return &NullDriver{}
}

func (d *NullDriver) Broadcast(_ context.Context, _ []broadcasting.Channel, _ string, _ map[string]any) error {
	return nil
}
