package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
)

type MemoryTestSuite struct {
	suite.Suite
	mockConfig *configmock.Config
	memory     *Memory
}

func TestMemoryTestSuite(t *testing.T) {
	suite.Run(t, new(MemoryTestSuite))
}

func (s *MemoryTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	memoryStore, err := getMemoryStore()
	s.Nil(err)
	s.memory = memoryStore
}

func (s *MemoryTestSuite) TestAdd() {
	s.Nil(s.memory.Put("name", "Goravel", 1*time.Second))
	s.False(s.memory.Add("name", "World", 1*time.Second))
	s.True(s.memory.Add("name1", "World", 1*time.Second))
	s.True(s.memory.Has("name1"))
	time.Sleep(2 * time.Second)
	s.False(s.memory.Has("name1"))
	s.True(s.memory.Flush())
}

func (s *MemoryTestSuite) TestDecrement() {
	res, err := s.memory.Decrement("decrement")
	s.Equal(-1, res)
	s.Nil(err)

	s.Equal(-1, s.memory.GetInt("decrement"))

	res, err = s.memory.Decrement("decrement", 2)
	s.Equal(-3, res)
	s.Nil(err)

	res, err = s.memory.Decrement("decrement1", 2)
	s.Equal(-2, res)
	s.Nil(err)

	s.Equal(-2, s.memory.GetInt("decrement1"))

	s.True(s.memory.Add("decrement2", 4, 2*time.Second))
	res, err = s.memory.Decrement("decrement2")
	s.Equal(3, res)
	s.Nil(err)

	res, err = s.memory.Decrement("decrement2", 2)
	s.Equal(1, res)
	s.Nil(err)
}

func (s *MemoryTestSuite) TestForever() {
	s.True(s.memory.Forever("name", "Goravel"))
	s.Equal("Goravel", s.memory.Get("name", "").(string))
	s.True(s.memory.Flush())
}

func (s *MemoryTestSuite) TestForget() {
	val := s.memory.Forget("test-forget")
	s.True(val)

	err := s.memory.Put("test-forget", "goravel", 5*time.Second)
	s.Nil(err)
	s.True(s.memory.Forget("test-forget"))
}

func (s *MemoryTestSuite) TestFlush() {
	s.Nil(s.memory.Put("test-flush", "goravel", 5*time.Second))
	s.Equal("goravel", s.memory.Get("test-flush", nil).(string))

	s.True(s.memory.Flush())
	s.False(s.memory.Has("test-flush"))
}

func (s *MemoryTestSuite) TestGet() {
	s.Nil(s.memory.Put("name", "Goravel", 1*time.Second))
	s.Equal("Goravel", s.memory.Get("name", "").(string))
	s.Equal("World", s.memory.Get("name1", "World").(string))
	s.Equal("World1", s.memory.Get("name2", func() any {
		return "World1"
	}).(string))
	s.True(s.memory.Forget("name"))
	s.True(s.memory.Flush())
}

func (s *MemoryTestSuite) TestGetBool() {
	s.Equal(true, s.memory.GetBool("test-get-bool", true))
	s.Nil(s.memory.Put("test-get-bool", true, 2*time.Second))
	s.Equal(true, s.memory.GetBool("test-get-bool", false))
}

func (s *MemoryTestSuite) TestGetInt() {
	s.Equal(2, s.memory.GetInt("test-get-int", 2))
	s.Nil(s.memory.Put("test-get-int", 3, 2*time.Second))
	s.Equal(3, s.memory.GetInt("test-get-int", 2))
}

func (s *MemoryTestSuite) TestGetString() {
	s.Equal("2", s.memory.GetString("test-get-string", "2"))
	s.Nil(s.memory.Put("test-get-string", "3", 2*time.Second))
	s.Equal("3", s.memory.GetString("test-get-string", "2"))
}

func (s *MemoryTestSuite) TestHas() {
	s.False(s.memory.Has("test-has"))
	s.Nil(s.memory.Put("test-has", "goravel", 5*time.Second))
	s.True(s.memory.Has("test-has"))
}

func (s *MemoryTestSuite) TestIncrement() {
	res, err := s.memory.Increment("Increment")
	s.Equal(1, res)
	s.Nil(err)

	s.Equal(1, s.memory.GetInt("Increment"))

	res, err = s.memory.Increment("Increment", 2)
	s.Equal(3, res)
	s.Nil(err)

	res, err = s.memory.Increment("Increment1", 2)
	s.Equal(2, res)
	s.Nil(err)

	s.Equal(2, s.memory.GetInt("Increment1"))

	s.True(s.memory.Add("Increment2", 1, 2*time.Second))
	res, err = s.memory.Increment("Increment2")
	s.Equal(2, res)
	s.Nil(err)

	res, err = s.memory.Increment("Increment2", 2)
	s.Equal(4, res)
	s.Nil(err)
}

