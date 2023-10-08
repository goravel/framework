package queue

import (
	"errors"
	"sync"

	"github.com/goravel/framework/contracts/queue"
)

var mutex sync.RWMutex

// Register registers jobs to the registry.
// Register 将作业注册到注册表。
func Register(jobs []queue.Job) error {
	mutex.Lock()
	defer mutex.Unlock()

	for _, job := range jobs {
		signature := job.Signature()
		if _, exists := JobRegistry[signature]; exists {
			return errors.New("Job with signature " + signature + " already exists")
		}
		JobRegistry[signature] = job
	}

	return nil
}

// Call calls a registered job using its signature.
// Call 使用其签名调用已注册的作业。
func Call(signature string, args ...any) error {
	mutex.RLock()
	defer mutex.RUnlock()

	job, exists := JobRegistry[signature]
	if !exists {
		return errors.New("job not found")
	}
	return job.Handle(args...)
}
