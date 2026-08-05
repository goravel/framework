package broadcasters

import (
	"context"
)

type NullDriver struct{}

func NewNullDriver() *NullDriver {
	return &NullDriver{}
}

func (d *NullDriver) Broadcast(_ context.Context, _ []string, _ string, _ map[string]any) error {
	return nil
}
