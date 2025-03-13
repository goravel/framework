package limit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	cachemocks "github.com/goravel/framework/mocks/cache"
	"github.com/goravel/framework/support/debug"
)

type StoreTestSuite struct {
	suite.Suite
	store     *Store
	mockCache *cachemocks.Cache
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (s *StoreTestSuite) SetupTest() {
	impl, err := NewStore(s.mockCache, nil, 10, time.Second)
	s.NoError(err)
	s.store = impl
}

func (s *StoreTestSuite) TestStore_Take() {
	s.mockCache.On("Get", "testKey").Return(NewBucket(10, time.Minute)).Once()
	tokens, remaining, reset, ok, err := s.store.Take(context.Background(), "testKey")

	s.NoError(err)
	s.True(ok)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(9), remaining)
	s.NotZero(reset)
	s.mockCache.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestStore_Get() {
	s.mockCache.On("Get", "testKey").Return(NewBucket(10, time.Minute)).Once()
	tokens, remaining, err := s.store.Get(context.Background(), "testKey")

	s.NoError(err)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(10), remaining)
	s.mockCache.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestStore_Set() {
	s.mockCache.On("Get", "testKey").Return(NewBucket(5, time.Second)).Once()
	s.mockCache.On("Put", "testKey", mock.Anything, time.Second).Return(nil).Once()
	err := s.store.Set(context.Background(), "testKey", 5, time.Second)

	s.NoError(err)
	tokens, remaining, err := s.store.Get(context.Background(), "testKey")

	s.NoError(err)
	s.Equal(uint64(5), tokens)
	s.Equal(uint64(5), remaining)
	s.mockCache.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestStore_Burst() {
	s.mockCache.On("Get", "testKey").Return(NewBucket(10, time.Minute)).Twice()
	err := s.store.Burst(context.Background(), "testKey", 5)
	s.NoError(err)

	tokens, remaining, err := s.store.Get(context.Background(), "testKey")

	s.NoError(err)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(15), remaining)
	s.mockCache.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestBucket_NewBucket() {
	b := NewBucket(10, time.Second)
	s.NotNil(b)
	s.Equal(uint64(10), b.MaxTokens)
	s.Equal(uint64(10), b.AvailableTokens)
	s.Equal(time.Second, b.Interval)
}

func (s *StoreTestSuite) TestBucket_Get() {
	b := NewBucket(10, time.Second)
	tokens, remaining, err := b.get()

	s.NoError(err)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(10), remaining)
}

func (s *StoreTestSuite) TestBucket_Take_Success() {
	b := NewBucket(10, time.Second)
	tokens, remaining, reset, ok, err := b.take()

	s.NoError(err)
	s.True(ok)
	s.Equal(uint64(10), tokens)
	s.Equal(uint64(9), remaining)
	s.NotZero(reset)
}

func (s *StoreTestSuite) TestBucket_Take_Failure() {
	b := NewBucket(0, time.Second)
	tokens, remaining, reset, ok, err := b.take()

	s.NoError(err)
	s.False(ok)
	s.Equal(uint64(0), tokens)
	s.Equal(uint64(0), remaining)
	s.NotZero(reset)
}

func TestA(t *testing.T) {

	bucket := NewBucket(10, time.Second)
	data, err := json.Marshal(bucket)
	debug.Dump(string(data), err)
	assert.True(t, false)
}
