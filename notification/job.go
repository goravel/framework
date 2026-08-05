package notification

import (
	"encoding/json"
	"time"

	"github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/errors"
)

// dispatchItem is the plain, JSON-serializable unit queued per channel.
type dispatchItem struct {
	Channel        string `json:"channel"`
	Route          string `json:"route"`
	Payload        []byte `json:"payload"`
	BackoffSeconds int    `json:"backoff_seconds,omitempty"`
	RetryUntilUnix int64  `json:"retry_until_unix,omitempty"`
}

func encodeDispatchItem(item dispatchItem) (string, error) {
	b, err := json.Marshal(item)
	if err != nil {
		return "", errors.NotificationInvalidQueuePayload
	}
	return string(b), nil
}

var DefaultMaxRetryAttempts = 10

type deliveryError struct {
	err           error
	hasBackoff    bool
	backoff       time.Duration
	hasRetryUntil bool
	retryUntil    time.Time
}

func (e *deliveryError) Error() string { return e.err.Error() }
func (e *deliveryError) Unwrap() error { return e.err }

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

	ch := j.manager.Channel(item.Channel)
	if ch == nil {
		return errors.NotificationChannelNotFound.Args(item.Channel)
	}

	resolvable, ok := ch.(notification.ResolvableChannel)
	if !ok {
		return errors.NotificationChannelNotQueueable.Args(item.Channel)
	}

	err := resolvable.Deliver(item.Route, item.Payload)
	if err == nil {
		return nil
	}

	if item.BackoffSeconds == 0 && item.RetryUntilUnix == 0 {
		return err
	}

	wrapped := &deliveryError{err: err}
	if item.BackoffSeconds > 0 {
		wrapped.hasBackoff = true
		wrapped.backoff = time.Duration(item.BackoffSeconds) * time.Second
	}
	if item.RetryUntilUnix > 0 {
		wrapped.hasRetryUntil = true
		wrapped.retryUntil = time.Unix(item.RetryUntilUnix, 0)
	}
	return wrapped
}

func (j *DispatchJob) ShouldRetry(err error, attempt int) (bool, time.Duration) {
	var de *deliveryError
	if !errors.As(err, &de) {
		return false, 0
	}

	if de.hasRetryUntil {
		if time.Now().After(de.retryUntil) {
			return false, 0
		}
	} else if attempt >= DefaultMaxRetryAttempts {
		return false, 0
	}

	if de.hasBackoff {
		return true, de.backoff
	}

	return false, 0
}
