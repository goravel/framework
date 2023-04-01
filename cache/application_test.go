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
	stores      map[string]cache.Driver
	redisDocker *dockertest.Resource
}

func TestApplicationTestSuite(t *testing.T) {
	redisPool, redisDocker, redisStore, err := getRedisDocker()
	if err != nil {
		log.Fatalf("Get redis store error: %s", err)
	}
	memoryStore, err := getMemoryStore()
	if err != nil {
		log.Fatalf("Get memory store error: %s", err)
	}

	suite.Run(t, &ApplicationTestSuite{
		stores: map[string]cache.Driver{
			"redis":  redisStore,
			"memory": memoryStore,
		},
		redisDocker: redisDocker,
	})

	if err := redisPool.Purge(redisDocker); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func (s *ApplicationTestSuite) SetupTest() {
}

func (s *ApplicationTestSuite) TestInitMemory() {
	tests := []struct {
		description string
		setup       func()
	}{
		{
			description: "success",
			setup: func() {
				mockConfig := mock.Config()
				mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()

				s.NotNil(initMemory())

				mockConfig.AssertExpectations(s.T())
			},
		},
	}

	for _, test := range tests {
		s.Run(test.description, func() {
			test.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestInitRedis() {
	tests := []struct {
		description string
		setup       func()
	}{
		{
			description: "success",
			setup: func() {
				mockConfig := mock.Config()
				mockConfig.On("GetString", "cache.stores.redis.connection", "default").Return("default").Once()
				mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Once()
				mockConfig.On("GetString", "database.redis.default.port").Return(s.redisDocker.GetPort("6379/tcp")).Once()
				mockConfig.On("GetString", "database.redis.default.password").Return("").Once()
				mockConfig.On("GetInt", "database.redis.default.database").Return(0).Once()
				mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()

				s.NotNil(initRedis("redis"))

				mockConfig.AssertExpectations(s.T())
			},
		},
		{
			description: "error",
			setup: func() {
				mockConfig := mock.Config()
				mockConfig.On("GetString", "cache.stores.redis.connection", "default").Return("default").Once()
				mockConfig.On("GetString", "database.redis.default.host").Return("").Once()

				s.Nil(initRedis("redis"))

				mockConfig.AssertExpectations(s.T())
			},
		},
	}

	for _, test := range tests {
		s.Run(test.description, func() {
			test.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestInitCustom() {
	mockConfig := mock.Config()
	mockConfig.On("Get", "cache.stores.store.via").Return(&Store{}).Once()

	store := initCustom("store")
	s.NotNil(store)
	s.Equal("name", store.Get("name", "Goravel").(string))

	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestStore() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "cache.stores.memory.driver").Return("memory").Once()
	mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()

	memory := NewApplication("memory")
	s.NotNil(memory)
	s.True(memory.Add("hello", "goravel", 5*time.Second))
	s.Equal("goravel", memory.GetString("hello"))

	mockConfig.On("GetString", "cache.stores.redis.driver").Return("redis").Once()
	mockConfig.On("GetString", "cache.stores.redis.connection", "default").Return("default").Once()
	mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Once()
	mockConfig.On("GetString", "database.redis.default.port").Return(s.redisDocker.GetPort("6379/tcp")).Once()
	mockConfig.On("GetString", "database.redis.default.password").Return("").Once()
	mockConfig.On("GetInt", "database.redis.default.database").Return(0).Once()
	mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()

	redis := memory.Store("redis")
	s.NotNil(redis)
	s.Equal("", redis.GetString("hello"))
	s.True(redis.Add("hello", "world", 5*time.Second))
	s.Equal("world", redis.GetString("hello"))

	s.Equal("goravel", memory.GetString("hello"))

	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestAdd() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.Nil(store.Put("name", "Goravel", 1*time.Second))
			s.False(store.Add("name", "World", 1*time.Second))
			s.True(store.Add("name1", "World", 1*time.Second))
			s.True(store.Has("name1"))
			time.Sleep(2 * time.Second)
			s.False(store.Has("name1"))
			s.True(store.Flush())
		})
	}
}

func (s *ApplicationTestSuite) TestDecrement() {
	for name, store := range s.stores {
		s.Run(name, func() {
			res, err := store.Decrement("decrement")
			s.Equal(-1, res)
			s.Nil(err)

			s.Equal(-1, store.GetInt("decrement"))

			res, err = store.Decrement("decrement", 2)
			s.Equal(-3, res)
			s.Nil(err)

			res, err = store.Decrement("decrement1", 2)
			s.Equal(-2, res)
			s.Nil(err)

			s.Equal(-2, store.GetInt("decrement1"))

			s.True(store.Add("decrement2", 4, 2*time.Second))
			res, err = store.Decrement("decrement2")
			s.Equal(3, res)
			s.Nil(err)

			res, err = store.Decrement("decrement2", 2)
			s.Equal(1, res)
			s.Nil(err)
		})
	}
}

func (s *ApplicationTestSuite) TestForever() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.True(store.Forever("name", "Goravel"))
			s.Equal("Goravel", store.Get("name", "").(string))
			s.True(store.Flush())
		})
	}
}

func (s *ApplicationTestSuite) TestForget() {
	for name, store := range s.stores {
		s.Run(name, func() {
			val := store.Forget("test-forget")
			s.True(val)

			err := store.Put("test-forget", "goravel", 5*time.Second)
			s.Nil(err)
			s.True(store.Forget("test-forget"))
		})
	}
}

func (s *ApplicationTestSuite) TestFlush() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.Nil(store.Put("test-flush", "goravel", 5*time.Second))
			s.Equal("goravel", store.Get("test-flush", nil).(string))

			s.True(store.Flush())
			s.False(store.Has("test-flush"))
		})
	}
}

