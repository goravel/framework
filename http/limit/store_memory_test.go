package limit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)

	store := NewStore(memoryCache{memory}, json.New(), tokens, interval)

	allowed := make([]int, rounds)
	start := time.Now()
	for {
		sent := time.Since(start)
		round := int(sent / interval)
		if round >= rounds {
			break
		}

		_, _, _, ok, err := store.Take(context.Background(), "127.0.0.1")
		require.NoError(t, err)
		// The bucket counts the interval the request is served in, the loop the
		// one it was sent in. A request that crossed a boundary in between, for
		// instance while it waited for the lock, belongs to neither.
		if ok && round == int(time.Since(start)/interval) {
			allowed[round]++
		}

		time.Sleep(interval / 60)
	}

	for round, count := range allowed {
		assert.LessOrEqual(t, count, tokens, "interval %d let %d requests through, the limit is %d", round, count, tokens)
		// A limiter that refuses every request also stays under the limit.
		assert.NotZero(t, count, "interval %d let nothing through", round)
	}
}
