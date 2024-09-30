package limit

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/http"
	"github.com/goravel/framework/support/carbon"
)

// Code from https://github.com/sethvargo/go-limiter

// Store is an interface for limiter storage backends.
//
// Keys should be hash, sanitized, or otherwise scrubbed of identifiable
// information they will be given to the store in plaintext. If you're rate
// limiting by IP address, for example, the IP address would be stored in the
// storage system in plaintext. This may be undesirable in certain situations,
// like when the store is a public database. In those cases, you should hash or
// HMAC the key before passing giving it to the store. If you want to encrypt
// the value, you must use homomorphic encryption to ensure the value always
// encrypts to the same ciphertext.
type Store interface {
	// Take takes a token from the given key if available, returning:
	//
	// - the configured limit size
	// - the number of remaining tokens in the interval
	// - the server time when new tokens will be available
	// - whether the take was successful
	// - any errors that occurred while performing the take - these should be
	//   backend errors (e.g. connection failures); Take() should never return an
	//   error for a bucket.
	//
	// If "ok" is false, the take was unsuccessful and the caller should NOT
	// service the request.
	//
	// See the note about keys on the interface documentation.
	Take(ctx context.Context, key string) (tokens, remaining, reset uint64, ok bool, err error)

	// Get gets the current limit and remaining tokens for the provided key. It
	// does not change any of the values.
	Get(ctx context.Context, key string) (tokens, remaining uint64, err error)

	// Set configures the limit at the provided key. If a limit already exists, it
	// is overwritten. This also sets the number of tokens in the bucket to the
	// limit.
	Set(ctx context.Context, key string, tokens uint64, interval time.Duration) error

	// Burst adds more tokens to the key's current bucket until the next interval
	// tick. This will allow the current bucket tick to exceed the maximum number
	// maximum ticks until the next interval.
	Burst(ctx context.Context, key string, tokens uint64) error
}

type store struct {
	tokens   uint64
	interval time.Duration
	cache    cache.Cache
}

func NewStore(tokens uint64, interval time.Duration) (Store, error) {
	if tokens <= 0 {
		tokens = 1
	}

	if interval <= 0 {
		interval = 1 * time.Second
	}

	if http.CacheFacade == nil {
		return nil, errors.New("cache facade is not initialized")
	}

	s := &store{
		tokens:   tokens,
		interval: interval,
		cache:    http.CacheFacade,
	}

	return s, nil
}

// Take attempts to remove a token from the named key. If the take is
// successful, it returns true, otherwise false. It also returns the configured
// limit, remaining tokens, and reset time.
func (s *store) Take(_ context.Context, key string) (uint64, uint64, uint64, bool, error) {
	b, ok := s.cache.Get(key).(*Bucket)
	if ok {
		return b.take()
	}

	nb := NewBucket(s.tokens, s.interval)
	if err := s.cache.Put(key, nb, s.interval); err != nil {
		return 0, 0, 0, false, err
	}

	return nb.take()
}

// Get retrieves the information about the key, if any exists.
func (s *store) Get(_ context.Context, key string) (uint64, uint64, error) {
	b, ok := s.cache.Get(key).(*Bucket)
	if ok {
		return b.get()
	}

	return 0, 0, nil
}

// Set configures the Bucket-specific tokens and interval.
func (s *store) Set(_ context.Context, key string, tokens uint64, interval time.Duration) error {
	b := NewBucket(tokens, interval)
	return s.cache.Put(key, b, interval)
}

// Burst adds the provided value to the Bucket's currently available tokens.
func (s *store) Burst(_ context.Context, key string, tokens uint64) error {
	b, ok := s.cache.Get(key).(*Bucket)
	if ok {
		b.lock.Lock()
		b.availableTokens = b.availableTokens + tokens
		b.lock.Unlock()
		return nil
	}

	nb := NewBucket(s.tokens+tokens, s.interval)
	return s.cache.Put(key, nb, s.interval)
}

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

	// lock guards the mutable fields.
	lock sync.Mutex
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
	b.lock.Lock()
	defer b.lock.Unlock()

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

	b.lock.Lock()
	defer b.lock.Unlock()

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
