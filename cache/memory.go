package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cast"

	contractscache "github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/testing/docker"
	"github.com/goravel/framework/errors"
)

type Memory struct {
	ctx       context.Context
	prefix    string
	instance  sync.Map
	lastSweep atomic.Int64
}

// sweepInterval bounds how often a write may start a sweep of expired items.
const sweepInterval = time.Minute

func NewMemory(config config.Config) (*Memory, error) {
	memory := &Memory{
		prefix: prefix(config),
	}
	// Start the sweep clock now: a cache that has just been built has nothing
	// to reclaim.
	memory.lastSweep.Store(time.Now().UnixNano())

	return memory, nil
}

// Add an item in the cache if the key does not exist.
func (r *Memory) Add(key string, value any, t time.Duration) bool {
	r.sweepExpired()

	k := r.key(key)

	// A key that is already taken is the common outcome on the contended path:
	// Lock.Get is built on Add and polls it every 10ms. Answer it before
	// building an item that would be thrown away.
	if stored, loaded := r.instance.Load(k); loaded {
		if cached, ok := stored.(*item); ok && !cached.expired() {
			return false
		}
	}

	fresh := newItem(value, t)
	for {
		stored, loaded := r.instance.LoadOrStore(k, fresh)
		if !loaded {
			return true
		}

		// The key is taken. An expired item does not count as present, but it
		// has to be replaced atomically: another goroutine may be doing the
		// same, and only one of us may win.
		if cached, ok := stored.(*item); ok && !cached.expired() {
			return false
		}
		if r.instance.CompareAndSwap(k, stored, fresh) {
			return true
		}
	}
}

// Decrement decrements the value of an item in the cache.
func (r *Memory) Decrement(key string, value ...int64) (int64, error) {
	if len(value) == 0 {
		value = append(value, 1)
	}

	r.Add(key, new(int64), NoExpiration)
	pv := r.Get(key)
	switch nv := pv.(type) {
	case *atomic.Int64:
		return nv.Add(-value[0]), nil
	case *atomic.Int32:
		return int64(nv.Add(int32(-value[0]))), nil
	case *int64:
		return atomic.AddInt64(nv, -value[0]), nil
	case *int32:
		return int64(atomic.AddInt32(nv, int32(-value[0]))), nil
	default:
		return 0, errors.CacheMemoryInvalidIntValueType.Args(key)
	}
}

func (r *Memory) Docker() (docker.CacheDriver, error) {
	return nil, errors.CacheMemoryDriverNotSupportDocker
}

// Forever Put an item in the cache indefinitely.
func (r *Memory) Forever(key string, value any) bool {
	if err := r.Put(key, value, NoExpiration); err != nil {
		return false
	}

	return true
}

// Forget Remove an item from the cache.
func (r *Memory) Forget(key string) bool {
	r.instance.Delete(r.key(key))

	return true
}

// Flush Remove all items from the cache.
// Clear in place: replacing the sync.Map would swap out the mutex held by concurrent readers.
func (r *Memory) Flush() bool {
	r.instance.Clear()
	return true
}

// Get Retrieve an item from the cache by key.
func (r *Memory) Get(key string, def ...any) any {
	val, exist := r.load(r.key(key))
	if exist {
		return val
	}
	if len(def) == 0 {
		return nil
	}

	switch s := def[0].(type) {
	case func() any:
		return s()
	default:
		return s
	}
}

func (r *Memory) GetBool(key string, def ...bool) bool {
	if len(def) == 0 {
		def = append(def, false)
	}
	res := r.Get(key, def[0])

	return cast.ToBool(res)
}

func (r *Memory) GetInt(key string, def ...int) int {
	if len(def) == 0 {
		def = append(def, 0)
	}

	return cast.ToInt(r.Get(key, def[0]))
}

func (r *Memory) GetInt64(key string, def ...int64) int64 {
	if len(def) == 0 {
		def = append(def, 0)
	}

	return cast.ToInt64(r.Get(key, def[0]))
}

func (r *Memory) GetString(key string, def ...string) string {
	if len(def) == 0 {
		def = append(def, "")
	}

	return cast.ToString(r.Get(key, def[0]))
}

// Has Checks an item exists in the cache.
func (r *Memory) Has(key string) bool {
	_, exist := r.load(r.key(key))
	return exist
}

