package notification

import (
    "github.com/goravel/framework/contracts/foundation"
    "github.com/goravel/framework/contracts/notification"
    "github.com/goravel/framework/notification/channels"
    "strings"
    "sync"
)

// channelRegistry manages registered channels in a thread-safe manner.
type channelRegistry struct {
    mu       sync.RWMutex
    channels map[string]notification.Channel
}

var registry = &channelRegistry{
    channels: make(map[string]notification.Channel),
}

// RegisterChannel allows registering a custom channel, typically during application boot.
// Channel names are normalized to lowercase.
func RegisterChannel(name string, ch notification.Channel) {
    registry.mu.Lock()
    defer registry.mu.Unlock()
    registry.channels[strings.ToLower(name)] = ch
}

// GetChannel returns a previously registered channel by name.
func GetChannel(name string) (notification.Channel, bool) {
    registry.mu.RLock()
    defer registry.mu.RUnlock()
    ch, ok := registry.channels[strings.ToLower(name)]
    return ch, ok
}

// RegisterDefaultChannels registers built-in default channels: mail and database.
func RegisterDefaultChannels(app foundation.Application) {
    RegisterChannel("mail", &channels.MailChannel{})
    RegisterChannel("database", &channels.DatabaseChannel{})
}
