package cache

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"

	cachecontract "github.com/goravel/framework/contracts/cache"
)

type Memory struct {
	ctx      context.Context
	prefix   string
	instance *cache.Cache
}

func NewMemory() (*Memory, error) {
	memory := cache.New(5*time.Minute, 10*time.Minute)

	return &Memory{
		prefix:   prefix(),
		instance: memory,
	}, nil
}

func (r *Memory) WithContext(ctx context.Context) cachecontract.Store {
	return r
}

//Add Store an item in the cache if the key does not exist.
func (r *Memory) Add(key string, value any, seconds time.Duration) bool {
	if err := r.instance.Add(r.prefix+key, value, seconds); err != nil {
		return false
	}

	return true
}

//Forever Store an item in the cache indefinitely.
func (r *Memory) Forever(key string, value any) bool {
	if err := r.Put(key, value, cache.NoExpiration); err != nil {
		return false
	}

	return true
}

//Forget Remove an item from the cache.
func (r *Memory) Forget(key string) bool {
	r.instance.Delete(r.prefix + key)

	return true
}

//Flush Remove all items from the cache.
func (r *Memory) Flush() bool {
	r.instance.Flush()

	return true
}

//Get Retrieve an item from the cache by key.
func (r *Memory) Get(key string, def any) any {
	val, exist := r.instance.Get(r.prefix + key)
	if exist {
		return val
	}

	switch s := def.(type) {
	case func() any:
		return s()
	default:
		return def
	}
}

func (r *Memory) GetBool(key string, def bool) bool {
	res := r.Get(key, def)

	return res.(bool)
}

func (r *Memory) GetInt(key string, def int) int {
	res := r.Get(key, def)

	return res.(int)
}

func (r *Memory) GetString(key string, def string) string {
	return r.Get(key, def).(string)
}

//Has Check an item exists in the cache.
func (r *Memory) Has(key string) bool {
	_, exist := r.instance.Get(r.prefix + key)

	return exist
}

//Pull Retrieve an item from the cache and delete it.
func (r *Memory) Pull(key string, def any) any {
	res := r.Get(key, def)
	r.Forget(key)

	return res
}

//Put Store an item in the cache for a given number of seconds.
func (r *Memory) Put(key string, value any, seconds time.Duration) error {
	r.instance.Set(r.prefix+key, value, seconds)

	return nil
}

//Remember Get an item from the cache, or execute the given Closure and store the result.
func (r *Memory) Remember(key string, seconds time.Duration, callback func() any) (any, error) {
	val := r.Get(key, nil)

	if val != nil {
		return val, nil
	}

	val = callback()

	if err := r.Put(key, val, seconds); err != nil {
		return nil, err
	}

	return val, nil
}

//RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
func (r *Memory) RememberForever(key string, callback func() any) (any, error) {
	val := r.Get(key, nil)

	if val != nil {
		return val, nil
	}

	val = callback()

	if err := r.Put(key, val, cache.NoExpiration); err != nil {
		return nil, err
	}

	return val, nil
}
