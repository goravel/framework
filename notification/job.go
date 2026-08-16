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
		return errors.NotificationQueuePayloadDecodeFailed.Args(err)
	}

	j.mu.Lock()
	j.item = &item
	j.mu.Unlock()

	ch := j.manager.Channel(item.Channel)
	if ch == nil {
		j.mu.Lock()
		j.item = nil
		j.mu.Unlock()
		return errors.NotificationChannelNotFound.Args(item.Channel)
	}

	resolvable, ok := ch.(notification.ResolvableChannel)
	if !ok {
		j.mu.Lock()
		j.item = nil
		j.mu.Unlock()
		return errors.NotificationChannelNotQueueable.Args(item.Channel)
	}

	if err := resolvable.Deliver(item.Route, item.Payload); err != nil {
		return err
	}

	j.mu.Lock()
	j.item = nil
	j.mu.Unlock()

	return nil
}

func (j *DispatchJob) ShouldRetry(err error, attempt int) (retryable bool, delay time.Duration) {
	j.mu.Lock()
	item := j.item
	j.mu.Unlock()

	if item == nil || item.Tries <= 0 {
		return false, 0
	}
	if attempt < 1 {

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
