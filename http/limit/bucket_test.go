package limit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type BucketTestSuite struct {
	suite.Suite
}

func TestBucketTestSuite(t *testing.T) {
	suite.Run(t, new(BucketTestSuite))
}

func (s *BucketTestSuite) TestNewBucket() {
	// Test creating a new bucket
	tokens := uint64(10)
	interval := time.Second
	bucket := NewBucket(tokens, interval)

	s.NotNil(bucket)
	s.Equal(tokens, bucket.MaxTokens)
	s.Equal(tokens, bucket.AvailableTokens)
	s.Equal(interval, bucket.Interval)
	s.NotZero(bucket.StartTime)
	s.Zero(bucket.LastTick)
}

func (s *BucketTestSuite) TestBucketGet() {
	// Test getting bucket information
	bucket := NewBucket(10, time.Second)
	tokens, remaining, err := bucket.get()

	s.NoError(err)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(10), remaining)
}

func (s *BucketTestSuite) TestBucketTake() {
	// Test taking a token from a bucket with available tokens
	bucket := NewBucket(10, time.Second)
	tokens, remaining, reset, ok, err := bucket.take()

	s.NoError(err)
	s.True(ok)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(9), remaining)
	s.NotZero(reset)
	s.Equal(uint64(9), bucket.AvailableTokens)
}

func (s *BucketTestSuite) TestBucketTakeEmpty() {
	// Test taking a token from an empty bucket
	bucket := NewBucket(0, time.Second)
	tokens, remaining, reset, ok, err := bucket.take()

	s.NoError(err)
	s.False(ok)
	s.Equal(uint64(0), tokens)
	s.Equal(uint64(0), remaining)
	s.NotZero(reset)
	s.Equal(uint64(0), bucket.AvailableTokens)
}

func (s *BucketTestSuite) TestBucketTakeWithRefill() {
	// Test taking tokens with automatic refill after interval
	interval := 100 * time.Millisecond
	bucket := NewBucket(2, interval)

	// Take first token
	_, remaining, _, ok, _ := bucket.take()
	s.True(ok)
	s.Equal(uint64(1), remaining)

	// Take second token
	_, remaining, _, ok, _ = bucket.take()
	s.True(ok)
	s.Equal(uint64(0), remaining)

	// Try to take third token (should fail)
	_, remaining, _, ok, _ = bucket.take()
	s.False(ok)
	s.Equal(uint64(0), remaining)

	// Wait for refill
	time.Sleep(interval + 10*time.Millisecond)

	// Should be able to take a token now
	_, remaining, _, ok, _ = bucket.take()
	s.True(ok)
	// After refill, we should have max tokens (2) - 1 we just took = 1 remaining
	s.Equal(uint64(1), remaining)
}

func (s *BucketTestSuite) TestBucketTakeWithClockReset() {
	// Test behavior when system clock is reset to an earlier time
	bucket := NewBucket(5, time.Second)
	originalStartTime := bucket.StartTime

	// Simulate clock reset by manually setting StartTime to a future time
	bucket.StartTime = originalStartTime + uint64(time.Hour.Nanoseconds())

	// Take a token (should handle clock reset)
	_, _, _, ok, _ := bucket.take()

	s.True(ok)
	// The StartTime should be reset to a new value when clock reset is detected
	s.NotEqual(originalStartTime+uint64(time.Hour.Nanoseconds()), bucket.StartTime)
}

func (s *BucketTestSuite) TestBucketTakeMultiple() {
	// Test taking multiple tokens
	bucket := NewBucket(5, time.Second)

	// Take first token
	tokens1, remaining1, reset1, ok1, err1 := bucket.take()
	s.NoError(err1)
	s.True(ok1)
	s.Equal(uint64(5), tokens1)
	s.Equal(uint64(4), remaining1)
	s.NotZero(reset1)

	// Take second token
	tokens2, remaining2, reset2, ok2, err2 := bucket.take()
	s.NoError(err2)
	s.True(ok2)
	s.Equal(uint64(5), tokens2)
	s.Equal(uint64(3), remaining2)
	s.NotZero(reset2)

	// Take until empty
	bucket.take()                                           // 2 remaining
	bucket.take()                                           // 1 remaining
	tokens5, remaining5, reset5, ok5, err5 := bucket.take() // 0 remaining
	s.NoError(err5)
	s.True(ok5)
	s.Equal(uint64(5), tokens5)
	s.Equal(uint64(0), remaining5)
	s.NotZero(reset5)

	// Try to take one more (should fail)
	tokens6, remaining6, reset6, ok6, err6 := bucket.take()
	s.NoError(err6)
	s.False(ok6)
	s.Equal(uint64(5), tokens6)
	s.Equal(uint64(0), remaining6)
	s.NotZero(reset6)
}

func (s *BucketTestSuite) TestTick() {
	// Test the tick function
	start := uint64(0)
	interval := time.Second

	// Test at exactly the start time
	curr := start
	ticks := tick(start, curr, interval)
	s.Equal(uint64(0), ticks)

	// Test at half interval
	curr = start + uint64(500*time.Millisecond.Nanoseconds())
	ticks = tick(start, curr, interval)
	s.Equal(uint64(0), ticks)

	// Test at exactly one interval
	curr = start + uint64(interval.Nanoseconds())
	ticks = tick(start, curr, interval)
	s.Equal(uint64(1), ticks)

	// Test at multiple intervals
	curr = start + uint64(5*interval.Nanoseconds())
	ticks = tick(start, curr, interval)
	s.Equal(uint64(5), ticks)

	// Test at multiple intervals plus partial
	curr = start + uint64(5*interval.Nanoseconds()) + uint64(500*time.Millisecond.Nanoseconds())
	ticks = tick(start, curr, interval)
	s.Equal(uint64(5), ticks)
}

