package queue

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/exp/slices"

	contractsqueue "github.com/goravel/framework/contracts/queue"
)

type ASync struct {
	connection string
	size       uint64
	mu         sync.Mutex
}

// asyncJobs is a map to store all registered jobs.
var asyncJobs = make(map[string][]contractsqueue.Jobs)

func NewASync(connection string) *ASync {
	return &ASync{
		connection: connection,
		size:       0,
	}
}

func (r *ASync) ConnectionName() string {
	return r.connection
}

func (r *ASync) DriverName() string {
	return DriverASync
}

func (r *ASync) Push(job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size++
	asyncJobs[queue] = append(asyncJobs[queue], contractsqueue.Jobs{Job: job, Args: args})

	return nil
}

func (r *ASync) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size += uint64(len(jobs))
	asyncJobs[queue] = append(asyncJobs[queue], jobs...)

	return nil
}

func (r *ASync) Later(delay uint, job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	time.AfterFunc(time.Duration(delay)*time.Second, func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		r.size++
		asyncJobs[queue] = append(asyncJobs[queue], contractsqueue.Jobs{Job: job, Args: args})
	})

	return nil
}

func (r *ASync) Pop(queue string) (contractsqueue.Job, []contractsqueue.Arg, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := asyncJobs[queue]; !exists {
		time.Sleep(1 * time.Second)
		return nil, nil, errors.New("no job found in queue")
	}
	if len(asyncJobs[queue]) == 0 {
		delete(asyncJobs, queue)
		time.Sleep(1 * time.Second)
		return nil, nil, errors.New("no job found in queue")
	}

	job := asyncJobs[queue][0]

	if len(asyncJobs[queue]) == 1 {
		delete(asyncJobs, queue)
	} else {
		asyncJobs[queue] = asyncJobs[queue][1:]
	}

	return job.Job, job.Args, nil
}

func (r *ASync) Delete(queue string, job contractsqueue.Jobs) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := asyncJobs[queue]; !exists {
		return errors.New("no job found in queue")
	}

	for i, j := range asyncJobs[queue] {
		if j.Job.Signature() == job.Job.Signature() && slices.Equal(j.Args, job.Args) {
			asyncJobs[queue] = append(asyncJobs[queue][:i], asyncJobs[queue][i+1:]...)
			r.size--
			return nil
		}
	}

	return fmt.Errorf("job %s not found", job.Job.Signature())
}

func (r *ASync) Release(queue string, job contractsqueue.Jobs, delay uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := asyncJobs[queue]; !exists {
		return errors.New("no job found in queue")
	}

	job.Delay = delay
	r.size++

	asyncJobs[queue] = append(asyncJobs[queue], job)
	return nil
}

func (r *ASync) Clear(queue string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(asyncJobs, queue)
	r.size = 0
	return nil
}

func (r *ASync) Size(queue string) (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.size, nil
}
