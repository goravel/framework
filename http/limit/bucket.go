package limit

import (
	"time"

	"github.com/goravel/framework/support/carbon"
)

// Bucket is an internal wrapper around a taker.
type Bucket struct {
	// startTime is the number of nanoseconds from unix epoch when this Bucket was
	// initially created.
	startTime uint64

	// maxTokens is the maximum number of tokens permitted on the Bucket at any
	// time. The number of available tokens will never exceed this value.
	maxTokens uint64

	// interval is the time at which ticking should occur.
	interval time.Duration

	// availableTokens is the current point-in-time number of tokens remaining.
	availableTokens uint64

	// lastTick is the last clock tick, used to re-calculate the number of tokens
	// on the Bucket.
	lastTick uint64
}

// NewBucket creates a new Bucket from the given tokens and interval.
func NewBucket(tokens uint64, interval time.Duration) *Bucket {
	b := &Bucket{
		startTime:       uint64(carbon.Now().TimestampNano()),
		maxTokens:       tokens,
		availableTokens: tokens,
		interval:        interval,
	}
	return b
}

// get returns information about the Bucket.
func (b *Bucket) get() (tokens uint64, remaining uint64, retErr error) {
	tokens = b.maxTokens
	remaining = b.availableTokens
	return
}

// take attempts to remove a token from the Bucket. If there are no tokens
// available and the clock has ticked forward, it recalculates the number of
// tokens and retries. It returns the limit, remaining tokens, time until
// refresh, and whether the take was successful.
func (b *Bucket) take() (tokens uint64, remaining uint64, reset uint64, ok bool, retErr error) {
	// Capture the current request time, current tick, and amount of time until
	// the Bucket resets.
	now := uint64(carbon.Now().TimestampNano())

	// If the current time is before the start time, it means the server clock was
	// reset to an earlier time. In that case, rebase to 0.
	if now < b.startTime {
		b.startTime = now
		b.lastTick = 0
	}

	currTick := tick(b.startTime, now, b.interval)

	tokens = b.maxTokens
	reset = b.startTime + ((currTick + 1) * uint64(b.interval))

	// If we're on a new tick since last assessment, perform
	// a full reset up to maxTokens.
	if b.lastTick < currTick {
		b.availableTokens = b.maxTokens
		b.lastTick = currTick
	}

	if b.availableTokens > 0 {
		b.availableTokens--
		ok = true
		remaining = b.availableTokens
	}

	return
}

// tick is the total number of times the current interval has occurred between
// when the time started (start) and the current time (curr). For example, if
// the start time was 12:30pm and it's currently 1:00pm, and the interval was 5
// minutes, tick would return 6 because 1:00pm is the 6th 5-minute tick. Note
// that tick would return 5 at 12:59pm, because it hasn't reached the 6th tick
// yet.
func tick(start, curr uint64, interval time.Duration) uint64 {
	return (curr - start) / uint64(interval.Nanoseconds())
}
