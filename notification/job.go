package notification

import (
	"encoding/json"

	"github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/errors"
)

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

	return resolvable.Deliver(item.Route, item.Payload)
}
