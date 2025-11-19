package notification

import (
	contractsnotification "github.com/goravel/framework/contracts/notification"
	"sync"
)

var (
	notifMu       sync.RWMutex
	notifRegistry = map[string]func() contractsnotification.Notif{}

	notifiableMu       sync.RWMutex
	notifiableRegistry = map[string]func(map[string]interface{}) contractsnotification.Notifiable{}
)

func RegisterNotificationType(name string, factory func() contractsnotification.Notif) {
	notifMu.Lock()
	defer notifMu.Unlock()
	notifRegistry[name] = factory
}

func GetNotificationInstance(name string) (contractsnotification.Notif, bool) {
	notifMu.RLock()
	defer notifMu.RUnlock()
	if f, ok := notifRegistry[name]; ok {
		return f(), true
	}
	return nil, false
}

// RegisterNotifiableType keeps backward compatibility for tests/users
func RegisterNotifiableType(name string, factory func(map[string]interface{}) contractsnotification.Notifiable) {
	notifiableMu.Lock()
	defer notifiableMu.Unlock()
	notifiableRegistry[name] = factory
}

func GetNotifiableInstance(name string, routes map[string]any) (contractsnotification.Notifiable, bool) {
	notifiableMu.RLock()
	defer notifiableMu.RUnlock()
	if f, ok := notifiableRegistry[name]; ok {
		return f(routes), true
	}
	return nil, false
}

func NotifiableHasWithRoutes(name string) bool {
	notifiableMu.RLock()
	defer notifiableMu.RUnlock()
	_, ok := notifiableRegistry[name]
	return ok
}
