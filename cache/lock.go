package cache

import (
	"time"

	contractscache "github.com/goravel/framework/contracts/cache"
)

type Lock struct {
	store contractscache.Driver
	key   string
	time  *time.Duration
}

func NewLock(instance contractscache.Driver, key string, t ...time.Duration) *Lock {
	if len(t) == 0 {
		return &Lock{
			store: instance,
			key:   key,
		}
	}

	return &Lock{
		store: instance,
		key:   key,
		time:  &t[0],
	}
}

func (r *Lock) Block(t time.Duration, callback ...func()) bool {
	timer := time.NewTimer(t)
	res := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-timer.C:
				if r.Get(callback...) {
					res <- true
					return
				}

				res <- false
				return
			case <-time.Tick(1 * time.Second):
				if r.Get(callback...) {
					res <- true
					return
				}
			}
		}
	}()

	return <-res
}

func (r *Lock) Get(callback ...func()) bool {
	var res bool
	if r.time == nil {
		res = r.store.Add(r.key, 1, NoExpiration)
	} else {
		res = r.store.Add(r.key, 1, *r.time)
	}

	if !res {
		return false
	}

	if len(callback) == 0 {
		return true
	}

	callback[0]()

	return r.Release()
}

func (r *Lock) Release() bool {
	return r.store.Forget(r.key)
}
