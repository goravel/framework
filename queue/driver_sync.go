package queue

import (
	"sync"
	"time"

	"github.com/goravel/framework/contracts/queue"
)

type Sync struct {
	connection string
	size       uint
	mu         sync.Mutex
}

func NewSync(connection string) *Sync {
	return &Sync{
		connection: connection,
		size:       0,
	}
}

func (r *Sync) ConnectionName() string {
	return r.connection
}

func (r *Sync) DriverName() string {
	return DriverSync
}

func (r *Sync) Push(job queue.Job, args []queue.Payloads, queue string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size++
	err := Call(job.Signature(), args)
	r.size--

	return err
}

func (r *Sync) Bulk(jobs []queue.Jobs, queue string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size += uint(len(jobs))
	for _, job := range jobs {
		err := Call(job.Job.Signature(), job.Payloads)
		if err != nil {
			r.size -= uint(len(jobs))
			return err
		}
	}
	r.size -= uint(len(jobs))

	return nil
}

func (r *Sync) Later(delay uint, job queue.Job, args []queue.Payloads, queue string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size++
	time.Sleep(time.Duration(delay) * time.Second)
	err := Call(job.Signature(), args)
	r.size--

	return err
}

func (r *Sync) Pop(queue string) (queue.Job, []queue.Payloads, error) {
	return nil, nil, nil
}

func (r *Sync) Delete(queue string, job queue.Job) error {
	return nil
}

func (r *Sync) Release(queue string, job queue.Job, delay uint) error {
	return nil
}

func (r *Sync) Clear(queue string) error {
	r.size = 0
	return nil
}

func (r *Sync) Size(queue string) (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return uint64(r.size), nil
}
