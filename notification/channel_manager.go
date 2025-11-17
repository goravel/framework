package notification

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/notification/channels"
	"strings"
	"sync"
)

// channelRegistry 管理已注册通道（线程安全）
type channelRegistry struct {
	mu       sync.RWMutex
	channels map[string]notification.Channel
}

var registry = &channelRegistry{
	channels: make(map[string]notification.Channel),
}

// RegisterChannel 允许用户在应用启动时注册自定义通道（注册一次即可）
func RegisterChannel(name string, ch notification.Channel) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.channels[strings.ToLower(name)] = ch
}

// GetChannel 获取已注册通道
func GetChannel(name string) (notification.Channel, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	ch, ok := registry.channels[strings.ToLower(name)]
	return ch, ok
}

// Boot: 注册内置默认通道（mail, database）
func RegisterDefaultChannels(app foundation.Application) {
	RegisterChannel("mail", &channels.MailChannel{})
	RegisterChannel("database", &channels.DatabaseChannel{})
}
