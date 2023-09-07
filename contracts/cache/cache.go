package cache

import (
	"context"
	"time"
)

//go:generate mockery --name=Cache
type Cache interface {
	Driver
	Store(name string) Driver
}

//go:generate mockery --name=Driver
type Driver interface {
	//Add Driver an item in the cache if the key does not exist.
	Add(key string, value any, t time.Duration) bool
	//Decrement Decrements the value of an item in the cache.
	Decrement(key string, value ...int) (int, error)
	//Forever Driver an item in the cache indefinitely.
	Forever(key string, value any) bool
	//Forget Remove an item from the cache.
	Forget(key string) bool
	//Flush Remove all items from the cache.
	Flush() bool
	//Get Retrieve an item from the cache by key.
	Get(key string, def ...any) any
	// GetBool retrieves an item from the cache by key as a boolean.
	GetBool(key string, def ...bool) bool
	// GetInt retrieves an item from the cache by key as an integer.
	GetInt(key string, def ...int) int
	// GetInt64 retrieves an item from the cache by key as a 64-bit integer.
	GetInt64(key string, def ...int64) int64
	// GetString retrieves an item from the cache by key as a string.
	GetString(key string, def ...string) string
	//Has Check an item exists in the cache.
	Has(key string) bool
	//Increment Increments the value of an item in the cache.
	Increment(key string, value ...int) (int, error)
	//Lock Get a lock instance.
	Lock(key string, t ...time.Duration) Lock
	//Put Driver an item in the cache for a given time.
	Put(key string, value any, t time.Duration) error
	//Pull Retrieve an item from the cache and delete it.
	Pull(key string, def ...any) any
	//Remember Get an item from the cache, or execute the given Closure and store the result.
	Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error)
	//RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
	RememberForever(key string, callback func() (any, error)) (any, error)
	//WithContext returns a new Cache instance with the given context.
	WithContext(ctx context.Context) Driver
}

//go:generate mockery --name=Lock
type Lock interface {
	//Block Attempt to acquire the lock for the given number of seconds.
	Block(t time.Duration, callback ...func()) bool
	//Get Attempt to acquire the lock.
	Get(callback ...func()) bool
	//Release the lock.
	Release() bool
	//ForceRelease Releases the lock in disregard of ownership.
	ForceRelease() bool
}
