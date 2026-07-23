package broadcasting

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

type Driver interface {
	Broadcast(channels []Channel, event string, payload map[string]any) error
}

type Broadcast interface {
	Channel(pattern string, callback ChannelAuthFunc)
	Connection(connection string) Broadcast
	Dispatch(event ShouldBroadcast) error
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