func (r *Memory) Increment(key string, value ...int64) (int64, error) {
	if len(value) == 0 {
		value = append(value, 1)
	}

	r.Add(key, new(int64), NoExpiration)
	pv := r.Get(key)
	switch nv := pv.(type) {
	case *atomic.Int64:
		return nv.Add(value[0]), nil
	case *atomic.Int32:
		return int64(nv.Add(int32(value[0]))), nil
	case *int64:
		return atomic.AddInt64(nv, value[0]), nil
	case *int32:
		return int64(atomic.AddInt32(nv, int32(value[0]))), nil
	default:
		return 0, errors.CacheMemoryInvalidIntValueType.Args(key)
	}
}

func (r *Memory) Lock(key string, t ...time.Duration) contractscache.Lock {
	return NewLock(r, key, t...)
}

// Pull Retrieve an item from the cache and delete it.
func (r *Memory) Pull(key string, def ...any) any {
	var res any
	if len(def) == 0 {
		res = r.Get(key)
	} else {
		res = r.Get(key, def[0])
	}
	r.Forget(key)

	return res
}

// Put an item in the cache for a given number of seconds.
func (r *Memory) Put(key string, value any, t time.Duration) error {
	r.sweepExpired()
	r.instance.Store(r.key(key), newItem(value, t))
	return nil
}

// Remember Get an item from the cache, or execute the given Closure and store the result.
func (r *Memory) Remember(key string, seconds time.Duration, callback func() (any, error)) (any, error) {
	val := r.Get(key, nil)
	if val != nil {
		return val, nil
	}

	var err error
	val, err = callback()
	if err != nil {
		return nil, err
	}

	if err := r.Put(key, val, seconds); err != nil {
		return nil, err
	}

	return val, nil
}

// RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
func (r *Memory) RememberForever(key string, callback func() (any, error)) (any, error) {
	val := r.Get(key, nil)
	if val != nil {
		return val, nil
	}

	var err error
	val, err = callback()
	if err != nil {
		return nil, err
	}

	if err = r.Put(key, val, NoExpiration); err != nil {
		return nil, err
	}

	return val, nil
}

func (r *Memory) WithContext(ctx context.Context) contractscache.Driver {
	r.ctx = ctx

	return r
}

// item is a cached value together with the moment it expires. The expiry is
// stored with the value instead of being scheduled with a timer: a timer only
// knows the key, so it deletes whatever sits under it when it fires, even a
// value stored after the timer was armed.
type item struct {
	value    any
	expireAt time.Time
}

func newItem(value any, t time.Duration) *item {
	res := &item{value: value}
	if t != NoExpiration {
		res.expireAt = time.Now().Add(t)
	}

	return res
}

func (i *item) expired() bool {
	return !i.expireAt.IsZero() && !time.Now().Before(i.expireAt)
}

// load returns the value stored under key, dropping it first if it has expired.
func (r *Memory) load(key string) (any, bool) {
	stored, exist := r.instance.Load(key)
	if !exist {
		return nil, false
	}

	cached, ok := stored.(*item)
	if !ok || cached.expired() {
		// Delete this item only, a fresh one may have taken its place already.
		r.instance.CompareAndDelete(key, stored)
		return nil, false
	}

	return cached.value, true
}

// sweepExpired drops every item that has expired. Reading or writing a key
// reclaims the item under it, but a key written once and never touched again
// would hold its value for the life of the process, so a write sweeps the whole
// map once every sweepInterval. Memory has no shutdown hook to hang a janitor
// goroutine on; ranging the map is O(n), so the sweep runs in the background
// rather than on the write that happened to reach the interval.
func (r *Memory) sweepExpired() {
	last := r.lastSweep.Load()
	now := time.Now().UnixNano()
	if now-last < int64(sweepInterval) || !r.lastSweep.CompareAndSwap(last, now) {
		return
	}

	go r.instance.Range(func(key, stored any) bool {
		if cached, ok := stored.(*item); !ok || cached.expired() {
			// Delete this item only, a fresh one may have taken its place already.
			r.instance.CompareAndDelete(key, stored)
		}

		return true
	})
}

func (r *Memory) key(key string) string {
	return r.prefix + key
}
