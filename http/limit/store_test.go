package limit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/http"
	cachemocks "github.com/goravel/framework/mocks/cache"
)

type StoreTestSuite struct {
	suite.Suite
	store     Store
	mockCache *cachemocks.Cache
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (s *StoreTestSuite) SetupTest() {
	mockCache := &cachemocks.Cache{}
	s.mockCache = mockCache
	http.CacheFacade = mockCache

	impl, err := NewStore(10, time.Second)
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
	s.Equal(uint64(10), b.maxTokens)
	s.Equal(uint64(10), b.availableTokens)
	s.Equal(time.Second, b.interval)
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
