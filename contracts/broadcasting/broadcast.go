package broadcasting

import "github.com/goravel/framework/contracts/http"

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

type ChannelAuthFunc func(user any, channelName string, params map[string]string) bool

type Broadcast interface {
	Channel(pattern string, callback ChannelAuthFunc)
	Dispatch(event ShouldBroadcast) error
	Authenticate(ctx http.Context) http.Response
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

type ShouldBroadcastWithConnection interface {
	BroadcastConnection() string
}
