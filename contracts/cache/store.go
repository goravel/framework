package cache

import "time"

type Store interface {
	//Get Retrieve an item from the cache by key.
	Get(key string, defaults interface{}) interface{}
	//Has Determine if an item exists in the cache.
	Has(key string) bool
	//Put Store an item in the cache for a given number of seconds.
	Put(key string, value interface{}, seconds time.Duration) error
	//Pull Retrieve an item from the cache and delete it.
	Pull(key string, defaults interface{}) interface{}
	//Add Store an item in the cache if the key does not exist.
	Add(key string, value interface{}, seconds time.Duration) bool
	//Remember Get an item from the cache, or execute the given Closure and store the result.
	Remember(key string, ttl time.Duration, callback func() interface{}) (interface{}, error)
	//RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
	RememberForever(key string, callback func() interface{}) (interface{}, error)
	//Forever Store an item in the cache indefinitely.
	Forever(key string, value interface{}) bool
	//Forget Remove an item from the cache.
	Forget(key string) bool
	//Flush Remove all items from the cache.
	Flush() bool
}
