package queue

import (
	"fmt"
	"sync"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
)

var asyncQueues sync.Map

type ASync struct {
	connection string
	size       int
}

func NewASync(connection string, size int) *ASync {
	return &ASync{
		connection: connection,
		size:       size,
	}
}

func (r *ASync) Connection() string {
	return r.connection
}

func (r *ASync) Driver() string {
	return DriverASync
}

func (r *ASync) Push(job contractsqueue.Job, args []any, queue string) error {
	r.getQueue(queue) <- contractsqueue.Jobs{Job: job, Args: args}
	return nil
}

func (r *ASync) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	for _, job := range jobs {
		if job.Delay > 0 {
			go func(j contractsqueue.Jobs) {
				time.Sleep(j.Delay)
				r.getQueue(queue) <- j
			}(job)
			continue
		}

		r.getQueue(queue) <- job
	}

	return nil
}

func (r *ASync) Later(delay time.Duration, job contractsqueue.Job, args []any, queue string) error {
	go func() {
		time.Sleep(delay)
		r.getQueue(queue) <- contractsqueue.Jobs{Job: job, Args: args}
	}()

	return nil
}

func (r *ASync) Pop(queue string) (contractsqueue.Job, []any, error) {
	ch, ok := asyncQueues.Load(queue)
	if !ok {
		return nil, nil, fmt.Errorf("no queue found: %s", queue)
	}

	queueChan := ch.(chan contractsqueue.Jobs)
	select {
	case job := <-queueChan:
		return job.Job, job.Args, nil
	default:
		return nil, nil, fmt.Errorf("no job found in %s queue", queue)
	}
}

func (r *ASync) getQueue(queue string) chan contractsqueue.Jobs {
	ch, ok := asyncQueues.Load(queue)
	if !ok {
		ch = make(chan contractsqueue.Jobs, r.size)
		actual, _ := asyncQueues.LoadOrStore(queue, ch)
		return actual.(chan contractsqueue.Jobs)
	}
	return ch.(chan contractsqueue.Jobs)
}
