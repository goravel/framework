package broadcasting

const (
	ChannelPrefixPrivate  = "private-"
	ChannelPrefixPresence = "presence-"
)

type Channel struct {
	Name string
}

func PublicChannel(name string) Channel {
	return Channel{Name: name}
}

func PrivateChannel(name string) Channel {
	return Channel{Name: ChannelPrefixPrivate + name}
}

func PresenceChannel(name string) Channel {
	return Channel{Name: ChannelPrefixPresence + name}
}

func (c Channel) String() string {
	return c.Name
}

func (c Channel) IsPrivate() bool {
	return len(c.Name) >= len(ChannelPrefixPrivate) && c.Name[:len(ChannelPrefixPrivate)] == ChannelPrefixPrivate
}

func (c Channel) IsPresence() bool {
	return len(c.Name) >= len(ChannelPrefixPresence) && c.Name[:len(ChannelPrefixPresence)] == ChannelPrefixPresence
}

func (c Channel) BaseName() string {
	if c.IsPresence() {
		return c.Name[len(ChannelPrefixPresence):]
	}
	if c.IsPrivate() {
		return c.Name[len(ChannelPrefixPrivate):]
	}
	return c.Name
}

type AuthResponse struct {
	Auth        string `json:"auth"`
	ChannelData string `json:"channel_data,omitempty"`
}

type ChannelAuthFunc func(user any, channelName string, params map[string]string) bool

type Broadcaster interface {
	Channel(pattern string, callback ChannelAuthFunc)
	Dispatch(event ShouldBroadcast) error
}

type ShouldBroadcast interface {
	BroadcastOn() []Channel
	BroadcastAs() string
	BroadcastWith() map[string]any
	BroadcastWhen() bool
	BroadcastQueue() string
	BroadcastConnection() string
}
