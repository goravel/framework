package notification

import (
	"reflect"
	"slices"
	"sync"
	"sync/atomic"

	contractsnotification "github.com/goravel/framework/contracts/notification"
)

var _ contractsnotification.Fake = (*Fake)(nil)

// Fake records the notifications sent while faking. The Notification
// binding is a singleton, so the fake belongs to the Manager and every send
// through the facade sees it.
type Fake struct {
	mu      sync.RWMutex
	enabled atomic.Bool
	sent    []*sent
}

func NewFake() *Fake {
	return &Fake{}
}

func (r *Fake) AssertCount(count int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.sent) == count
}

func (r *Fake) AssertNothingSent() bool {
	return r.AssertCount(0)
}

func (r *Fake) AssertNothingSentTo(notifiable contractsnotification.Notifiable) bool {
	return !slices.ContainsFunc(r.records(), func(record *sent) bool {
		return sameNotifiable(record.notifiable, notifiable)
	})
}

func (r *Fake) AssertNotSentTo(notifiable contractsnotification.Notifiable, notification contractsnotification.Notification, assertion ...func(sent contractsnotification.Sent) bool) bool {
	return !r.AssertSentTo(notifiable, notification, assertion...)
}

func (r *Fake) AssertSentOnDemand(notification contractsnotification.Notification, assertion ...func(sent contractsnotification.Sent) bool) bool {
	onDemand := func(sent contractsnotification.Sent) bool {
		_, ok := sent.Notifiable().(contractsnotification.OnDemandNotifiable)
		return ok
	}

	return len(r.Sent(notification, append([]func(sent contractsnotification.Sent) bool{onDemand}, assertion...)...)) > 0
}

func (r *Fake) AssertSentTimes(notification contractsnotification.Notification, times int) bool {
	return len(r.Sent(notification)) == times
}

func (r *Fake) AssertSentTo(notifiable contractsnotification.Notifiable, notification contractsnotification.Notification, assertion ...func(sent contractsnotification.Sent) bool) bool {
	to := func(sent contractsnotification.Sent) bool {
		return sameNotifiable(sent.Notifiable(), notifiable)
	}

	return len(r.Sent(notification, append([]func(sent contractsnotification.Sent) bool{to}, assertion...)...)) > 0
}

func (r *Fake) Sent(notification contractsnotification.Notification, filter ...func(sent contractsnotification.Sent) bool) []contractsnotification.Sent {
	notificationType := elemType(reflect.TypeOf(notification))
	if notificationType == nil {
		return nil
	}

	// Filters run without the lock, so they may call the fake again. The
	// records are never changed once kept, so they can be handed out as they
	// are.
	var result []contractsnotification.Sent
	for _, record := range r.records() {
		if elemType(reflect.TypeOf(record.notification)) != notificationType {
			continue
		}

		if matchesAll(record, filter) {
			result = append(result, record)
		}
	}

	return result
}

func (r *Fake) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled.Store(false)
	r.sent = nil
}

// records returns a snapshot of the records.
func (r *Fake) records() []*sent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return slices.Clone(r.sent)
}

// enable starts faking. It is idempotent: while faking, the records are kept.
func (r *Fake) enable() {
	r.enabled.Store(true)
}

// record keeps the notification with the channels it would have been sent
// through, like Laravel's fake: the ones ShouldSend rejects are left out,
// and nothing is recorded when no channel is left. It reports whether the
// fake took the notification, in which case it must not be delivered. It
// decides and records under one lock, so a send racing Reset is either
// recorded before the records are cleared or delivered.
func (r *Fake) record(notifiable contractsnotification.Notifiable, notification contractsnotification.Notification, queued bool) bool {
	if r == nil || !r.enabled.Load() {
		return false
	}

	// Via and ShouldSend are user code, so they run without the lock.
	channels := slices.Clone(notification.Via(notifiable))
	if shouldSend, ok := notification.(contractsnotification.NotificationWithShouldSend); ok {
		channels = slices.DeleteFunc(channels, func(channel string) bool {
			return !shouldSend.ShouldSend(notifiable, channel)
		})
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.enabled.Load() {
		return false
	}
	if len(channels) > 0 {
		r.sent = append(r.sent, &sent{
			notifiable:   notifiable,
			notification: notification,
			channels:     channels,
			queued:       queued,
		})
	}

	return true
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

var _ contractsnotification.Sent = (*sent)(nil)

// sent is a record of the fake. It is never changed once kept, and Channels
// hands out a copy, so a test can not change it.
type sent struct {
	notifiable   contractsnotification.Notifiable
	notification contractsnotification.Notification
	channels     []string
	queued       bool
}

func (r *sent) Channels() []string {
	return slices.Clone(r.channels)
}

func (r *sent) HasChannel(channel string) bool {
	return slices.Contains(r.channels, channel)
}

func (r *sent) Notifiable() contractsnotification.Notifiable {
	return r.notifiable
}

func (r *sent) Notification() contractsnotification.Notification {
	return r.notification
}

func (r *sent) Queued() bool {
	return r.queued
}
