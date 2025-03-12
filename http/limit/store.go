package limit

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type Store struct {
	tokens   uint64
	interval time.Duration
	cache    cache.Cache
	json     foundation.Json
}

func NewStore(cache cache.Cache, json foundation.Json, tokens uint64, interval time.Duration) (*Store, error) {
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

	return s, nil
}

// Take attempts to remove a token from the named key. If the take is
// successful, it returns true, otherwise false. It also returns the configured
// limit, remaining tokens, and reset time.
func (s *Store) Take(ctx context.Context, key string) (uint64, uint64, uint64, bool, error) {
	lock, err := s.lock(key)
	if err != nil {
		return 0, 0, 0, false, err
	}

	defer lock.Release()

	bucket, err := s.getBucket(ctx, key)
	if err != nil {
		return 0, 0, 0, false, err
	}

	if bucket != nil {
		return bucket.take()
	}

	bucket, err = s.putBucket(ctx, key, s.tokens, s.interval)
	if err != nil {
		return 0, 0, 0, false, err
	}

	return bucket.take()
}

// Get retrieves the information about the key.
func (s *Store) Get(ctx context.Context, key string) (uint64, uint64, error) {
	lock, err := s.lock(key)
	if err != nil {
		return 0, 0, err
	}

	defer lock.Release()

	bucket, err := s.getBucket(ctx, key)
	if err != nil {
		return 0, 0, err
	}

	if bucket != nil {
		return bucket.get()
	}

	return 0, 0, nil
}

// Set configures the Bucket-specific tokens and interval.
func (s *Store) Set(ctx context.Context, key string, tokens uint64, interval time.Duration) error {
	lock, err := s.lock(key)
	if err != nil {
		return err
	}

	defer lock.Release()

	_, err = s.putBucket(ctx, key, tokens, interval)

	return err
}

// Burst adds the provided value to the Bucket's currently available tokens.
func (s *Store) Burst(ctx context.Context, key string, tokens uint64) error {
	lock, err := s.lock(key)
	if err != nil {
		return err
	}

	defer lock.Release()

	bucket, err := s.getBucket(ctx, key)
	if err != nil {
		return err
	}

	if bucket != nil {
		bucket.availableTokens = bucket.availableTokens + tokens
		return nil
	}

	_, err = s.putBucket(ctx, key, s.tokens+tokens, s.interval)

	return err
}

func (s *Store) getBucket(ctx context.Context, key string) (*Bucket, error) {
	jsonData := s.cache.WithContext(ctx).GetString(key)
	if jsonData == "" {
		return nil, nil
	}

	bucket := &Bucket{}
	if err := s.json.Unmarshal([]byte(jsonData), bucket); err != nil {
		return nil, err
	}

	return bucket, nil
}

func (s *Store) putBucket(ctx context.Context, key string, tokens uint64, interval time.Duration) (*Bucket, error) {
	bucket := NewBucket(tokens, interval)
	jsonData, err := s.json.Marshal(*bucket)
	if err != nil {
		return nil, err
	}

	if err := s.cache.WithContext(ctx).Put(key, jsonData, interval); err != nil {
		return nil, err
	}

	return bucket, nil
}

func (s *Store) lock(key string) (cache.Lock, error) {
	lock := s.cache.Lock(key, s.interval)
	if !lock.Block(s.interval) {
		return nil, errors.HttpRateLimitFailedToTakeToken
	}

	return lock, nil
}
