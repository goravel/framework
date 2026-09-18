package limit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/cache"
	contractscache "github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/foundation/json"
	configmock "github.com/goravel/framework/mocks/config"
)

// memoryCache is the memory driver seen as a cache.Cache, the way the limiter
// receives it from the container.
type memoryCache struct {
	*cache.Memory
}

func (r memoryCache) Store(string) contractscache.Driver {
	return r
}

// TestStoreWithMemoryKeepsLimiting runs the limiter against the real memory
// driver. The bucket is stored with an expiration and the lock is taken on
// every request, so an expiration that outlives the value it belongs to lets
// requests through: the bucket is dropped mid-interval and the next request
// starts from a full one.
func TestStoreWithMemoryKeepsLimiting(t *testing.T) {
	const (
		tokens   = 20
		interval = 300 * time.Millisecond
		rounds   = 4
	)

	mockConfig := configmock.NewConfig(t)
	mockConfig.EXPECT().GetString("cache.prefix").Return("goravel_cache")
	memory, err := cache.NewMemory(mockConfig)
	assert.Nil(t, err)

	store := NewStore(memoryCache{memory}, json.New(), tokens, interval)

	allowed := make([]int, rounds)
	start := time.Now()
	for {
		round := int(time.Since(start) / interval)
		if round >= rounds {
			break
		}

		_, _, _, ok, err := store.Take(context.Background(), "127.0.0.1")
		assert.Nil(t, err)
		if ok {
			allowed[round]++
		}

		time.Sleep(interval / 60)
	}

	for round, count := range allowed {
		assert.LessOrEqual(t, count, tokens, "interval %d let %d requests through, the limit is %d", round, count, tokens)
	}
}
