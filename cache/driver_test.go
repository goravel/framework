package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/cache"
	configmock "github.com/goravel/framework/contracts/config/mocks"
	logmock "github.com/goravel/framework/contracts/log/mocks"
)

type DriverTestSuite struct {
	suite.Suite
	driver     *DriverImpl
	mockConfig *configmock.Config
	mockLog    *logmock.Log
}

func TestDriverTestSuite(t *testing.T) {
	suite.Run(t, new(DriverTestSuite))
}

func (s *DriverTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	s.mockLog = &logmock.Log{}
	s.driver = NewDriverImpl(s.mockConfig)
}

func (s *DriverTestSuite) TestMemory() {
	s.mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()
	memory, err := s.driver.memory()
	s.NotNil(memory)
	s.Nil(err)
}

func (s *DriverTestSuite) TestCustom() {
	s.mockConfig.On("Get", "cache.stores.store.via").Return(&Store{}).Once()

	store, err := s.driver.custom("store")
	s.NotNil(store)
	s.Nil(err)
	s.Equal("name", store.Get("name", "Goravel").(string))

	s.mockConfig.AssertExpectations(s.T())
}

func (s *DriverTestSuite) TestStore() {
	s.mockConfig.On("GetString", "cache.stores.memory.driver").Return("memory").Once()
	s.mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()

	memory, err := NewApplication(s.mockConfig, s.mockLog, "memory")
	s.NotNil(memory)
	s.Nil(err)
	s.True(memory.Add("hello", "goravel", 5*time.Second))
	s.Equal("goravel", memory.GetString("hello"))

	s.mockConfig.On("GetString", "cache.stores.custom.driver").Return("custom").Once()
	s.mockConfig.On("Get", "cache.stores.custom.via").Return(&Store{}).Once()

	custom := memory.Store("custom")
	s.NotNil(custom)
	s.Equal("", custom.GetString("hello"))
	s.True(custom.Add("hello", "world", 5*time.Second))
	s.Equal("", custom.GetString("hello"))

	s.Equal("goravel", memory.GetString("hello"))

	s.mockConfig.AssertExpectations(s.T())
}

type Store struct {
}

//Add Store an item in the cache if the key does not exist.
func (r *Store) Add(key string, value any, seconds time.Duration) bool {
	return true
}

func (r *Store) Decrement(key string, value ...int) (int, error) {
	return 1, nil
}

//Forever Store an item in the cache indefinitely.
func (r *Store) Forever(key string, value any) bool {
	return true
}

//Forget Remove an item from the cache.
func (r *Store) Forget(key string) bool {
	return true
}

//Flush Remove all items from the cache.
func (r *Store) Flush() bool {
	return true
}

//Get Retrieve an item from the cache by key.
func (r *Store) Get(key string, def ...any) any {
	return key
}

//Get Retrieve an item from the cache by key.
func (r *Store) GetBool(key string, def ...bool) bool {
	return false
}

//Get Retrieve an item from the cache by key.
func (r *Store) GetInt(key string, def ...int) int {
	return 1
}

//Get Retrieve an item from the cache by key.
func (r *Store) GetInt64(key string, def ...int64) int64 {
	return 1
}

//Get Retrieve an item from the cache by key.
func (r *Store) GetString(key string, def ...string) string {
	return ""
}

//Has Check an item exists in the cache.
func (r *Store) Has(key string) bool {
	return true
}

func (r *Store) Increment(key string, value ...int) (int, error) {
	return 1, nil
}

func (r *Store) Lock(key string, second ...time.Duration) cache.Lock {
	return nil
}

//Pull Retrieve an item from the cache and delete it.
func (r *Store) Pull(key string, def ...any) any {
	return def
}

//Put Store an item in the cache for a given number of seconds.
func (r *Store) Put(key string, value any, seconds time.Duration) error {
	return nil
}

//Remember Get an item from the cache, or execute the given Closure and store the result.
func (r *Store) Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error) {
	return "", nil
}

//RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
func (r *Store) RememberForever(key string, callback func() (any, error)) (any, error) {
	return "", nil
}

func (r *Store) WithContext(ctx context.Context) cache.Driver {
	return r
}

var _ cache.Driver = &Store{}
