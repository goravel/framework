package queue

import (
	contractscache "github.com/goravel/framework/contracts/cache"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/queue/utils"
)

// uniqueLockPrefix prefixes the cache lock that keeps a unique job from
// being pushed again while one is pending or running.
const uniqueLockPrefix = "goravel:queue-unique:"

// isUniqueJob reports whether the task's job is unique.
func isUniqueJob(task contractsqueue.Task) bool {
	_, ok := task.Job.(contractsqueue.JobWithUnique)

	return ok
}

// uniqueLockKey returns the key of the unique lock of the task's job, and
// whether the job is unique.
func uniqueLockKey(task contractsqueue.Task) (string, bool) {
	unique, ok := task.Job.(contractsqueue.JobWithUnique)
	if !ok {
		return "", false
	}

	key := uniqueLockPrefix + task.Job.Signature()
	if id := unique.UniqueID(utils.ConvertArgs(task.Args)...); id != "" {
		key += ":" + id
	}

	return key, true
}

// acquireUniqueLock takes the unique lock of the task's job. It returns false
// when a job with the same unique ID is already pending or running, in which
// case the task must not be pushed. It returns QueueUniqueJobWithoutCache
// when the job is unique but no cache store is bound.
func acquireUniqueLock(cache contractscache.Cache, task contractsqueue.Task) (bool, error) {
	key, ok := uniqueLockKey(task)
	if !ok {
		return true, nil
	}
	if cache == nil {
		return false, errors.QueueUniqueJobWithoutCache.Args(task.Job.Signature())
	}

	if uniqueFor := task.Job.(contractsqueue.JobWithUnique).UniqueFor(); uniqueFor > 0 {
		return cache.Lock(key, uniqueFor).Get(), nil
	}

	return cache.Lock(key).Get(), nil
}

// releaseUniqueLock releases the unique lock of the task's job, which may
// have been taken by another process when the job was dispatched.
func releaseUniqueLock(cache contractscache.Cache, task contractsqueue.Task) {
	if cache == nil {
		return
	}

	if key, ok := uniqueLockKey(task); ok {
		cache.Lock(key).ForceRelease()
	}
}