func (s *MemoryTestSuite) TestLock() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "once got lock, lock can't be got again",
			setup: func() {
				lock := s.memory.Lock("lock")
				s.True(lock.Get())

				lock1 := s.memory.Lock("lock")
				s.False(lock1.Get())

				lock.Release()
			},
		},
		{
			name: "lock can be got again when had been released",
			setup: func() {
				lock := s.memory.Lock("lock")
				s.True(lock.Get())

				s.True(lock.Release())

				lock1 := s.memory.Lock("lock")
				s.True(lock1.Get())

				s.True(lock1.Release())
			},
		},
		{
			name: "lock cannot be released when had been got",
			setup: func() {
				lock := s.memory.Lock("lock")
				s.True(lock.Get())

				lock1 := s.memory.Lock("lock")
				s.False(lock1.Get())
				s.False(lock1.Release())

				s.True(lock.Release())
			},
		},
		{
			name: "lock can be force released",
			setup: func() {
				lock := s.memory.Lock("lock")
				s.True(lock.Get())

				lock1 := s.memory.Lock("lock")
				s.False(lock1.Get())
				s.False(lock1.Release())
				s.True(lock1.ForceRelease())

				s.True(lock.Release())
			},
		},
		{
			name: "lock can be got again when timeout",
			setup: func() {
				lock := s.memory.Lock("lock", 1*time.Second)
				s.True(lock.Get())

				time.Sleep(2 * time.Second)

				lock1 := s.memory.Lock("lock")
				s.True(lock1.Get())
				s.True(lock1.Release())
			},
		},
		{
			name: "lock can be got again when had been released by callback",
			setup: func() {
				lock := s.memory.Lock("lock")
				s.True(lock.Get(func() {
					s.True(true)
				}))

				lock1 := s.memory.Lock("lock")
				s.True(lock1.Get())
				s.True(lock1.Release())
			},
		},
		{
			name: "block wait out",
			setup: func() {
				lock := s.memory.Lock("lock")
				s.True(lock.Get())

				go func() {
					lock1 := s.memory.Lock("lock")
					s.NotNil(lock1.Block(1 * time.Second))
				}()

				time.Sleep(2 * time.Second)

				lock.Release()
			},
		},
		{
			name: "get lock by block when just timeout",
			setup: func() {
				lock := s.memory.Lock("lock")
				s.True(lock.Get())

				go func() {
					lock1 := s.memory.Lock("lock")
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
				lock := s.memory.Lock("lock")
				s.True(lock.Get())

				go func() {
					lock1 := s.memory.Lock("lock")
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
				lock := s.memory.Lock("lock")
				s.True(lock.Get())

				go func() {
					lock1 := s.memory.Lock("lock")
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

func (s *MemoryTestSuite) TestPull() {
	s.Nil(s.memory.Put("name", "Goravel", 1*time.Second))
	s.True(s.memory.Has("name"))
	s.Equal("Goravel", s.memory.Pull("name", "").(string))
	s.False(s.memory.Has("name"))
}

func (s *MemoryTestSuite) TestPut() {
	s.Nil(s.memory.Put("name", "Goravel", 1*time.Second))
	s.True(s.memory.Has("name"))
	s.Equal("Goravel", s.memory.Get("name", "").(string))
	time.Sleep(2 * time.Second)
	s.False(s.memory.Has("name"))
}

func (s *MemoryTestSuite) TestRemember() {
	s.Nil(s.memory.Put("name", "Goravel", 1*time.Second))
	value, err := s.memory.Remember("name", 1*time.Second, func() (any, error) {
		return "World", nil
	})
	s.Nil(err)
	s.Equal("Goravel", value)

	value, err = s.memory.Remember("name1", 1*time.Second, func() (any, error) {
		return "World1", nil
	})
	s.Nil(err)
	s.Equal("World1", value)
	time.Sleep(2 * time.Second)
	s.False(s.memory.Has("name1"))
	s.True(s.memory.Flush())

	value, err = s.memory.Remember("name2", 1*time.Second, func() (any, error) {
		return nil, errors.New("error")
	})
	s.EqualError(err, "error")
	s.Nil(value)
}

func (s *MemoryTestSuite) TestRememberForever() {
	s.Nil(s.memory.Put("name", "Goravel", 1*time.Second))
	value, err := s.memory.RememberForever("name", func() (any, error) {
		return "World", nil
	})
	s.Nil(err)
	s.Equal("Goravel", value)

	value, err = s.memory.RememberForever("name1", func() (any, error) {
		return "World1", nil
	})
	s.Nil(err)
	s.Equal("World1", value)
	s.True(s.memory.Flush())

	value, err = s.memory.RememberForever("name2", func() (any, error) {
		return nil, errors.New("error")
	})
	s.EqualError(err, "error")
	s.Nil(value)
}

func getMemoryStore() (*Memory, error) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "cache.prefix").Return("goravel_cache").Once()

	memory, err := NewMemory(mockConfig)
	if err != nil {
		return nil, err
	}

	return memory, nil
}
