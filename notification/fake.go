package notification

import (
	"reflect"
	"slices"
	"sync"

	contractsnotification "github.com/goravel/framework/contracts/notification"
)

var _ contractsnotification.Fake = (*Fake)(nil)

// Fake records the notifications sent while faking. The Notification
// binding is a singleton, so the fake belongs to the Manager and every send
// through the facade sees it.
type Fake struct {
	mu      sync.RWMutex
	enabled bool
	sent    []contractsnotification.Sent
}

func NewFake() *Fake {
	return &Fake{}
}

func (r *Fake) AssertCount(count int) bool {
	return len(r.all()) == count
}

func (r *Fake) AssertNothingSent() bool {
	return r.AssertCount(0)
}

func (r *Fake) AssertNothingSentTo(notifiable contractsnotification.Notifiable) bool {
	return !slices.ContainsFunc(r.all(), func(sent contractsnotification.Sent) bool {
		return sameNotifiable(sent.Notifiable, notifiable)
	})
}

func (r *Fake) AssertNotSentTo(notifiable contractsnotification.Notifiable, notification contractsnotification.Notification, assertion ...func(sent contractsnotification.Sent) bool) bool {
	return !r.AssertSentTo(notifiable, notification, assertion...)
}

func (r *Fake) AssertSentOnDemand(notification contractsnotification.Notification, assertion ...func(sent contractsnotification.Sent) bool) bool {
	onDemand := func(sent contractsnotification.Sent) bool {
		_, ok := sent.Notifiable.(contractsnotification.OnDemandNotifiable)
		return ok
	}

	return len(r.Sent(notification, append([]func(sent contractsnotification.Sent) bool{onDemand}, assertion...)...)) > 0
}

func (r *Fake) AssertSentTimes(notification contractsnotification.Notification, times int) bool {
	return len(r.Sent(notification)) == times
}

func (r *Fake) AssertSentTo(notifiable contractsnotification.Notifiable, notification contractsnotification.Notification, assertion ...func(sent contractsnotification.Sent) bool) bool {
	to := func(sent contractsnotification.Sent) bool {
		return sameNotifiable(sent.Notifiable, notifiable)
	}

	return len(r.Sent(notification, append([]func(sent contractsnotification.Sent) bool{to}, assertion...)...)) > 0
}

func (r *Fake) Sent(notification contractsnotification.Notification, filter ...func(sent contractsnotification.Sent) bool) []contractsnotification.Sent {
	notificationType := elemType(reflect.TypeOf(notification))
	if notificationType == nil {
		return nil
	}

	// Filters run without the lock, so they may call the fake again.
	records := r.all()
	result := make([]contractsnotification.Sent, 0, len(records))
	for _, record := range records {
		if elemType(reflect.TypeOf(record.Notification)) != notificationType {
			continue
		}

		if matchesAll(record, filter) {
			result = append(result, record)
		}
	}

	return slices.Clip(result)
}

func (r *Fake) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled = false
	r.sent = nil
}

// all returns copies of every record.
func (r *Fake) all() []contractsnotification.Sent {
	r.mu.RLock()
	records := slices.Clone(r.sent)
	r.mu.RUnlock()

	for i := range records {
		records[i].Channels = slices.Clone(records[i].Channels)
	}

	return records
}

// enable starts faking. It is idempotent: while faking, the records are kept.
func (r *Fake) enable() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled = true
}

func (r *Fake) isEnabled() bool {
	if r == nil {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.enabled
}

// record keeps the notification with the channels it would have been sent
// through, like Laravel's fake: the ones ShouldSend rejects are left out,
// and nothing is recorded when no channel is left.
func (r *Fake) record(notifiable contractsnotification.Notifiable, notification contractsnotification.Notification, queued bool) {
	channels := notification.Via(notifiable)
	if shouldSend, ok := notification.(contractsnotification.NotificationWithShouldSend); ok {
		channels = slices.DeleteFunc(slices.Clone(channels), func(channel string) bool {
			return !shouldSend.ShouldSend(notifiable, channel)
		})
	}
	if len(channels) == 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.sent = append(r.sent, contractsnotification.Sent{
		Notifiable:   notifiable,
		Notification: notification,
		Channels:     slices.Clone(channels),
		Queued:       queued,
	})
}

// sameNotifiable reports whether both notifiables are the same recipient:
// the same type and, for a DatabaseRoutable, the same non-empty route, or
// deeply equal otherwise.
func sameNotifiable(a, b contractsnotification.Notifiable) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	if routableA, ok := a.(contractsnotification.DatabaseRoutable); ok {
		if key := routableA.RouteNotificationForDatabase(); key != "" {
			return key == b.(contractsnotification.DatabaseRoutable).RouteNotificationForDatabase()
		}
	}

	return reflect.DeepEqual(a, b)
}

func elemType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

func matchesAll(sent contractsnotification.Sent, filters []func(sent contractsnotification.Sent) bool) bool {
	for _, filter := range filters {
		if filter != nil && !filter(sent) {
			return false
		}
	}

	return true
}
