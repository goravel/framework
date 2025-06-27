package limit

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type Store struct {
	cache    cache.Cache
	json     foundation.Json
	tokens   uint64
	interval time.Duration
}

func NewStore(cache cache.Cache, json foundation.Json, tokens uint64, interval time.Duration) *Store {
	if tokens <= 0 {
		tokens = 1
	}

	if interval <= 0 {
		interval = 1 * time.Second
	}

	s := &Store{
		tokens:   tokens,
		interval: interval,
		cache:    cache,
		json:     json,
	}

	return s
}

// Take attempts to remove a token from the named key. If the take is
// successful, it returns true, otherwise false. It also returns the configured
// limit, remaining tokens, and reset time.
func (r *Store) Take(ctx context.Context, key string) (tokens uint64, remaining uint64, reset uint64, ok bool, err error) {
	lock, err := r.lock(key)
	if err != nil {
		return 0, 0, 0, false, err
	}

	defer lock.Release()

	bucket, err := r.getBucket(ctx, key)
	if err != nil {
		return 0, 0, 0, false, err
	}

	if bucket == nil {
		bucket = NewBucket(r.tokens, r.interval)
	}

	tokens, remaining, reset, ok, err = bucket.take()
	if err != nil {
		return 0, 0, 0, false, err
	}

	if err := r.putBucket(ctx, key, bucket); err != nil {
		return 0, 0, 0, false, err
	}

	return tokens, remaining, reset, ok, nil
}

// Get retrieves the information about the key.
func (r *Store) Get(ctx context.Context, key string) (tokens uint64, remaining uint64, err error) {
	lock, err := r.lock(key)
	if err != nil {
		return 0, 0, err
	}

	defer lock.Release()

	bucket, err := r.getBucket(ctx, key)
	if err != nil {
		return 0, 0, err
	}
	if bucket == nil {
		return 0, 0, nil
	}

	return bucket.get()
}

// Set configures the Bucket-specific tokens and interval.
func (r *Store) Set(ctx context.Context, key string, tokens uint64, interval time.Duration) error {
	lock, err := r.lock(key)
	if err != nil {
		return err
	}

	defer lock.Release()

	bucket := NewBucket(tokens, interval)

	return r.putBucket(ctx, key, bucket)
}

// Burst adds the provided value to the Bucket's currently available tokens.
func (r *Store) Burst(ctx context.Context, key string, tokens uint64) error {
	lock, err := r.lock(key)
	if err != nil {
		return err
	}

	defer lock.Release()

	bucket, err := r.getBucket(ctx, key)
	if err != nil {
		return err
	}

	if bucket != nil {
		bucket.AvailableTokens = bucket.AvailableTokens + tokens
	} else {
		bucket = NewBucket(r.tokens+tokens, r.interval)
	}

	return r.putBucket(ctx, key, bucket)
}

func (r *Store) getBucket(ctx context.Context, key string) (*Bucket, error) {
	jsonData := r.cache.WithContext(ctx).GetString(key)
	if jsonData == "" {
		return nil, nil
	}

	bucket := &Bucket{}
	if err := r.json.UnmarshalString(jsonData, bucket); err != nil {
		return nil, err
	}

	return bucket, nil
}

func (r *Store) putBucket(ctx context.Context, key string, bucket *Bucket) error {
	jsonData, err := r.json.MarshalString(bucket)
	if err != nil {
		return err
	}

	return r.cache.WithContext(ctx).Put(key, jsonData, bucket.Interval)
}

func (r *Store) lock(key string) (cache.Lock, error) {
	lock := r.cache.Lock(key+":lock", r.interval)
	if !lock.BlockWithTicker(r.interval, 10*time.Millisecond) {
		return nil, errors.HttpRateLimitFailedToTakeToken
	}

	return lock, nil
}
