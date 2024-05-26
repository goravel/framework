package queue

import (
	"fmt"
	"sync"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
)

type ASync struct {
	connection string
}

// asyncJobs is a map to store all registered jobs.
var asyncJobs = make(map[string][]contractsqueue.Jobs)

// asyncMu is a mutex
var asyncMu sync.Mutex

func NewASync(connection string) *ASync {
	return &ASync{
		connection: connection,
	}
}

func (r *ASync) Connection() string {
	return r.connection
}

func (r *ASync) Driver() string {
	return DriverASync
}

func (r *ASync) Push(job contractsqueue.Job, args []any, queue string) error {
	asyncMu.Lock()
	defer asyncMu.Unlock()

	asyncJobs[queue] = append(asyncJobs[queue], contractsqueue.Jobs{Job: job, Args: args})
	return nil
}

func (r *ASync) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	asyncMu.Lock()
	defer asyncMu.Unlock()

	for _, job := range jobs {
		if job.Delay > 0 {
			time.AfterFunc(time.Duration(job.Delay)*time.Second, func() {
				asyncMu.Lock()
				defer asyncMu.Unlock()

				asyncJobs[queue] = append(asyncJobs[queue], job)
			})
			continue
		}

		asyncJobs[queue] = append(asyncJobs[queue], job)
	}

	return nil
}

func (r *ASync) Later(delay uint, job contractsqueue.Job, args []any, queue string) error {
	time.AfterFunc(time.Duration(delay)*time.Second, func() {
		asyncMu.Lock()
		defer asyncMu.Unlock()

		asyncJobs[queue] = append(asyncJobs[queue], contractsqueue.Jobs{Job: job, Args: args})
	})

	return nil
}

func (r *ASync) Pop(queue string) (contractsqueue.Job, []any, error) {
	asyncMu.Lock()
	defer asyncMu.Unlock()

	if len(asyncJobs[queue]) == 0 {
		delete(asyncJobs, queue)
		return nil, nil, fmt.Errorf("no job found in %s queue", queue)
	}

	job := asyncJobs[queue][0]
	if len(asyncJobs[queue]) == 1 {
		delete(asyncJobs, queue)
	} else {
		asyncJobs[queue] = asyncJobs[queue][1:]
	}

	return job.Job, job.Args, nil
}
