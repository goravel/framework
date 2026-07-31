package notification

import (
	"errors"
	"testing"
	"time"

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

func TestDispatchJob_ShouldRetry_NoRetry_WhenPlainError(t *testing.T) {
	logger := mockslog.NewLog(t)
	mgr := NewManager(logger, nil)
	mgr.Extend(&fakeResolvableChannel{name: "a", deliverErr: errors.New("smtp down")})
	job := NewDispatchJob(mgr)

	encoded, err := encodeDispatchItem(dispatchItem{Channel: "a", Route: "r", Payload: []byte("{}")})
	assert.NoError(t, err)

	handleErr := job.Handle(encoded)
	assert.Error(t, handleErr)

	retryable, delay := job.ShouldRetry(handleErr, 1)
	assert.False(t, retryable)
	assert.Zero(t, delay)
}

func TestDispatchJob_ShouldRetry_RetriesWithBackoff(t *testing.T) {
	logger := mockslog.NewLog(t)
	mgr := NewManager(logger, nil)
	mgr.Extend(&fakeResolvableChannel{name: "a", deliverErr: errors.New("smtp down")})
	job := NewDispatchJob(mgr)

	encoded, err := encodeDispatchItem(dispatchItem{
		Channel: "a", Route: "r", Payload: []byte("{}"),
		BackoffSeconds: 45,
	})
	assert.NoError(t, err)

	handleErr := job.Handle(encoded)
	assert.Error(t, handleErr)

	retryable, delay := job.ShouldRetry(handleErr, 1)
	assert.True(t, retryable)
	assert.Equal(t, 45*time.Second, delay)
}

func TestDispatchJob_ShouldRetry_StopsAfterRetryUntilDeadline(t *testing.T) {
	logger := mockslog.NewLog(t)
	mgr := NewManager(logger, nil)
	mgr.Extend(&fakeResolvableChannel{name: "a", deliverErr: errors.New("smtp down")})
	job := NewDispatchJob(mgr)

	past := time.Now().Add(-1 * time.Hour).Unix()
	encoded, err := encodeDispatchItem(dispatchItem{
		Channel: "a", Route: "r", Payload: []byte("{}"),
		BackoffSeconds: 30, RetryUntilUnix: past,
	})
	assert.NoError(t, err)

	handleErr := job.Handle(encoded)
	assert.Error(t, handleErr)

	retryable, delay := job.ShouldRetry(handleErr, 5)
	assert.False(t, retryable, "past RetryUntil should stop retries even though Backoff is set")
	assert.Zero(t, delay)
}

func TestDispatchJob_ShouldRetry_CapsAttempts_WhenNoRetryUntilSet(t *testing.T) {
	logger := mockslog.NewLog(t)
	mgr := NewManager(logger, nil)
	mgr.Extend(&fakeResolvableChannel{name: "a", deliverErr: errors.New("smtp down")})
	job := NewDispatchJob(mgr)

	encoded, err := encodeDispatchItem(dispatchItem{
		Channel: "a", Route: "r", Payload: []byte("{}"),
		BackoffSeconds: 5, // RetryUntilUnix deliberately left unset
	})
	assert.NoError(t, err)

	handleErr := job.Handle(encoded)
	assert.Error(t, handleErr)

	retryable, delay := job.ShouldRetry(handleErr, DefaultMaxRetryAttempts-1)
	assert.True(t, retryable, "still below the cap, should retry")
	assert.Equal(t, 5*time.Second, delay)

	retryable, delay = job.ShouldRetry(handleErr, DefaultMaxRetryAttempts)
	assert.False(t, retryable, "at the cap, should stop even though Backoff is set")
	assert.Zero(t, delay)
}
