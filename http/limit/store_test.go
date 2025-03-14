package limit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	frameworkerrors "github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	cachemocks "github.com/goravel/framework/mocks/cache"
)

type StoreTestSuite struct {
	suite.Suite
	store      *Store
	mockCache  *cachemocks.Cache
	mockLock   *cachemocks.Lock
	json       foundation.Json
	ctx        context.Context
	testKey    string
	testBucket *Bucket
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (s *StoreTestSuite) SetupTest() {
	s.mockCache = cachemocks.NewCache(s.T())
	s.mockLock = cachemocks.NewLock(s.T())
	s.json = json.NewJson()
	s.store = NewStore(s.mockCache, s.json, 10, time.Second)
	s.ctx = context.Background()
	s.testKey = "testKey"
	s.testBucket = NewBucket(10, time.Minute)
}

func (s *StoreTestSuite) setupSuccessfulLock() {
	s.mockCache.EXPECT().Lock(s.testKey+":lock", time.Second).Return(s.mockLock).Once()
	s.mockLock.EXPECT().Block(time.Second).Return(true).Once()
	s.mockLock.EXPECT().Release().Return(true).Once()
}

func (s *StoreTestSuite) setupFailedLock() {
	s.mockCache.EXPECT().Lock(s.testKey+":lock", time.Second).Return(s.mockLock).Once()
	s.mockLock.EXPECT().Block(time.Second).Return(false).Once()
}

func (s *StoreTestSuite) setupSuccessfulGetBucket() {
	// Serialize a real bucket to ensure proper JSON structure
	bucket := NewBucket(10, time.Minute)
	jsonData, _ := s.json.Marshal(bucket)

	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().GetString(s.testKey).Return(string(jsonData)).Once()
}

func (s *StoreTestSuite) setupEmptyGetBucket() {
	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().GetString(s.testKey).Return("").Once()
}

func (s *StoreTestSuite) setupSuccessfulPutBucket(interval time.Duration) {
	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().Put(s.testKey, mock.AnythingOfType("string"), interval).Return(nil).Once()
}

func (s *StoreTestSuite) setupFailedPutBucket(interval time.Duration) {
	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().Put(s.testKey, mock.AnythingOfType("string"), interval).Return(errors.New("cache put error")).Once()
}

// Test constructor with valid values
func (s *StoreTestSuite) TestNewStore_ValidValues() {
	store := NewStore(s.mockCache, s.json, 10, time.Second)
	s.Equal(uint64(10), store.tokens)
	s.Equal(time.Second, store.interval)
	s.Equal(s.mockCache, store.cache)
	s.Equal(s.json, store.json)
}

// Test constructor with invalid values
func (s *StoreTestSuite) TestNewStore_InvalidValues() {
	// Test with zero tokens
	store := NewStore(s.mockCache, s.json, 0, time.Second)
	s.Equal(uint64(1), store.tokens, "Should default to 1 token when 0 is provided")

	// Test with negative interval
	store = NewStore(s.mockCache, s.json, 10, -time.Second)
	s.Equal(time.Second, store.interval, "Should default to 1 second when negative interval is provided")

	// Test with zero interval
	store = NewStore(s.mockCache, s.json, 10, 0)
	s.Equal(time.Second, store.interval, "Should default to 1 second when zero interval is provided")
}

// Test Take with successful flow
func (s *StoreTestSuite) TestStore_Take_Success() {
	s.setupSuccessfulLock()
	s.setupSuccessfulGetBucket()
	s.setupSuccessfulPutBucket(time.Minute)

	tokens, remaining, reset, ok, err := s.store.Take(s.ctx, s.testKey)

	s.NoError(err)
	s.True(ok)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(9), remaining)
	s.NotZero(reset)
}

// Test Take with lock failure
func (s *StoreTestSuite) TestStore_Take_LockFailure() {
	s.setupFailedLock()

	tokens, remaining, reset, ok, err := s.store.Take(s.ctx, s.testKey)

	s.Equal(frameworkerrors.HttpRateLimitFailedToTakeToken, err)
	s.False(ok)
	s.Equal(uint64(0), tokens)
	s.Equal(uint64(0), remaining)
	s.Equal(uint64(0), reset)
}

// Test Take with getBucket error
func (s *StoreTestSuite) TestStore_Take_GetBucketError() {
	s.setupSuccessfulLock()

	// Create a valid bucket structure but with invalid field values
	// This will prevent the divide by zero error in the tick function
	jsonData := `{"StartTime":1000,"MaxTokens":10,"Interval":"invalid","AvailableTokens":5,"LastTick":0}`
	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().GetString(s.testKey).Return(jsonData).Once()

	tokens, remaining, reset, ok, err := s.store.Take(s.ctx, s.testKey)

	s.Error(err)
	s.False(ok)
	s.Equal(uint64(0), tokens)
	s.Equal(uint64(0), remaining)
	s.Equal(uint64(0), reset)
}

