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
		m.log.Errorf("notifications: %v", errors.NotificationChannelNotFound.Args(name))
		return nil
	}
	return ch
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
// Unchanged from the original — errors from individual channels are
// logged but do not abort other channels; behavior verified by the
// ported notification_test.go.
func (m *Manager) dispatchSync(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	channels := n.Via(notifiable)
	if len(channels) == 0 {
		m.log.Errorf("notifications: %T.Via() returned no channels for %T — nothing sent", n, notifiable)
		return nil
	}

	var firstErr error
	for _, name := range channels {
		ch := m.Channel(name)
		if ch == nil {
			err := errors.NotificationChannelNotFound.Args(name)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		if err := ch.Send(notifiable, n); err != nil {
			m.log.Errorf("notifications: channel %q failed for %T: %v", name, n, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// dispatchQueued resolves each channel's payload eagerly (while notifiable
// and n are still live values on this goroutine) and queues the resolved,
// plain data — not the original job design. See job.go for the full
// rationale; short version: the original sendNotificationJob held
// unexported live interface fields that don't survive real queue drivers.
func (m *Manager) dispatchQueued(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
	sq contractsnotification.ShouldQueue,
) error {
	var firstErr error

	for _, name := range n.Via(notifiable) {
		ch := m.Channel(name)
		if ch == nil {
			err := errors.NotificationChannelNotFound.Args(name)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		resolvable, ok := ch.(contractsnotification.ResolvableChannel)
		if !ok {
			// Degrade clearly rather than silently drop data — a custom
			// channel that only implements Channel (not
			// ResolvableChannel) can't be queued safely.
			err := errors.NotificationChannelNotQueueable.Args(name)
			m.log.Errorf("notifications: %v", err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		route, payload, err := resolvable.Resolve(notifiable, n)
		if err != nil {
			m.log.Errorf("notifications: failed to resolve channel %q for %T: %v", name, n, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		item := dispatchItem{Channel: name, Route: route, Payload: payload}
		encoded, err := encodeDispatchItem(item)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		pending := m.queue.Job(NewDispatchJob(m), []queue.Arg{{Type: "string", Value: encoded}})
		if conn := sq.OnConnection(); conn != "" {
			pending = pending.OnConnection(conn)
		}
		if q := sq.OnQueue(); q != "" {
			pending = pending.OnQueue(q)
		}
		if err := pending.Dispatch(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
