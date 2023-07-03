package cache

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/spf13/cast"

	contractscache "github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
)

type Memory struct {
	ctx      context.Context
	prefix   string
	instance *cache.Cache
}

func NewMemory(config config.Config) (*Memory, error) {
	memory := cache.New(5*time.Minute, 10*time.Minute)

	return &Memory{
		prefix:   prefix(config),
		instance: memory,
	}, nil
}

//Add Driver an item in the cache if the key does not exist.
func (r *Memory) Add(key string, value any, t time.Duration) bool {
	if t == NoExpiration {
		t = cache.NoExpiration
	}

	err := r.instance.Add(r.key(key), value, t)

	return err == nil
}

func (r *Memory) Decrement(key string, value ...int) (int, error) {
	if len(value) == 0 {
		value = append(value, 1)
	}
	r.Add(key, 0, cache.NoExpiration)

	return r.instance.DecrementInt(r.key(key), value[0])
}

//Forever Driver an item in the cache indefinitely.
func (r *Memory) Forever(key string, value any) bool {
	if err := r.Put(key, value, cache.NoExpiration); err != nil {
		return false
	}

	return true
}

//Forget Remove an item from the cache.
func (r *Memory) Forget(key string) bool {
	r.instance.Delete(r.key(key))

	return true
}

//Flush Remove all items from the cache.
func (r *Memory) Flush() bool {
	r.instance.Flush()

	return true
}

//Get Retrieve an item from the cache by key.
func (r *Memory) Get(key string, def ...any) any {
	val, exist := r.instance.Get(r.key(key))
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

//Has Check an item exists in the cache.
func (r *Memory) Has(key string) bool {
	_, exist := r.instance.Get(r.key(key))

	return exist
}

func (r *Memory) Increment(key string, value ...int) (int, error) {
	if len(value) == 0 {
		value = append(value, 1)
	}
	r.Add(key, 0, cache.NoExpiration)

	return r.instance.IncrementInt(r.key(key), value[0])
}

func (r *Memory) Lock(key string, t ...time.Duration) contractscache.Lock {
	return NewLock(r, key, t...)
}

//Pull Retrieve an item from the cache and delete it.
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

//Put Driver an item in the cache for a given number of seconds.
func (r *Memory) Put(key string, value any, t time.Duration) error {
	r.instance.Set(r.key(key), value, t)

	return nil
}

//Remember Get an item from the cache, or execute the given Closure and store the result.
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

//RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
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

	if err := r.Put(key, val, cache.NoExpiration); err != nil {
		return nil, err
	}

	return val, nil
}

func (r *Memory) WithContext(ctx context.Context) contractscache.Driver {
	r.ctx = ctx

	return r
}

func (r *Memory) key(key string) string {
	return r.prefix + key
}