// Test Take with putBucket error
func (s *StoreTestSuite) TestStore_Take_PutBucketError() {
	s.setupSuccessfulLock()
	s.setupSuccessfulGetBucket()
	s.setupFailedPutBucket(time.Minute)

	tokens, remaining, reset, ok, err := s.store.Take(s.ctx, s.testKey)

	s.Error(err)
	s.False(ok)
	s.Equal(uint64(0), tokens)
	s.Equal(uint64(0), remaining)
	s.Equal(uint64(0), reset)
}

// Test Take with empty bucket (cache miss)
func (s *StoreTestSuite) TestStore_Take_EmptyBucket() {
	s.setupSuccessfulLock()
	s.setupEmptyGetBucket()
	s.setupSuccessfulPutBucket(time.Second) // Default interval from store

	tokens, remaining, reset, ok, err := s.store.Take(s.ctx, s.testKey)

	s.NoError(err)
	s.True(ok)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(9), remaining)
	s.NotZero(reset)
}

// Test Take with no available tokens
func (s *StoreTestSuite) TestStore_Take_NoAvailableTokens() {
	s.setupSuccessfulLock()

	// Create a bucket with no available tokens
	bucket := NewBucket(10, time.Minute)
	bucket.AvailableTokens = 0
	jsonData, _ := s.json.Marshal(bucket)

	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().GetString(s.testKey).Return(string(jsonData)).Once()
	s.setupSuccessfulPutBucket(time.Minute)

	tokens, remaining, reset, ok, err := s.store.Take(s.ctx, s.testKey)

	s.NoError(err)
	s.False(ok, "Should return false when no tokens are available")
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(0), remaining)
	s.NotZero(reset)
}

// Test Get with successful flow
func (s *StoreTestSuite) TestStore_Get_Success() {
	s.setupSuccessfulLock()
	s.setupSuccessfulGetBucket()

	tokens, remaining, err := s.store.Get(s.ctx, s.testKey)

	s.NoError(err)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(10), remaining)
}

// Test Get with lock failure
func (s *StoreTestSuite) TestStore_Get_LockFailure() {
	s.setupFailedLock()

	tokens, remaining, err := s.store.Get(s.ctx, s.testKey)

	s.Equal(frameworkerrors.HttpRateLimitFailedToTakeToken, err)
	s.Equal(uint64(0), tokens)
	s.Equal(uint64(0), remaining)
}

// Test Get with getBucket error
func (s *StoreTestSuite) TestStore_Get_GetBucketError() {
	s.setupSuccessfulLock()

	// Completely invalid JSON that will definitely cause an unmarshal error
	jsonData := `{this is not valid JSON at all}`
	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().GetString(s.testKey).Return(jsonData).Once()

	tokens, remaining, err := s.store.Get(s.ctx, s.testKey)

	s.Error(err)
	s.Equal(uint64(0), tokens)
	s.Equal(uint64(0), remaining)
}

// Test Get with empty bucket (cache miss)
func (s *StoreTestSuite) TestStore_Get_EmptyBucket() {
	s.setupSuccessfulLock()
	s.setupEmptyGetBucket()

	tokens, remaining, err := s.store.Get(s.ctx, s.testKey)

	s.NoError(err)
	s.Equal(uint64(0), tokens)
	s.Equal(uint64(0), remaining)
}

// Test Set with successful flow
func (s *StoreTestSuite) TestStore_Set_Success() {
	s.setupSuccessfulLock()
	s.setupSuccessfulPutBucket(time.Second)

	err := s.store.Set(s.ctx, s.testKey, 5, time.Second)

	s.NoError(err)
}

// Test Set with lock failure
func (s *StoreTestSuite) TestStore_Set_LockFailure() {
	s.setupFailedLock()

	err := s.store.Set(s.ctx, s.testKey, 5, time.Second)

	s.Equal(frameworkerrors.HttpRateLimitFailedToTakeToken, err)
}

// Test Set with putBucket error
func (s *StoreTestSuite) TestStore_Set_PutBucketError() {
	s.setupSuccessfulLock()
	s.setupFailedPutBucket(time.Second)

	err := s.store.Set(s.ctx, s.testKey, 5, time.Second)

	s.Error(err)
}

// Test Burst with successful flow and existing bucket
func (s *StoreTestSuite) TestStore_Burst_SuccessWithExistingBucket() {
	s.setupSuccessfulLock()
	s.setupSuccessfulGetBucket()
	s.setupSuccessfulPutBucket(time.Minute)

	err := s.store.Burst(s.ctx, s.testKey, 5)

	s.NoError(err)
}

