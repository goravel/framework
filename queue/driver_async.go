package queue

import (
	"fmt"
	"sync"

	contractsqueue "github.com/goravel/framework/contracts/queue"
)

type ASync struct {
	connection string
	size       uint64
	jobs       []contractsqueue.Jobs
	mu         sync.Mutex
}

func NewASync(connection string) *ASync {
	return &ASync{
		connection: connection,
		size:       0,
		jobs:       make([]contractsqueue.Jobs, 0),
	}
}

func (r *ASync) ConnectionName() string {
	return r.connection
}

func (r *ASync) DriverName() string {
	return DriverASync
}

func (r *ASync) Push(job contractsqueue.Job, payloads []contractsqueue.Payloads, queue string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size++
	r.jobs = append(r.jobs, contractsqueue.Jobs{Job: job, Payloads: payloads})

	return nil
}

func (r *ASync) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size += uint64(len(jobs))
	r.jobs = append(r.jobs, jobs...)

	return nil
}

func (r *ASync) Later(delay uint, job contractsqueue.Job, payloads []contractsqueue.Payloads, queue string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size++
	r.jobs = append(r.jobs, contractsqueue.Jobs{Job: job, Payloads: payloads, Delay: delay})

	return nil
}

func (r *ASync) Pop(queue string) (contractsqueue.Job, []contractsqueue.Payloads, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.jobs) == 0 {
		return nil, nil, nil
	}

	job := r.jobs[0]
	r.jobs = r.jobs[1:]

	return job.Job, job.Payloads, nil
}

func (r *ASync) Delete(queue string, job contractsqueue.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	flag := true
	for i, j := range r.jobs {
		if j.Job == job {
			r.jobs = append(r.jobs[:i], r.jobs[i+1:]...)
			r.size--
			flag = false
			break
		}
	}

	if flag {
		return fmt.Errorf("job %s not found", job.Signature())
	}

	return nil
}

func (r *ASync) Release(queue string, job contractsqueue.Job, delay uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size++
	r.jobs = append(r.jobs, contractsqueue.Jobs{Job: job, Delay: delay})
	return nil
}

func (r *ASync) Clear(queue string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.jobs = make([]contractsqueue.Jobs, 0)
	r.size = 0
	return nil
}

func (r *ASync) Size(queue string) (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.size, nil
}
