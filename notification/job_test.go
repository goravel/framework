package notification_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocklog "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/notification"
)

// wireDispatchItem mirrors the private dispatchItem's JSON shape so tests
// outside the package can build valid/invalid queue payloads without
// needing access to the unexported type.
type wireDispatchItem struct {
	Channel string `json:"channel"`
	Route   string `json:"route"`
	Payload []byte `json:"payload"`
}

func encodeWireItem(t *testing.T, item wireDispatchItem) string {
	t.Helper()
	b, err := json.Marshal(item)
	assert.NoError(t, err)
	return string(b)
}

func TestDispatchJob_Signature(t *testing.T) {
	job := notification.NewDispatchJob(notification.NewManager(mocklog.NewLog(t), nil))
	assert.Equal(t, "goravel_notifications:dispatch", job.Signature())
}

func TestDispatchJob_Handle_ReturnsError_WhenNoArgs(t *testing.T) {
	job := notification.NewDispatchJob(notification.NewManager(mocklog.NewLog(t), nil))
	err := job.Handle()
	assert.Error(t, err)
}

func TestDispatchJob_Handle_ReturnsError_WhenTooManyArgs(t *testing.T) {
	job := notification.NewDispatchJob(notification.NewManager(mocklog.NewLog(t), nil))
	err := job.Handle("one", "two")
	assert.Error(t, err)
}

func TestDispatchJob_Handle_ReturnsError_WhenArgNotString(t *testing.T) {
	job := notification.NewDispatchJob(notification.NewManager(mocklog.NewLog(t), nil))
	err := job.Handle(42)
	assert.Error(t, err)
}

func TestDispatchJob_Handle_ReturnsError_WhenMalformedJSON(t *testing.T) {
	job := notification.NewDispatchJob(notification.NewManager(mocklog.NewLog(t), nil))
	err := job.Handle("{not valid json")
	assert.Error(t, err)
}

func TestDispatchJob_Handle_ReturnsError_WhenChannelNotRegistered(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Errorf", mock.Anything, mock.Anything).Once()

	mgr := notification.NewManager(logger, nil)
	job := notification.NewDispatchJob(mgr)

	payload := encodeWireItem(t, wireDispatchItem{Channel: "missing", Route: "r", Payload: []byte("{}")})
	err := job.Handle(payload)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestDispatchJob_Handle_ReturnsError_WhenChannelNotResolvable(t *testing.T) {
	logger := mocklog.NewLog(t)

	mgr := notification.NewManager(logger, nil)
	mgr.Extend(&fakeChannel{name: "plain"}) // Channel only, not ResolvableChannel
	job := notification.NewDispatchJob(mgr)

	payload := encodeWireItem(t, wireDispatchItem{Channel: "plain", Route: "r", Payload: []byte("{}")})
	err := job.Handle(payload)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support queued dispatch")
}

func TestDispatchJob_Handle_DeliversSuccessfully(t *testing.T) {
	logger := mocklog.NewLog(t)

	mgr := notification.NewManager(logger, nil)
	ch := &fakeResolvableChannel{name: "ok"}
	mgr.Extend(ch)
	job := notification.NewDispatchJob(mgr)

	payload := encodeWireItem(t, wireDispatchItem{Channel: "ok", Route: "user@example.com", Payload: []byte(`{"subject":"hi"}`)})
	err := job.Handle(payload)

	assert.NoError(t, err)
	assert.Len(t, ch.delivered, 1)
	assert.Equal(t, "user@example.com", ch.delivered[0].route)
	assert.JSONEq(t, `{"subject":"hi"}`, string(ch.delivered[0].payload))
}

func TestDispatchJob_Handle_PropagatesDeliverError(t *testing.T) {
	logger := mocklog.NewLog(t)

	mgr := notification.NewManager(logger, nil)
	mgr.Extend(&fakeResolvableChannel{name: "broken", deliverErr: errors.New("smtp down")})
	job := notification.NewDispatchJob(mgr)

	payload := encodeWireItem(t, wireDispatchItem{Channel: "broken", Route: "r", Payload: []byte("{}")})
	err := job.Handle(payload)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "smtp down")
}
