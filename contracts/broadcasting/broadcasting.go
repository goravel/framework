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

type ShouldBroadcastWithTries interface {
	// BroadcastTries returns the maximum number of attempts for the
	// queued broadcast. 0 / not implementing the interface means the
	// broadcast is single-shot.
	BroadcastTries() int
}

type ShouldBroadcastWithBackoff interface {
	// BroadcastBackoff returns the delay before each retry attempt, in
	// order; the last value repeats for subsequent attempts. It only takes
	// effect together with BroadcastTries; without it the broadcast is
	// single-shot.
	BroadcastBackoff() []time.Duration
}
