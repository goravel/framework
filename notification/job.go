package notification

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/errors"
)

// dispatchItem is the plain, JSON-serializable unit queued per channel.
// Tries/Backoff mirror broadcasting's broadcastItem exactly (same field
// names, same wire format — Backoff in milliseconds, not seconds) for
// consistency between the two queued-retry mechanisms in this codebase.
type dispatchItem struct {
	Channel string  `json:"channel"`
	Route   string  `json:"route"`
	Payload []byte  `json:"payload"`
	Tries   int     `json:"tries,omitempty"`
	Backoff []int64 `json:"backoff,omitempty"` // per-attempt delay in ms
}

func encodeDispatchItem(item dispatchItem) (string, error) {
	b, err := json.Marshal(item)
	if err != nil {
		return "", errors.NotificationInvalidQueuePayload
	}
	return string(b), nil
}

// DispatchJob delivers one resolved channel item. It's registered once
// with the queue at Boot() (see service_provider.go) rather than
// constructed per-dispatch, since persisting queue drivers (database,
// Redis) look up a registered Job by Signature() and call Handle() on a
// freshly constructed instance with the dispatch-time []queue.Arg. That's
// why Manager.dispatchQueued resolves each channel's payload eagerly via
// ResolvableChannel.Resolve — while notifiable/notification are still
// live — and queues only the resulting plain (channel, route, payload,
// tries, backoff).
//
// Retry state (item) is a shared, mutex-guarded field rather than
// carried in the error, matching broadcasting.BroadcastJob's exact
// pattern rather than an independently-derived design — see that type's
// own doc comment for the full reasoning, copied here:
//
// item is the payload of the task being processed, set by Handle and
// read by ShouldRetry. DispatchJob is a shared singleton, so access is
// guarded by mu.
//
// Limitation: the mutex guarantees memory safety, not logical
// isolation. ShouldRetry has no access to task args by contract, so a
// concurrent failed task can still overwrite this payload between
// another task's Handle returning an error and its ShouldRetry call.
// Clearing item on Handle's success path converts the common
// interleaving case into the safe single-shot fallback; a full fix
// requires per-task state passed by the queue worker, which is out of
// scope.
type DispatchJob struct {
	manager *Manager

	mu   sync.Mutex
	item *dispatchItem
}

func NewDispatchJob(manager *Manager) *DispatchJob {
	return &DispatchJob{manager: manager}
}

func (j *DispatchJob) Signature() string {
	return "goravel_notifications:dispatch"
}

func (j *DispatchJob) Handle(args ...any) error {
	if len(args) != 1 {
		j.mu.Lock()
		j.item = nil
		j.mu.Unlock()
		return errors.NotificationInvalidQueuePayload
	}

	raw, ok := args[0].(string)
	if !ok {
		j.mu.Lock()
		j.item = nil
		j.mu.Unlock()
		return errors.NotificationInvalidQueuePayload
	}

	var item dispatchItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		j.mu.Lock()
		j.item = nil
		j.mu.Unlock()
		return errors.NotificationInvalidQueuePayload
	}

	j.mu.Lock()
	j.item = &item
	j.mu.Unlock()

	ch := j.manager.Channel(item.Channel)
	if ch == nil {
		return errors.NotificationChannelNotFound.Args(item.Channel)
	}

	resolvable, ok := ch.(notification.ResolvableChannel)
	if !ok {
		return errors.NotificationChannelNotQueueable.Args(item.Channel)
	}

	if err := resolvable.Deliver(item.Route, item.Payload); err != nil {
		return err
	}

	// A successful task is never consulted by ShouldRetry, so release
	// the payload: an interleaving concurrent failed task then reads
	// the safe single-shot fallback (item == nil) instead of a wrong
	// retry policy.
	j.mu.Lock()
	j.item = nil
	j.mu.Unlock()

	return nil
}

// ShouldRetry implements the optional queue.Job retry-control interface
// documented at https://www.goravel.dev/digging-deeper/queues.html#job-retry.
// Logic mirrors broadcasting.BroadcastJob.ShouldRetry exactly: without
// Tries the notification is single-shot regardless of the worker's own
// tries config; with Tries it retries while attempt < Tries using the
// configured per-attempt Backoff (last value repeats).
func (j *DispatchJob) ShouldRetry(err error, attempt int) (retryable bool, delay time.Duration) {
	j.mu.Lock()
	item := j.item
	j.mu.Unlock()

	if item == nil || item.Tries <= 0 {
		return false, 0
	}
	if attempt < 1 {
		// Defensive: attempts come from the pop-incremented reservation
		// (or a chain counter starting at 1), so this is unreachable in
		// practice. Returning false avoids an accidental infinite
		// retry loop.
		return false, 0
	}
	if attempt >= item.Tries {
		return false, 0
	}
	if len(item.Backoff) == 0 {
		return true, 0
	}

	idx := min(attempt-1, len(item.Backoff)-1)
	return true, time.Duration(item.Backoff[idx]) * time.Millisecond
}