func (s *ApplicationTestSuite) TestGet() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.Nil(store.Put("name", "Goravel", 1*time.Second))
			s.Equal("Goravel", store.Get("name", "").(string))
			s.Equal("World", store.Get("name1", "World").(string))
			s.Equal("World1", store.Get("name2", func() any {
				return "World1"
			}).(string))
			s.True(store.Forget("name"))
			s.True(store.Flush())
		})
	}
}

func (s *ApplicationTestSuite) TestGetBool() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.Equal(true, store.GetBool("test-get-bool", true))
			s.Nil(store.Put("test-get-bool", true, 2*time.Second))
			s.Equal(true, store.GetBool("test-get-bool", false))
		})
	}
}

func (s *ApplicationTestSuite) TestGetInt() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.Equal(2, store.GetInt("test-get-int", 2))
			s.Nil(store.Put("test-get-int", 3, 2*time.Second))
			s.Equal(3, store.GetInt("test-get-int", 2))
		})
	}
}

func (s *ApplicationTestSuite) TestGetString() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.Equal("2", store.GetString("test-get-string", "2"))
			s.Nil(store.Put("test-get-string", "3", 2*time.Second))
			s.Equal("3", store.GetString("test-get-string", "2"))
		})
	}
}

func (s *ApplicationTestSuite) TestHas() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.False(store.Has("test-has"))
			s.Nil(store.Put("test-has", "goravel", 5*time.Second))
			s.True(store.Has("test-has"))
		})
	}
}

func (s *ApplicationTestSuite) TestIncrement() {
	for name, store := range s.stores {
		s.Run(name, func() {
			res, err := store.Increment("Increment")
			s.Equal(1, res)
			s.Nil(err)

			s.Equal(1, store.GetInt("Increment"))

			res, err = store.Increment("Increment", 2)
			s.Equal(3, res)
			s.Nil(err)

			res, err = store.Increment("Increment1", 2)
			s.Equal(2, res)
			s.Nil(err)

			s.Equal(2, store.GetInt("Increment1"))

			s.True(store.Add("Increment2", 1, 2*time.Second))
			res, err = store.Increment("Increment2")
			s.Equal(2, res)
			s.Nil(err)

			res, err = store.Increment("Increment2", 2)
			s.Equal(4, res)
			s.Nil(err)
		})
	}
}

