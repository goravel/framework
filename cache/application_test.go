package cache

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/goravel/framework/contracts/cache"
	testingdocker "github.com/goravel/framework/testing/docker"
	"github.com/goravel/framework/testing/mock"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type ApplicationTestSuite struct {
	suite.Suite
	stores      []cache.Store
	redisDocker *dockertest.Resource
}

func TestApplicationTestSuite(t *testing.T) {
	redisPool, redisDocker, redisStore, err := getRedisDocker()
	if err != nil {
		log.Fatalf("Get redis error: %s", err)
	}

	suite.Run(t, &ApplicationTestSuite{
		stores: []cache.Store{
			redisStore,
		},
		redisDocker: redisDocker,
	})

	if err := redisPool.Purge(redisDocker); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func (s *ApplicationTestSuite) SetupTest() {
}

func (s *ApplicationTestSuite) TestInitRedis() {
	tests := []struct {
		description string
		setup       func(description string)
	}{
		{
			description: "success",
			setup: func(description string) {
				mockConfig := mock.Config()
				mockConfig.On("GetString", "cache.default").Return("redis").Twice()
				mockConfig.On("GetString", "cache.stores.redis.driver").Return("redis").Once()
				mockConfig.On("GetString", "cache.stores.redis.connection").Return("default").Once()
				mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Once()
				mockConfig.On("GetString", "database.redis.default.port").Return(s.redisDocker.GetPort("6379/tcp")).Once()
				mockConfig.On("GetString", "database.redis.default.password").Return("").Once()
				mockConfig.On("GetInt", "database.redis.default.database").Return(0).Once()
				mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()

				app := Application{}
				s.NotNil(app.Init(), description)

				mockConfig.AssertExpectations(s.T())
			},
		},
		{
			description: "error",
			setup: func(description string) {
				mockConfig := mock.Config()
				mockConfig.On("GetString", "cache.default").Return("redis").Twice()
				mockConfig.On("GetString", "cache.stores.redis.driver").Return("redis").Once()
				mockConfig.On("GetString", "cache.stores.redis.connection").Return("default").Once()
				mockConfig.On("GetString", "database.redis.default.host").Return("").Once()

				app := Application{}
				s.Nil(app.Init(), description)

				mockConfig.AssertExpectations(s.T())
			},
		},
	}

	for _, test := range tests {
		test.setup(test.description)
	}
}

func (s *ApplicationTestSuite) TestAdd() {
	for _, store := range s.stores {
		s.Nil(store.Put("name", "Goravel", 1*time.Second))
		s.False(store.Add("name", "World", 1*time.Second))
		s.True(store.Add("name1", "World", 1*time.Second))
		s.True(store.Has("name1"))
		time.Sleep(2 * time.Second)
		s.False(store.Has("name1"))
		s.True(store.Flush())
	}
}

func (s *ApplicationTestSuite) TestForever() {
	for _, store := range s.stores {
		s.True(store.Forever("name", "Goravel"))
		s.Equal("Goravel", store.Get("name", "").(string))
		s.True(store.Flush())
	}
}

func (s *ApplicationTestSuite) TestForget() {
	for _, store := range s.stores {
		val := store.Forget("test-forget")
		s.True(val)

		err := store.Put("test-forget", "goravel", 5*time.Second)
		s.Nil(err)
		s.True(store.Forget("test-forget"))
	}
}

func (s *ApplicationTestSuite) TestFlush() {
	for _, store := range s.stores {
		s.Nil(store.Put("test-flush", "goravel", 5*time.Second))
		s.Equal("goravel", store.Get("test-flush", nil).(string))

		s.True(store.Flush())
		s.False(store.Has("test-flush"))
	}
}

func (s *ApplicationTestSuite) TestGet() {
	for _, store := range s.stores {
		s.Nil(store.Put("name", "Goravel", 1*time.Second))
		s.Equal("Goravel", store.Get("name", "").(string))
		s.Equal("World", store.Get("name1", "World").(string))
		s.Equal("World1", store.Get("name2", func() interface{} {
			return "World1"
		}).(string))
		s.True(store.Forget("name"))
		s.True(store.Flush())
	}
}

func (s *ApplicationTestSuite) TestGetBool() {
	for _, store := range s.stores {
		s.Equal(true, store.GetBool("test-get-bool", true))
		s.Nil(store.Put("test-get-bool", true, 2*time.Second))
		s.Equal(true, store.GetBool("test-get-bool", false))
	}
}

func (s *ApplicationTestSuite) TestGetInt() {
	for _, store := range s.stores {
		s.Equal(2, store.GetInt("test-get-int", 2))
		s.Nil(store.Put("test-get-int", 3, 2*time.Second))
		s.Equal(3, store.GetInt("test-get-int", 2))
	}
}

func (s *ApplicationTestSuite) TestGetString() {
	for _, store := range s.stores {
		s.Equal("2", store.GetString("test-get-string", "2"))
		s.Nil(store.Put("test-get-string", "3", 2*time.Second))
		s.Equal("3", store.GetString("test-get-string", "2"))
	}
}

func (s *ApplicationTestSuite) TestHas() {
	for _, store := range s.stores {
		s.False(store.Has("test-has"))
		s.Nil(store.Put("test-has", "goravel", 5*time.Second))
		s.True(store.Has("test-has"))
	}
}

func (s *ApplicationTestSuite) TestPull() {
	for _, store := range s.stores {
		s.Nil(store.Put("name", "Goravel", 1*time.Second))
		s.True(store.Has("name"))
		s.Equal("Goravel", store.Pull("name", "").(string))
		s.False(store.Has("name"))
	}
}

func (s *ApplicationTestSuite) TestPut() {
	for _, store := range s.stores {
		s.Nil(store.Put("name", "Goravel", 1*time.Second))
		s.True(store.Has("name"))
		s.Equal("Goravel", store.Get("name", "").(string))
		time.Sleep(2 * time.Second)
		s.False(store.Has("name"))
	}
}

func (s *ApplicationTestSuite) TestRemember() {
	for _, store := range s.stores {
		s.Nil(store.Put("name", "Goravel", 1*time.Second))
		value, err := store.Remember("name", 1*time.Second, func() interface{} {
			return "World"
		})
		s.Nil(err)
		s.Equal("Goravel", value)

		value, err = store.Remember("name1", 1*time.Second, func() interface{} {
			return "World1"
		})
		s.Nil(err)
		s.Equal("World1", value)
		time.Sleep(2 * time.Second)
		s.False(store.Has("name1"))
		s.True(store.Flush())
	}
}

func (s *ApplicationTestSuite) TestRememberForever() {
	for _, store := range s.stores {
		s.Nil(store.Put("name", "Goravel", 1*time.Second))
		value, err := store.RememberForever("name", func() interface{} {
			return "World"
		})
		s.Nil(err)
		s.Equal("Goravel", value)

		value, err = store.RememberForever("name1", func() interface{} {
			return "World1"
		})
		s.Nil(err)
		s.Equal("World1", value)
		s.True(store.Flush())
	}
}

func (s *ApplicationTestSuite) TestCustomDriver() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "cache.default").Return("store").Once()
	mockConfig.On("GetString", "cache.stores.store.driver").Return("custom").Once()
	mockConfig.On("Get", "cache.stores.store.via").Return(&Store{}).Once()

	app := Application{}
	store := app.Init()
	s.NotNil(store)
	s.Equal("Goravel", store.Get("name", "Goravel").(string))

	mockConfig.AssertExpectations(s.T())
}

