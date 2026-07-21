package broadcasting

import "github.com/goravel/framework/contracts/broadcasting"

type Driver interface {
	Broadcast(channels []broadcasting.Channel, event string, payload map[string]any) error
}
