package notification

import (
	"encoding/json"

	"github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/errors"
)

// ---- Why this file differs from the original standalone package ----
//
// The original sendNotificationJob held three unexported fields —
// manager *Manager, notifiable Notifiable, n Notification — populated at
// construction time in Manager.dispatchQueued, then handed directly to
// m.queue.Job(job).
//
// That's correct for Goravel's "sync" queue driver, which (per the docs)
// executes inline with no serialization step — the exact same *Job value
// you passed in is the one whose Handle() gets called. It is NOT correct
// for any driver that persists/transmits the job (database, Redis): those
// drivers work by looking up a *registered* Job type (matched by
// Signature()) and calling Handle() on a freshly constructed instance
// with the dispatch-time []queue.Arg — see mail.ServiceProvider's
// registerJobs, which registers NewSendMailJob(configFacade) once at
// Boot() rather than relying on a specific job instance surviving. A
// freshly-constructed instance of the original sendNotificationJob would
// have nil manager/notifiable/n, and even if it didn't, unexported struct
// fields don't survive most Go serialization anyway (encoding/json drops
// them silently; gob requires them exported too).
//
// Net effect: the original design works in every test in this package
// (which exercises SendNow, not the queue path) and would appear to work
// in local development with the sync driver, but would silently fail —
// nil pointer panic in the worker, most likely — the first time someone
// runs it against a persistent queue driver in production. Worth
// confirming this reading against a maintainer or the queue driver source
// directly before treating it as certain, but the mail module's own
// registration pattern is strong corroborating evidence.
//
// Fix: resolve eagerly (Manager.dispatchQueued now calls
// ResolvableChannel.Resolve while notifiable/n are still live), queue
// only plain data, and register DispatchJob once via queueFacade.Register
// in the ServiceProvider's Boot() — exactly mirroring Mail's pattern.

// dispatchItem is the plain, JSON-serializable unit queued per channel.
type dispatchItem struct {
	Channel string `json:"channel"`
	Route   string `json:"route"`
	Payload []byte `json:"payload"`
}

func encodeDispatchItem(item dispatchItem) (string, error) {
	b, err := json.Marshal(item)
	if err != nil {
		return "", errors.NotificationInvalidQueuePayload
	}
	return string(b), nil
}

// DispatchJob delivers one resolved channel item. Registered once at
// Boot() (see service_provider.go), not constructed per-dispatch.
type DispatchJob struct {
	manager *Manager
}

func NewDispatchJob(manager *Manager) *DispatchJob {
	return &DispatchJob{manager: manager}
}

func (j *DispatchJob) Signature() string {
	return "goravel_notifications:dispatch"
}

func (j *DispatchJob) Handle(args ...any) error {
	if len(args) != 1 {
		return errors.NotificationInvalidQueuePayload
	}

	raw, ok := args[0].(string)
	if !ok {
		return errors.NotificationInvalidQueuePayload
	}

	var item dispatchItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return errors.NotificationInvalidQueuePayload
	}

	ch, err := j.manager.Channel(item.Channel)
	if err != nil {
		return err
	}

	resolvable, ok := ch.(notification.ResolvableChannel)
	if !ok {
		return errors.NotificationChannelNotQueueable.Args(item.Channel)
	}

	return resolvable.Deliver(item.Route, item.Payload)
}
