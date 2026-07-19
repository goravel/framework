package notification

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockslog "github.com/goravel/framework/mocks/log"
)

// fakeChannel and fakeResolvableChannel (used below) are defined in
// notification_test.go — shared fixtures, same package.

func TestDispatchJob_Signature(t *testing.T) {
	job := NewDispatchJob(NewManager(mockslog.NewLog(t), nil))
	assert.Equal(t, "goravel_notifications:dispatch", job.Signature())
}

func TestDispatchJob_Handle_ReturnsError_WhenNoArgs(t *testing.T) {
	job := NewDispatchJob(NewManager(mockslog.NewLog(t), nil))
	err := job.Handle()
	assert.Error(t, err)
}

func TestDispatchJob_Handle_ReturnsError_WhenTooManyArgs(t *testing.T) {
	job := NewDispatchJob(NewManager(mockslog.NewLog(t), nil))
	err := job.Handle("one", "two")
	assert.Error(t, err)
}

func TestDispatchJob_Handle_ReturnsError_WhenArgNotString(t *testing.T) {
	job := NewDispatchJob(NewManager(mockslog.NewLog(t), nil))
	err := job.Handle(42)
	assert.Error(t, err)
}

func TestDispatchJob_Handle_ReturnsError_WhenMalformedJSON(t *testing.T) {
	job := NewDispatchJob(NewManager(mockslog.NewLog(t), nil))
	err := job.Handle("{not valid json")
	assert.Error(t, err)
}

func TestDispatchJob_Handle_ReturnsError_WhenChannelNotRegistered(t *testing.T) {
	logger := mockslog.NewLog(t)
	logger.EXPECT().Errorf("%s", mock.Anything).Once()

	mgr := NewManager(logger, nil)
	job := NewDispatchJob(mgr)

	encoded, err := encodeDispatchItem(dispatchItem{Channel: "missing", Route: "r", Payload: []byte("{}")})
	assert.NoError(t, err)

	err = job.Handle(encoded)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestDispatchJob_Handle_ReturnsError_WhenChannelNotResolvable(t *testing.T) {
	logger := mockslog.NewLog(t)

	mgr := NewManager(logger, nil)
	mgr.Extend(&fakeChannel{name: "plain"}) // Channel only, not ResolvableChannel
	job := NewDispatchJob(mgr)

	encoded, err := encodeDispatchItem(dispatchItem{Channel: "plain", Route: "r", Payload: []byte("{}")})
	assert.NoError(t, err)

	err = job.Handle(encoded)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support queued dispatch")
}

func TestDispatchJob_Handle_DeliversSuccessfully(t *testing.T) {
	logger := mockslog.NewLog(t)

	mgr := NewManager(logger, nil)
	ch := &fakeResolvableChannel{name: "ok"}
	mgr.Extend(ch)
	job := NewDispatchJob(mgr)

	encoded, err := encodeDispatchItem(dispatchItem{Channel: "ok", Route: "user@example.com", Payload: []byte(`{"subject":"hi"}`)})
	assert.NoError(t, err)

	err = job.Handle(encoded)
	assert.NoError(t, err)
	assert.Len(t, ch.delivered, 1)
	assert.Equal(t, "user@example.com", ch.delivered[0].route)
	assert.JSONEq(t, `{"subject":"hi"}`, string(ch.delivered[0].payload))
}

func TestDispatchJob_Handle_PropagatesDeliverError(t *testing.T) {
	logger := mockslog.NewLog(t)

	mgr := NewManager(logger, nil)
	mgr.Extend(&fakeResolvableChannel{name: "broken", deliverErr: errors.New("smtp down")})
	job := NewDispatchJob(mgr)

	encoded, err := encodeDispatchItem(dispatchItem{Channel: "broken", Route: "r", Payload: []byte("{}")})
	assert.NoError(t, err)

	err = job.Handle(encoded)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "smtp down")
}
