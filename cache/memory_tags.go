package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"slices"
	"strconv"
	"time"

	contractscache "github.com/goravel/framework/contracts/cache"
)

// Tags begins a tagged cache operation. Like Laravel, an item is namespaced by
// its tags, so it is only reachable through the same tags in the same order,
// while flushing any one of the tags removes it.
func (r *Memory) Tags(names ...string) contractscache.Repository {
	return &memoryTaggedCache{
		memory:    r,
		names:     slices.Clone(names),
		namespace: tagNamespace(names),
	}
}

// pruneTags drops keys whose item is gone (expired or removed) from the tag index.
func (r *Memory) pruneTags() {
	r.tagsMu.Lock()
	defer r.tagsMu.Unlock()

	for name, keys := range r.tags {
		for key := range keys {
			if _, exist := r.instance.Load(r.key(key)); !exist {
				delete(keys, key)
			}
		}
		if len(keys) == 0 {
			delete(r.tags, name)
		}
	}
}

// tagNamespace builds the key namespace of a tag set. Every name is prefixed by
// its length, so no two different tag lists hash the same input.
func tagNamespace(names []string) string {
	hash := sha1.New()
	for _, name := range names {
		hash.Write([]byte(strconv.Itoa(len(name)) + ":" + name))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// memoryTaggedCache is the tagged view of Memory returned by Tags.
type memoryTaggedCache struct {
	memory    *Memory
	names     []string
	namespace string
}

// Add an item in the cache if the key does not exist.
func (r *memoryTaggedCache) Add(key string, value any, t time.Duration) bool {
	r.memory.tagsMu.Lock()
	defer r.memory.tagsMu.Unlock()

	itemKey := r.itemKey(key)
	if !r.memory.Add(itemKey, value, t) {
		return false
	}
	r.register(itemKey)

	return true
}

// Decrement decrements the value of an item in the cache.
func (r *memoryTaggedCache) Decrement(key string, value ...int64) (int64, error) {
	r.memory.tagsMu.Lock()
	defer r.memory.tagsMu.Unlock()

	itemKey := r.itemKey(key)
	r.register(itemKey)

	return r.memory.Decrement(itemKey, value...)
}

// Forever add an item in the cache indefinitely.
func (r *memoryTaggedCache) Forever(key string, value any) bool {
	return r.Put(key, value, NoExpiration) == nil
}

// Forget removes an item from the cache.
func (r *memoryTaggedCache) Forget(key string) bool {
	r.memory.tagsMu.Lock()
	defer r.memory.tagsMu.Unlock()

	itemKey := r.itemKey(key)
	for _, name := range r.names {
		delete(r.memory.tags[name], itemKey)
	}

	return r.memory.Forget(itemKey)
}

// Flush removes every item stored under any of the tags.
func (r *memoryTaggedCache) Flush() bool {
	r.memory.tagsMu.Lock()
	defer r.memory.tagsMu.Unlock()

	for _, name := range r.names {
		for itemKey := range r.memory.tags[name] {
			r.memory.Forget(itemKey)
		}
		delete(r.memory.tags, name)
	}

	return true
}

// Get retrieve an item from the cache by key.
func (r *memoryTaggedCache) Get(key string, def ...any) any {
	return r.memory.Get(r.itemKey(key), def...)
}

// GetBool retrieves an item from the cache by key as a boolean.
func (r *memoryTaggedCache) GetBool(key string, def ...bool) bool {
	return r.memory.GetBool(r.itemKey(key), def...)
}

// GetInt retrieves an item from the cache by key as an integer.
func (r *memoryTaggedCache) GetInt(key string, def ...int) int {
	return r.memory.GetInt(r.itemKey(key), def...)
}

// GetInt64 retrieves an item from the cache by key as a 64-bit integer.
func (r *memoryTaggedCache) GetInt64(key string, def ...int64) int64 {
	return r.memory.GetInt64(r.itemKey(key), def...)
}

// GetString retrieves an item from the cache by key as a string.
func (r *memoryTaggedCache) GetString(key string, def ...string) string {
	return r.memory.GetString(r.itemKey(key), def...)
}

// Has check an item exists in the cache.
func (r *memoryTaggedCache) Has(key string) bool {
	return r.memory.Has(r.itemKey(key))
}

// Increment increments the value of an item in the cache.
func (r *memoryTaggedCache) Increment(key string, value ...int64) (int64, error) {
	r.memory.tagsMu.Lock()
	defer r.memory.tagsMu.Unlock()

	itemKey := r.itemKey(key)
	r.register(itemKey)

	return r.memory.Increment(itemKey, value...)
}

// Put an item in the cache for a given time.
func (r *memoryTaggedCache) Put(key string, value any, t time.Duration) error {
	r.memory.tagsMu.Lock()
	defer r.memory.tagsMu.Unlock()

	itemKey := r.itemKey(key)
	if err := r.memory.Put(itemKey, value, t); err != nil {
		return err
	}
	r.register(itemKey)

	return nil
}

// Pull retrieve an item from the cache and delete it.
func (r *memoryTaggedCache) Pull(key string, def ...any) any {
	res := r.Get(key, def...)
	r.Forget(key)

	return res
}

// Remember gets an item from the cache, or execute the given Closure and store the result.
func (r *memoryTaggedCache) Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error) {
	if val := r.Get(key); val != nil {
		return val, nil
	}

	val, err := callback()
	if err != nil {
		return nil, err
	}

	if err := r.Put(key, val, ttl); err != nil {
		return nil, err
	}

	return val, nil
}

// RememberForever get an item from the cache, or execute the given Closure and store the result forever.
func (r *memoryTaggedCache) RememberForever(key string, callback func() (any, error)) (any, error) {
	return r.Remember(key, NoExpiration, callback)
}

func (r *memoryTaggedCache) itemKey(key string) string {
	return r.namespace + ":" + key
}

// register adds the item key to the index of every tag, the caller must hold tagsMu.
func (r *memoryTaggedCache) register(itemKey string) {
	for _, name := range r.names {
		keys, exist := r.memory.tags[name]
		if !exist {
			keys = make(map[string]struct{})
			r.memory.tags[name] = keys
		}
		keys[itemKey] = struct{}{}
	}
}