func getRedisDocker() (*dockertest.Pool, *dockertest.Resource, cache.Store, error) {
	pool, resource, err := testingdocker.Redis()
	if err != nil {
		return nil, nil, nil, err
	}

	_ = resource.Expire(60)

	var store cache.Store
	if err := pool.Retry(func() error {
		var err error
		mockConfig := mock.Config()
		mockConfig.On("GetString", "cache.default").Return("redis").Once()
		mockConfig.On("GetString", "cache.stores.redis.connection").Return("default").Once()
		mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Once()
		mockConfig.On("GetString", "database.redis.default.port").Return(resource.GetPort("6379/tcp")).Once()
		mockConfig.On("GetString", "database.redis.default.password").Return(resource.GetPort("")).Once()
		mockConfig.On("GetInt", "database.redis.default.database").Return(0).Once()
		mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()
		store, err = NewRedis(context.Background())

		return err
	}); err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, store, nil
}

type Store struct {
}

func (r *Store) WithContext(ctx context.Context) cache.Store {
	return r
}

//Get Retrieve an item from the cache by key.
func (r *Store) Get(key string, def interface{}) interface{} {
	return def
}

//Get Retrieve an item from the cache by key.
func (r *Store) GetInt(key string, def int) int {
	return def
}

//Get Retrieve an item from the cache by key.
func (r *Store) GetBool(key string, def bool) bool {
	return def
}

//Get Retrieve an item from the cache by key.
func (r *Store) GetString(key string, def string) string {
	return def
}

//Has Check an item exists in the cache.
func (r *Store) Has(key string) bool {
	return true
}

//Put Store an item in the cache for a given number of seconds.
func (r *Store) Put(key string, value interface{}, seconds time.Duration) error {
	return nil
}

//Pull Retrieve an item from the cache and delete it.
func (r *Store) Pull(key string, def interface{}) interface{} {
	return def
}

//Add Store an item in the cache if the key does not exist.
func (r *Store) Add(key string, value interface{}, seconds time.Duration) bool {
	return true
}

//Remember Get an item from the cache, or execute the given Closure and store the result.
func (r *Store) Remember(key string, ttl time.Duration, callback func() interface{}) (interface{}, error) {
	return "", nil
}

//RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
func (r *Store) RememberForever(key string, callback func() interface{}) (interface{}, error) {
	return "", nil
}

//Forever Store an item in the cache indefinitely.
func (r *Store) Forever(key string, value interface{}) bool {
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

var _ cache.Store = &Store{}