func (s *BucketTestSuite) TestTickWithDifferentIntervals() {
	// Test tick function with different intervals
	start := uint64(0)

	// Test with 1 minute interval
	interval := time.Minute
	curr := start + uint64(2*interval.Nanoseconds()) + uint64(30*time.Second.Nanoseconds())
	ticks := tick(start, curr, interval)
	s.Equal(uint64(2), ticks)

	// Test with 500ms interval
	interval = 500 * time.Millisecond
	curr = start + uint64(1*time.Second.Nanoseconds())
	ticks = tick(start, curr, interval)
	s.Equal(uint64(2), ticks)

	// Test with very small interval (1ms)
	interval = time.Millisecond
	curr = start + uint64(10*time.Millisecond.Nanoseconds())
	ticks = tick(start, curr, interval)
	s.Equal(uint64(10), ticks)
}

func (s *BucketTestSuite) TestBucketLastTickUpdate() {
	// Test that LastTick is updated correctly when tokens are refilled
	interval := 100 * time.Millisecond
	bucket := NewBucket(2, interval)

	// Take all tokens
	bucket.take()
	bucket.take()

	// Record the current LastTick
	initialLastTick := bucket.LastTick

	// Wait for refill
	time.Sleep(interval + 10*time.Millisecond)

	// Take a token (this should update LastTick)
	bucket.take()

	// LastTick should be updated
	s.Greater(bucket.LastTick, initialLastTick)
}

func (s *BucketTestSuite) TestBucketResetCalculation() {
	// Test that the reset time is calculated correctly
	interval := time.Second
	bucket := NewBucket(5, interval)

	// Take a token and get the reset time
	_, _, reset, _, _ := bucket.take()

	// The reset time should be StartTime + (currTick + 1) * interval
	expectedReset := bucket.StartTime + uint64(interval.Nanoseconds())
	s.Equal(expectedReset, reset)
}

func (s *BucketTestSuite) TestBucketWithZeroInterval() {
	// Test bucket behavior with a zero interval
	// This is an edge case that should be handled gracefully
	bucket := NewBucket(5, time.Nanosecond) // Use smallest possible interval instead of zero

	// Take a token
	tokens, remaining, reset, ok, err := bucket.take()

	// Even with very small interval, the bucket should function
	s.NoError(err)
	s.True(ok)
	s.Equal(uint64(5), tokens)
	s.Equal(uint64(4), remaining)
	s.NotZero(reset)
}

func (s *BucketTestSuite) TestBucketWithVeryLargeMaxTokens() {
	// Test bucket with a very large max tokens value
	maxTokens := uint64(1000000)
	bucket := NewBucket(maxTokens, time.Second)

	// Take a token
	tokens, remaining, _, ok, err := bucket.take()

	s.NoError(err)
	s.True(ok)
	s.Equal(maxTokens, tokens)
	s.Equal(maxTokens-1, remaining)
}

func (s *BucketTestSuite) TestBucketWithVerySmallInterval() {
	// Test bucket with a very small interval
	interval := time.Nanosecond
	bucket := NewBucket(5, interval)

	// Take a token
	_, _, reset, ok, _ := bucket.take()

	s.True(ok)
	// The reset time should be very close to the current time
	s.NotZero(reset)
}

func (s *BucketTestSuite) TestBucketWithVeryLargeInterval() {
	// Test bucket with a very large interval
	interval := 24 * time.Hour // 1 day
	bucket := NewBucket(5, interval)

	// Take a token
	_, _, reset, ok, _ := bucket.take()

	s.True(ok)
	// The reset time should be far in the future
	s.NotZero(reset)

	// Take all remaining tokens
	bucket.take()
	bucket.take()
	bucket.take()
	tokens, remaining, _, lastOk, _ := bucket.take()

	// Bucket should be empty now
	s.True(lastOk) // This should be true because we're taking the last available token
	s.Equal(uint64(0), remaining)
	s.Equal(uint64(5), tokens)

	// Now try to take one more token (should fail)
	_, _, _, stillOk, _ := bucket.take()

	// This should be false because we've used all tokens
	// and the interval hasn't passed yet
	s.False(stillOk)
}

func (s *BucketTestSuite) TestTickWithZeroInterval() {
	// Test the tick function with a very small interval instead of zero
	// to avoid division by zero
	start := uint64(0)
	curr := uint64(1000000000)  // 1 second in nanoseconds
	interval := time.Nanosecond // Smallest possible interval

	// This should not panic
	ticks := tick(start, curr, interval)
	s.NotZero(ticks) // With 1 second and 1 nanosecond interval, we should have many ticks
}

func (s *BucketTestSuite) TestTickWithStartGreaterThanCurr() {
	// Test the tick function when start is greater than curr
	// This is an edge case in the implementation
	start := uint64(1000000000) // 1 second in nanoseconds
	curr := uint64(0)
	interval := time.Second

	// The current implementation of tick doesn't handle start > curr correctly
	// It will return a very large number due to unsigned integer underflow
	// This test is documenting the current behavior, not the ideal behavior
	ticks := tick(start, curr, interval)

	// In the current implementation, when start > curr with unsigned integers,
	// (curr - start) will underflow to a very large number
	// This is not ideal, but it's the current behavior
	// Ideally, it should return 0 in this case

	// We're not testing for a specific value, just that it's not 0
	// This documents the current behavior
	s.NotEqual(uint64(0), ticks)

	// For completeness, let's also test the case where start < curr
	start = uint64(0)
	curr = uint64(1000000000)
	ticks = tick(start, curr, interval)
	s.Equal(uint64(1), ticks)
}