func (s *ApplicationTestSuite) TestLock() {
	for _, store := range s.stores {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "once got lock, lock can't be got again",
				setup: func() {
					lock := store.Lock("lock")
					s.True(lock.Get())

					lock1 := store.Lock("lock")
					s.False(lock1.Get())

					lock.Release()
				},
			},
			{
				name: "lock can be got again when had been released",
				setup: func() {
					lock := store.Lock("lock")
					s.True(lock.Get())

					s.True(lock.Release())

					lock1 := store.Lock("lock")
					s.True(lock1.Get())

					s.True(lock1.Release())
				},
			},
			{
				name: "lock cannot be released when had been got",
				setup: func() {
					lock := store.Lock("lock")
					s.True(lock.Get())

					lock1 := store.Lock("lock")
					s.False(lock1.Get())
					s.False(lock1.Release())

					s.True(lock.Release())
				},
			},
			{
				name: "lock can be force released",
				setup: func() {
					lock := store.Lock("lock")
					s.True(lock.Get())

					lock1 := store.Lock("lock")
					s.False(lock1.Get())
					s.False(lock1.Release())
					s.True(lock1.ForceRelease())

					s.True(lock.Release())
				},
			},
			{
				name: "lock can be got again when timeout",
				setup: func() {
					lock := store.Lock("lock", 1*time.Second)
					s.True(lock.Get())

					time.Sleep(2 * time.Second)

					lock1 := store.Lock("lock")
					s.True(lock1.Get())
					s.True(lock1.Release())
				},
			},
			{
				name: "lock can be got again when had been released by callback",
				setup: func() {
					lock := store.Lock("lock")
					s.True(lock.Get(func() {
						s.True(true)
					}))

					lock1 := store.Lock("lock")
					s.True(lock1.Get())
					s.True(lock1.Release())
				},
			},
			{
				name: "block wait out",
				setup: func() {
					lock := store.Lock("lock")
					s.True(lock.Get())

					go func() {
						lock1 := store.Lock("lock")
						s.NotNil(lock1.Block(1 * time.Second))
					}()

					time.Sleep(2 * time.Second)

					lock.Release()
				},
			},
			{
				name: "get lock by block when just timeout",
				setup: func() {
					lock := store.Lock("lock")
					s.True(lock.Get())

					go func() {
						lock1 := store.Lock("lock")
						s.True(lock1.Block(2 * time.Second))
						s.True(lock1.Release())
					}()

					time.Sleep(1 * time.Second)

					lock.Release()

					time.Sleep(2 * time.Second)
				},
			},
			{
				name: "get lock by block",
				setup: func() {
					lock := store.Lock("lock")
					s.True(lock.Get())

					go func() {
						lock1 := store.Lock("lock")
						s.True(lock1.Block(3 * time.Second))
						s.True(lock1.Release())
					}()

					time.Sleep(1 * time.Second)

					lock.Release()

					time.Sleep(3 * time.Second)
				},
			},
			{
				name: "get lock by block with callback",
				setup: func() {
					lock := store.Lock("lock")
					s.True(lock.Get())

					go func() {
						lock1 := store.Lock("lock")
						s.True(lock1.Block(2*time.Second, func() {
							s.True(true)
						}))
					}()

					time.Sleep(1 * time.Second)

					lock.Release()

					time.Sleep(2 * time.Second)
				},
			},
		}

		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *ApplicationTestSuite) TestPull() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.Nil(store.Put("name", "Goravel", 1*time.Second))
			s.True(store.Has("name"))
			s.Equal("Goravel", store.Pull("name", "").(string))
			s.False(store.Has("name"))
		})
	}
}

func (s *ApplicationTestSuite) TestPut() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.Nil(store.Put("name", "Goravel", 1*time.Second))
			s.True(store.Has("name"))
			s.Equal("Goravel", store.Get("name", "").(string))
			time.Sleep(2 * time.Second)
			s.False(store.Has("name"))
		})
	}
}

func (s *ApplicationTestSuite) TestRemember() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.Nil(store.Put("name", "Goravel", 1*time.Second))
			value, err := store.Remember("name", 1*time.Second, func() any {
				return "World"
			})
			s.Nil(err)
			s.Equal("Goravel", value)

			value, err = store.Remember("name1", 1*time.Second, func() any {
				return "World1"
			})
			s.Nil(err)
			s.Equal("World1", value)
			time.Sleep(2 * time.Second)
			s.False(store.Has("name1"))
			s.True(store.Flush())
		})
	}
}

func (s *ApplicationTestSuite) TestRememberForever() {
	for name, store := range s.stores {
		s.Run(name, func() {
			s.Nil(store.Put("name", "Goravel", 1*time.Second))
			value, err := store.RememberForever("name", func() any {
				return "World"
			})
			s.Nil(err)
			s.Equal("Goravel", value)

			value, err = store.RememberForever("name1", func() any {
				return "World1"
			})
			s.Nil(err)
			s.Equal("World1", value)
			s.True(store.Flush())
		})
	}
}

func getRedisDocker() (*dockertest.Pool, *dockertest.Resource, cache.Driver, error) {
	pool, resource, err := testingdocker.Redis()
	if err != nil {
		return nil, nil, nil, err
	}

	var store cache.Driver
	if err := pool.Retry(func() error {
		var err error
		mockConfig := mock.Config()
		mockConfig.On("GetString", "cache.stores.redis.connection").Return("default").Once()
		mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Once()
		mockConfig.On("GetString", "database.redis.default.port").Return(resource.GetPort("6379/tcp")).Once()
		mockConfig.On("GetString", "database.redis.default.password").Return(resource.GetPort("")).Once()
		mockConfig.On("GetInt", "database.redis.default.database").Return(0).Once()
		mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()
		store, err = NewRedis(context.Background(), "default")

		return err
	}); err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, store, nil
}

func getMemoryStore() (*Memory, error) {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()

	memory, err := NewMemory()
	if err != nil {
		return nil, err
	}

	return memory, nil
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
func (r *Store) Remember(key string, ttl time.Duration, callback func() any) (any, error) {
	return "", nil
}

//RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
func (r *Store) RememberForever(key string, callback func() any) (any, error) {
	return "", nil
}

func (r *Store) WithContext(ctx context.Context) cache.Driver {
	return r
}

var _ cache.Driver = &Store{}
