package console

type Stubs struct{}

func (r Stubs) Channel() string {
	return `package DummyPackage

import (
	"context"

	"github.com/goravel/framework/contracts/broadcasting"
)

// DummyChannel authenticates a user for a channel.
// Returns (authorized, userInfo).
//   - authorized: whether the user is allowed to access the channel.
//   - userInfo: optional user data for presence channels (nil = fallback to userID).
func DummyChannel(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any) {
	return false, nil
}

var _ broadcasting.ChannelAuthFunc = DummyChannel
`
}
