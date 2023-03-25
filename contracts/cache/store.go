package cache

import (
	"context"
	"time"
)

//go:generate mockery --name=Store
type Store interface {
	//Add Store an item in the cache if the key does not exist.
	Add(key string, value any, sec time.Duration) bool
	Decrement(key string, value ...int) int
	//Forever Store an item in the cache indefinitely.
	Forever(key string, value any) bool
	//Forget Remove an item from the cache.
	Forget(key string) bool
	//Flush Remove all items from the cache.
	Flush() bool
	//Get Retrieve an item from the cache by key.
	Get(key string, def ...any) any
	GetBool(key string, def ...bool) bool
	GetInt(key string, def ...int) int
	GetInt64(key string, def ...int64) int64
	GetString(key string, def ...string) string
	//Has Check an item exists in the cache.
	Has(key string) bool
	Increment(key string, value ...int) int
	Lock(key string, second int)
	//Put Store an item in the cache for a given number of seconds.
	Put(key string, value any, sec time.Duration) error
	//Pull Retrieve an item from the cache and delete it.
	Pull(key string, def ...any) any
	//Remember Get an item from the cache, or execute the given Closure and store the result.
	Remember(key string, ttl time.Duration, callback func() any) (any, error)
	//RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
	RememberForever(key string, callback func() any) (any, error)
	Tags(name ...string) Store
	WithContext(ctx context.Context) Store
}
