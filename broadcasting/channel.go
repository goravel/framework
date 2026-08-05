package broadcasting

import (
	"strings"

	"github.com/goravel/framework/contracts/broadcasting"
)

// PublicChannel returns the channel name unchanged. It is kept as a semantic
// wrapper for readability and API stability, mirroring PrivateChannel and
// PresenceChannel (and Laravel's channel constructors): callers write
// broadcasting.PublicChannel("orders") instead of a bare "orders" string.
func PublicChannel(name string) string {
	return name
}

func PrivateChannel(name string) string {
	return broadcasting.ChannelPrefixPrivate + name
}

func PresenceChannel(name string) string {
	return broadcasting.ChannelPrefixPresence + name
}

func IsPrivateChannel(name string) bool {
	return strings.HasPrefix(name, broadcasting.ChannelPrefixPrivate)
}

func IsPresenceChannel(name string) bool {
	return strings.HasPrefix(name, broadcasting.ChannelPrefixPresence)
}

func ChannelBaseName(name string) string {
	if IsPresenceChannel(name) {
		return name[len(broadcasting.ChannelPrefixPresence):]
	}
	if IsPrivateChannel(name) {
		return name[len(broadcasting.ChannelPrefixPrivate):]
	}
	return name
}