// Test Burst with successful flow and no existing bucket
func (s *StoreTestSuite) TestStore_Burst_SuccessWithNoExistingBucket() {
	s.setupSuccessfulLock()
	s.setupEmptyGetBucket()
	s.setupSuccessfulPutBucket(time.Second) // Default interval from store

	err := s.store.Burst(s.ctx, s.testKey, 5)

	s.NoError(err)
}

// Test Burst with lock failure
func (s *StoreTestSuite) TestStore_Burst_LockFailure() {
	s.setupFailedLock()

	err := s.store.Burst(s.ctx, s.testKey, 5)

	s.Equal(frameworkerrors.HttpRateLimitFailedToTakeToken, err)
}

// Test Burst with getBucket error
func (s *StoreTestSuite) TestStore_Burst_GetBucketError() {
	s.setupSuccessfulLock()

	// Completely invalid JSON that will definitely cause an unmarshal error
	jsonData := `{this is not valid JSON at all}`
	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().GetString(s.testKey).Return(jsonData).Once()

	// We don't need to mock the Put method because the function should return early with an error

	err := s.store.Burst(s.ctx, s.testKey, 5)

	s.Error(err)
}

// Test Burst with putBucket error
func (s *StoreTestSuite) TestStore_Burst_PutBucketError() {
	s.setupSuccessfulLock()
	s.setupSuccessfulGetBucket()
	s.setupFailedPutBucket(time.Minute)

	err := s.store.Burst(s.ctx, s.testKey, 5)

	s.Error(err)
}

// Test cache operations with context
func (s *StoreTestSuite) TestStore_CacheWithContext() {
	//nolint:all
	customCtx := context.WithValue(context.Background(), "key", "value")

	s.mockCache.EXPECT().Lock(s.testKey+":lock", time.Second).Return(s.mockLock).Once()
	s.mockLock.EXPECT().Block(time.Second).Return(true).Once()
	s.mockLock.EXPECT().Release().Return(true).Once()

	s.mockCache.EXPECT().WithContext(customCtx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().GetString(s.testKey).Return("").Once()

	s.mockCache.EXPECT().WithContext(customCtx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().Put(s.testKey, mock.AnythingOfType("string"), time.Second).Return(nil).Once()

	tokens, remaining, reset, ok, err := s.store.Take(customCtx, s.testKey)

	s.NoError(err)
	s.True(ok)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(9), remaining)
	s.NotZero(reset)
}

// Test with cache Put error
func (s *StoreTestSuite) TestStore_CachePutError() {
	s.setupSuccessfulLock()
	s.setupSuccessfulGetBucket()
	s.setupFailedPutBucket(time.Minute)

	tokens, remaining, reset, ok, err := s.store.Take(s.ctx, s.testKey)

	s.Error(err)
	s.False(ok)
	s.Equal(uint64(0), tokens)
	s.Equal(uint64(0), remaining)
	s.Equal(uint64(0), reset)
}

// Test with multiple operations
func (s *StoreTestSuite) TestStore_MultipleOperations() {
	// First take a token
	s.setupSuccessfulLock()
	s.setupSuccessfulGetBucket()
	s.setupSuccessfulPutBucket(time.Minute)

	tokens1, remaining1, reset1, ok1, err1 := s.store.Take(s.ctx, s.testKey)

	s.NoError(err1)
	s.True(ok1)
	s.Equal(uint64(10), tokens1)
	s.Equal(uint64(9), remaining1)
	s.NotZero(reset1)

	// Then get the bucket state
	s.setupSuccessfulLock()

	// Create a bucket with 9 tokens remaining to simulate the previous take
	bucket := NewBucket(10, time.Minute)
	bucket.AvailableTokens = 9
	jsonData, _ := s.json.Marshal(bucket)

	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().GetString(s.testKey).Return(string(jsonData)).Once()

	tokens2, remaining2, err2 := s.store.Get(s.ctx, s.testKey)

	s.NoError(err2)
	s.Equal(uint64(10), tokens2)
	s.Equal(uint64(9), remaining2)

	// Then burst the bucket
	s.setupSuccessfulLock()

	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().GetString(s.testKey).Return(string(jsonData)).Once()
	s.setupSuccessfulPutBucket(time.Minute)

	err3 := s.store.Burst(s.ctx, s.testKey, 5)

	s.NoError(err3)

	// Finally, get the bucket state again
	s.setupSuccessfulLock()

	// Create a bucket with 14 tokens remaining to simulate the previous burst
	bucket = NewBucket(10, time.Minute)
	bucket.AvailableTokens = 14
	jsonData, _ = s.json.Marshal(bucket)

	s.mockCache.EXPECT().WithContext(s.ctx).Return(s.mockCache).Once()
	s.mockCache.EXPECT().GetString(s.testKey).Return(string(jsonData)).Once()

	tokens4, remaining4, err4 := s.store.Get(s.ctx, s.testKey)

	s.NoError(err4)
	s.Equal(uint64(10), tokens4)
	s.Equal(uint64(14), remaining4)
}
