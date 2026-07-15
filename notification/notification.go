// Package notification provides the Manager implementation for Goravel's
// notification module: dispatching to registered channels synchronously
// or via the queue. See job.go for how queued dispatch stays safe across
// a real queue round-trip.
package notification

import (
	"sync"

	"github.com/goravel/framework/contracts/log"
	contractsnotification "github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

// Manager is the concrete implementation of contracts/notification.Manager.
// It is instantiated once by the ServiceProvider and bound into Goravel's
// service container under binding.Notification.
type Manager struct {
	mu       sync.RWMutex
	channels map[string]contractsnotification.Channel
	log      log.Log
	queue    queue.Queue // may be nil when queue is not configured
}

// NewManager constructs a Manager. The queue argument may be nil;
// notifications that implement ShouldQueue will fall back to synchronous
// delivery in that case, same as the original package.
func NewManager(logger log.Log, q queue.Queue) *Manager {
	return &Manager{
		channels: make(map[string]contractsnotification.Channel),
		log:      logger,
		queue:    q,
	}
}

func (m *Manager) Extend(ch contractsnotification.Channel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[ch.Name()] = ch
}

func (m *Manager) Channel(name string) contractsnotification.Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ch, ok := m.channels[name]
	if !ok {
		m.log.Errorf("%s", errors.NotificationChannelNotFound.Args(name).Error())
		return nil
	}
	return ch
}

// Route begins an on-demand notification: send to a raw address without
// a backing Notifiable model.
func (m *Manager) Route(channel string, route any) contractsnotification.OnDemandNotifiable {
	return &onDemandNotifiable{
		manager: m,
		routes:  map[string]any{channel: route},
	}
}

func (m *Manager) Send(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	if sq, ok := n.(contractsnotification.ShouldQueue); ok && m.queue != nil {
		return m.dispatchQueued(notifiable, n, sq)
	}
	return m.dispatchSync(notifiable, n)
}

func (m *Manager) SendNow(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	return m.dispatchSync(notifiable, n)
}

// dispatchSync iterates over Via() channels and calls each driver's Send.
// Errors from individual channels are logged and joined via errors.Join
// (not discarded after the first) so a caller can inspect every failure,
// not just whichever channel happened to fail first — while still
// attempting every channel regardless of earlier failures.
func (m *Manager) dispatchSync(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	viaChannels := n.Via(notifiable)
	if len(viaChannels) == 0 {
		m.log.Errorf("notifications: %T.Via() returned no channels for %T — nothing sent", n, notifiable)
		return nil
	}

	shouldSend, _ := n.(contractsnotification.NotificationWithShouldSend)
	afterSending, _ := n.(contractsnotification.NotificationWithAfterSending)

	var errs []error
	for _, name := range viaChannels {
		if shouldSend != nil && !shouldSend.ShouldSend(notifiable, name) {
			continue
		}

		ch := m.Channel(name)
		if ch == nil {
			errs = append(errs, errors.NotificationChannelNotFound.Args(name))
			continue
		}

		if err := ch.Send(notifiable, n); err != nil {
			m.log.Errorf("notifications: channel %q failed for %T: %v", name, n, err)
			errs = append(errs, err)
			continue
		}

		if afterSending != nil {
			if err := afterSending.AfterSending(notifiable, name); err != nil {
				m.log.Errorf("notifications: AfterSending hook failed for channel %q, %T: %v", name, n, err)
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

// dispatchQueued resolves each channel's payload eagerly (while notifiable
// and n are still live values on this goroutine) and queues the resolved,
// plain data — not the original job design. See job.go for the full
// rationale; short version: the original sendNotificationJob held
// unexported live interface fields that don't survive real queue drivers.
//
// NotificationWithAfterSending is NOT called on this path: by the time
// DispatchJob.Handle actually delivers (possibly in a different process),
// the live notification n no longer exists — only the resolved
// (channel, route, payload) does. Calling AfterSending here, before
// delivery has actually happened, would be a lie. This is a real scope
// limitation, not an oversight — flagged for discussion rather than
// worked around with something that would silently misbehave.
func (m *Manager) dispatchQueued(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
	sq contractsnotification.ShouldQueue,
) error {
	shouldSend, _ := n.(contractsnotification.NotificationWithShouldSend)

	var errs []error
	for _, name := range n.Via(notifiable) {
		if shouldSend != nil && !shouldSend.ShouldSend(notifiable, name) {
			continue
		}

		ch := m.Channel(name)
		if ch == nil {
			errs = append(errs, errors.NotificationChannelNotFound.Args(name))
			continue
		}

		resolvable, ok := ch.(contractsnotification.ResolvableChannel)
		if !ok {
			// Degrade clearly rather than silently drop data — a custom
			// channel that only implements Channel (not
			// ResolvableChannel) can't be queued safely.
			err := errors.NotificationChannelNotQueueable.Args(name)
			m.log.Errorf("notifications: %v", err)
			errs = append(errs, err)
			continue
		}

		route, payload, err := resolvable.Resolve(notifiable, n)
		if err != nil {
			m.log.Errorf("notifications: failed to resolve channel %q for %T: %v", name, n, err)
			errs = append(errs, err)
			continue
		}

		item := dispatchItem{Channel: name, Route: route, Payload: payload}
		encoded, err := encodeDispatchItem(item)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		pending := m.queue.Job(NewDispatchJob(m), []queue.Arg{{Type: "string", Value: encoded}})
		if conn := sq.OnConnection(); conn != "" {
			pending = pending.OnConnection(conn)
		}
		if q := sq.OnQueue(); q != "" {
			pending = pending.OnQueue(q)
		}
		if err := pending.Dispatch(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
