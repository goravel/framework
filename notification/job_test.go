package notification

import (
	"encoding/json"
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

// TestDispatchJob_ShouldRetry mirrors broadcasting/job_test.go's
// TestBroadcastJob_ShouldRetry exactly — same case names, same edge
// cases — since DispatchJob.ShouldRetry now uses the identical
// Tries/Backoff/mutex-guarded-item design as BroadcastJob, adopted for
// consistency across the two queued-retry mechanisms in this codebase.
func TestDispatchJob_ShouldRetry(t *testing.T) {
	marshalItem := func(tries int, backoff []int64) string {
		item := dispatchItem{
			Channel: "a",
			Route:   "r",
			Payload: []byte("{}"),
			Tries:   tries,
			Backoff: backoff,
		}
		data, _ := json.Marshal(item)
		return string(data)
	}

	newJob := func(t *testing.T) *DispatchJob {
		logger := mockslog.NewLog(t)
		mgr := NewManager(logger, nil)
		mgr.Extend(&fakeResolvableChannel{name: "a", deliverErr: errors.New("smtp down")})
		return NewDispatchJob(mgr)
	}

	tests := []struct {
		name    string
		payload string
		attempt int
		err     error
		want    bool
		wantD   time.Duration
	}{
		{
			name:    "tries zero is single-shot",
			payload: marshalItem(0, nil),
			attempt: 1,
			want:    false,
			wantD:   0,
		},
		{
			name:    "tries 3 first attempt retries",
			payload: marshalItem(3, nil),
			attempt: 1,
			want:    true,
			wantD:   0,
		},
		{
			name:    "tries 3 second attempt retries",
			payload: marshalItem(3, nil),
			attempt: 2,
			want:    true,
			wantD:   0,
		},
		{
			name:    "tries 3 third attempt stops",
			payload: marshalItem(3, nil),
			attempt: 3,
			want:    false,
			wantD:   0,
		},
		{
			name:    "backoff first attempt",
			payload: marshalItem(4, []int64{1000, 2000}),
			attempt: 1,
			err:     errors.New("test error"), // err is ignored by ShouldRetry
			want:    true,
			wantD:   1 * time.Second,
		},
		{
			name:    "backoff second attempt",
			payload: marshalItem(4, []int64{1000, 2000}),
			attempt: 2,
			want:    true,
			wantD:   2 * time.Second,
		},
		{
			name:    "backoff last value repeats",
			payload: marshalItem(4, []int64{1000, 2000}),
			attempt: 3,
			want:    true,
			wantD:   2 * time.Second,
		},
		{
			name:    "backoff with attempt 0 is single-shot fallback",
			payload: marshalItem(4, []int64{1000, 2000}),
			attempt: 0,
			want:    false,
			wantD:   0,
		},
		{
			name:    "backoff stop at final attempt before index",
			payload: marshalItem(2, []int64{1000, 2000}),
			attempt: 2,
			want:    false,
			wantD:   0,
		},
		{
			name:    "backoff last attempt stops",
			payload: marshalItem(4, []int64{1000, 2000}),
			attempt: 4,
			want:    false,
			wantD:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := newJob(t)
			// Handle fails via fakeResolvableChannel's deliverErr, which
			// retains the parsed item for ShouldRetry to read — the
			// realistic pre-ShouldRetry state, mirroring
			// broadcasting/job_test.go's own setup.
			assert.Error(t, job.Handle(tt.payload))

			retryable, delay := job.ShouldRetry(tt.err, tt.attempt)
			assert.Equal(t, tt.want, retryable)
			assert.Equal(t, tt.wantD, delay)
		})
	}
}
