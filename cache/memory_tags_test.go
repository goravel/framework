package cache

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	contractscache "github.com/goravel/framework/contracts/cache"
	frameworkerrors "github.com/goravel/framework/errors"
)

type MemoryTagsTestSuite struct {
	suite.Suite
	memory *Memory
}

func TestMemoryTagsTestSuite(t *testing.T) {
	suite.Run(t, new(MemoryTagsTestSuite))
}

func (s *MemoryTagsTestSuite) SetupTest() {
	memory, err := getMemoryStore()
	s.Require().NoError(err)
	s.memory = memory
}

func (s *MemoryTagsTestSuite) tags(names ...string) contractscache.TaggedCache {
	tagged, err := s.memory.Tags(names...)
	s.Require().NoError(err)

	return tagged
}

func (s *MemoryTagsTestSuite) TestTagsRequired() {
	tagged, err := s.memory.Tags()
	s.Nil(tagged)
	s.ErrorIs(err, frameworkerrors.CacheTagsRequired)
}

func (s *MemoryTagsTestSuite) TestPutAndGet() {
	s.NoError(s.tags("people", "authors").Put("john", "John", time.Minute))

	s.Equal("John", s.tags("people", "authors").Get("john"))
	s.Equal("John", s.tags("people", "authors").GetString("john"))
	s.True(s.tags("people", "authors").Has("john"))

	// Items are namespaced by their tags, in order, like Laravel.
	s.Nil(s.memory.Get("john"))
	s.Nil(s.tags("people").Get("john"))
	s.Nil(s.tags("authors", "people").Get("john"))
	s.Equal("default", s.tags("people").Get("john", "default"))
}

func (s *MemoryTagsTestSuite) TestSameKeyInDifferentTags() {
	s.NoError(s.memory.Put("key", "plain", time.Minute))
	s.NoError(s.tags("a").Put("key", "a", time.Minute))
	s.NoError(s.tags("b").Put("key", "b", time.Minute))

	s.True(s.tags("a").Flush())

	s.Equal("plain", s.memory.Get("key"))
	s.Nil(s.tags("a").Get("key"))
	s.Equal("b", s.tags("b").Get("key"))
}

func (s *MemoryTagsTestSuite) TestFlushRemovesItemsOfAnyTag() {
	s.NoError(s.tags("people", "artists").Put("john", "John", time.Minute))
	s.True(s.tags("people", "authors").Forever("anne", "Anne"))
	s.True(s.tags("artists").Forever("bob", "Bob"))
	s.True(s.tags("authors").Forever("ann", "Ann"))

	s.True(s.tags("authors").Flush())

	s.Nil(s.tags("people", "authors").Get("anne"))
	s.Nil(s.tags("authors").Get("ann"))
	s.Equal("John", s.tags("people", "artists").Get("john"))
	s.Equal("Bob", s.tags("artists").Get("bob"))

	s.True(s.tags("people").Flush())

	s.Nil(s.tags("people", "artists").Get("john"))
	s.Equal("Bob", s.tags("artists").Get("bob"))
}

func (s *MemoryTagsTestSuite) TestFlushOfUnknownTag() {
	s.True(s.tags("missing").Flush())
}

func (s *MemoryTagsTestSuite) TestAdd() {
	s.True(s.tags("a").Add("key", "first", time.Minute))
	s.False(s.tags("a").Add("key", "second", time.Minute))
	s.Equal("first", s.tags("a").Get("key"))

	s.True(s.tags("a").Flush())
	s.True(s.tags("a").Add("key", "third", time.Minute))
	s.Equal("third", s.tags("a").Get("key"))
}

func (s *MemoryTagsTestSuite) TestIncrementAndDecrement() {
	res, err := s.tags("counters").Increment("visits")
	s.NoError(err)
	s.Equal(int64(1), res)

	res, err = s.tags("counters").Increment("visits", 4)
	s.NoError(err)
	s.Equal(int64(5), res)

	res, err = s.tags("counters").Decrement("visits", 2)
	s.NoError(err)
	s.Equal(int64(3), res)
	s.Equal(int64(3), s.tags("counters").GetInt64("visits"))
	s.Equal(3, s.tags("counters").GetInt("visits"))

	s.True(s.tags("counters").Flush())
	s.False(s.tags("counters").Has("visits"))

	res, err = s.tags("counters").Decrement("visits")
	s.NoError(err)
	s.Equal(int64(-1), res)
}

