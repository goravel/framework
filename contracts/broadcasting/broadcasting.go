package broadcasting

import (
	"context"
	"time"
)

const (
	ChannelPrefixPrivate  = "private-"
	ChannelPrefixPresence = "presence-"
)

type Channel struct {
	Name string
}

type AuthResponse struct {
	Auth        string `json:"auth"`
	ChannelData string `json:"channel_data,omitempty"`
}

type ChannelAuthFunc func(ctx context.Context, userID any, channelName string, params map[string]string) (bool, any)

type Driver interface {
	Broadcast(ctx context.Context, channels []Channel, event string, payload map[string]any) error
}

type Broadcast interface {
	Channel(pattern string, callback ChannelAuthFunc)
	Dispatch(ctx context.Context, event ShouldBroadcast) error
}

type ShouldBroadcast interface {
	BroadcastOn() []Channel
	BroadcastAs() string
	BroadcastWith() map[string]any
	BroadcastWhen() bool
}

type ShouldBroadcastWithQueue interface {
	BroadcastQueue() string
}

type ShouldBroadcastWithConnections interface {
	BroadcastConnections() []string
}

type ShouldBroadcastNow interface {
	BroadcastNow() bool
}

type ShouldBroadcastWithQueueConnection interface {
	BroadcastQueueConnection() string
}

type ShouldBroadcastWithDelay interface {
	BroadcastDelay() time.Time
}

type ShouldBroadcastWithTimeout interface {
	BroadcastTimeout() time.Duration
}
