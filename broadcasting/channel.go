package broadcasting

import (
	"strings"

	"github.com/goravel/framework/contracts/broadcasting"
)

func PublicChannel(name string) broadcasting.Channel {
	return broadcasting.Channel{Name: name}
}

func PrivateChannel(name string) broadcasting.Channel {
	return broadcasting.Channel{Name: broadcasting.ChannelPrefixPrivate + name}
}

func PresenceChannel(name string) broadcasting.Channel {
	return broadcasting.Channel{Name: broadcasting.ChannelPrefixPresence + name}
}

func IsPrivateChannel(c broadcasting.Channel) bool {
	return strings.HasPrefix(c.Name, broadcasting.ChannelPrefixPrivate)
}

func IsPresenceChannel(c broadcasting.Channel) bool {
	return strings.HasPrefix(c.Name, broadcasting.ChannelPrefixPresence)
}

func ChannelBaseName(c broadcasting.Channel) string {
	if IsPresenceChannel(c) {
		return c.Name[len(broadcasting.ChannelPrefixPresence):]
	}
	if IsPrivateChannel(c) {
		return c.Name[len(broadcasting.ChannelPrefixPrivate):]
	}
	return c.Name
}