func (s *MemoryTagsTestSuite) TestForgetAndPull() {
	s.True(s.tags("a", "b").Forever("key", true))
	s.True(s.tags("a", "b").GetBool("key"))

	s.True(s.tags("a", "b").Forget("key"))
	s.False(s.tags("a", "b").Has("key"))
	s.Empty(s.memory.tags["a"])
	s.Empty(s.memory.tags["b"])

	s.True(s.tags("a").Forever("key", "value"))
	s.Equal("value", s.tags("a").Pull("key"))
	s.Nil(s.tags("a").Pull("key"))
	s.Equal("default", s.tags("a").Pull("key", "default"))
}

func (s *MemoryTagsTestSuite) TestExpiration() {
	s.NoError(s.tags("a").Put("key", "value", 50*time.Millisecond))
	s.True(s.tags("a").Has("key"))

	time.Sleep(100 * time.Millisecond)

	s.False(s.tags("a").Has("key"))
	s.Nil(s.tags("a").Get("key"))
}

func (s *MemoryTagsTestSuite) TestRemember() {
	calls := 0
	callback := func() (any, error) {
		calls++
		return "value", nil
	}

	val, err := s.tags("a").Remember("key", time.Minute, callback)
	s.NoError(err)
	s.Equal("value", val)

	val, err = s.tags("a").Remember("key", time.Minute, callback)
	s.NoError(err)
	s.Equal("value", val)
	s.Equal(1, calls)

	s.True(s.tags("a").Flush())

	val, err = s.tags("a").RememberForever("key", callback)
	s.NoError(err)
	s.Equal("value", val)
	s.Equal(2, calls)

	val, err = s.tags("b").Remember("key", time.Minute, func() (any, error) {
		return nil, errors.New("error")
	})
	s.EqualError(err, "error")
	s.Nil(val)
	s.False(s.tags("b").Has("key"))
}

func (s *MemoryTagsTestSuite) TestStoreFlushClearsTags() {
	s.True(s.tags("a").Forever("key", "value"))

	s.True(s.memory.Flush())

	s.Nil(s.tags("a").Get("key"))
	s.Empty(s.memory.tags)
}

func (s *MemoryTagsTestSuite) TestPruneTags() {
	s.NoError(s.tags("a", "b").Put("expired", "value", time.Millisecond))
	s.True(s.tags("a").Forever("kept", "value"))
	s.True(s.memory.Forget(tagNamespace([]string{"a"}) + ":kept"))
	s.True(s.tags("c").Forever("alive", "value"))

	time.Sleep(10 * time.Millisecond)
	s.memory.load(s.memory.key(tagNamespace([]string{"a", "b"}) + ":expired"))
	s.memory.pruneTags()

	s.NotContains(s.memory.tags, "a")
	s.NotContains(s.memory.tags, "b")
	s.Len(s.memory.tags["c"], 1)
}

func (s *MemoryTagsTestSuite) TestConcurrentPutAndFlush() {
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.NoError(s.tags("a", "b").Put("key", i, time.Minute))
		}()
		go func() {
			defer wg.Done()
			s.True(s.tags("b").Flush())
		}()
	}
	wg.Wait()

	// Whatever survived the race is still indexed, so a final flush removes it.
	s.True(s.tags("b").Flush())
	s.False(s.tags("a", "b").Has("key"))
}

func TestApplicationTags(t *testing.T) {
	memory, err := getMemoryStore()
	require.NoError(t, err)

	app := &Application{Driver: memory, store: "memory"}
	tagged, err := app.Tags("a")
	require.NoError(t, err)
	assert.NoError(t, tagged.Put("key", "value", time.Minute))
	assert.Equal(t, "value", tagged.Get("key"))

	app = &Application{Driver: driverWithoutTags{Driver: memory}, store: "custom"}
	tagged, err = app.Tags("a")
	assert.Nil(t, tagged)
	assert.ErrorIs(t, err, frameworkerrors.CacheTagsNotSupported)
	assert.EqualError(t, err, "cache store custom doesn't support tags")
}

// driverWithoutTags hides Tags to stand for a third-party driver that doesn't support tags.
type driverWithoutTags struct {
	contractscache.Driver
}
