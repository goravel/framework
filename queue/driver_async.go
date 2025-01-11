package queue

import (
	"sync"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

var asyncQueues sync.Map

type Async struct {
	connection string
	size       int
}

func NewAsync(connection string, size int) *Async {
	return &Async{
		connection: connection,
		size:       size,
	}
}

func (r *Async) Connection() string {
	return r.connection
}

func (r *Async) Driver() string {
	return contractsqueue.DriverAsync
}

func (r *Async) Push(job contractsqueue.Job, args []any, queue string) error {
	r.getQueue(queue) <- contractsqueue.Jobs{Job: job, Args: args}
	return nil
}

func (r *Async) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	for _, job := range jobs {
		go func() {
			time.Sleep(time.Until(job.Delay))
			r.getQueue(queue) <- job
		}()
	}

	return nil
}

func (r *Async) Later(delay time.Time, job contractsqueue.Job, args []any, queue string) error {
	go func() {
		time.Sleep(time.Until(delay))
		r.getQueue(queue) <- contractsqueue.Jobs{Job: job, Args: args}
	}()

	return nil
}

func (r *Async) Pop(queue string) (contractsqueue.Job, []any, error) {
	ch, ok := asyncQueues.Load(queue)
	if !ok {
		return nil, nil, errors.QueueDriverAsyncNoJobFound.Args(queue)
	}

	queueChan := ch.(chan contractsqueue.Jobs)
	select {
	case job := <-queueChan:
		return job.Job, job.Args, nil
	default:
		return nil, nil, errors.QueueDriverAsyncNoJobFound.Args(queue)
	}
}

func (r *Async) getQueue(queue string) chan contractsqueue.Jobs {
	ch, ok := asyncQueues.Load(queue)
	if !ok {
		ch = make(chan contractsqueue.Jobs, r.size)
		actual, _ := asyncQueues.LoadOrStore(queue, ch)
		return actual.(chan contractsqueue.Jobs)
	}
	return ch.(chan contractsqueue.Jobs)
}
