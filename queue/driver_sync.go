package queue

import (
	"time"

	"github.com/goravel/framework/contracts/queue"
)

type Sync struct {
	connection string
	size       int64
}

func NewSync(connection string) *Sync {
	return &Sync{
		connection: connection,
		size:       0,
	}
}

func (receiver *Sync) ConnectionName() string {
	return receiver.connection
}

func (receiver *Sync) Push(job queue.Job, args []queue.Arg, queue string) error {
	receiver.size++
	err := Call(job.Signature(), args)
	receiver.size--

	return err
}

func (receiver *Sync) Bulk(jobs []queue.Jobs, queue string) error {
	receiver.size += int64(len(jobs))

	for _, job := range jobs {
		err := Call(job.Job.Signature(), job.Args)
		if err != nil {
			receiver.size -= int64(len(jobs))
			return err
		}
	}

	receiver.size -= int64(len(jobs))
	return nil
}

func (receiver *Sync) Later(delay int, job queue.Job, args []queue.Arg, queue string) error {
	receiver.size++
	time.Sleep(time.Duration(delay) * time.Second)
	err := Call(job.Signature(), args)
	receiver.size--

	return err
}

func (receiver *Sync) Pop(queue string) (queue.Job, []queue.Arg, error) {
	// Sync driver does not need to pop job.
	return nil, nil, nil
}

func (receiver *Sync) Delete(queue string, job queue.Job) error {
	// Sync driver does not support delete job.
	return nil
}

func (receiver *Sync) Release(queue string, job queue.Job, delay int) error {
	// Sync driver does not support release job.
	return nil
}

func (receiver *Sync) Clear(queue string) error {
	receiver.size = 0
	return nil
}

func (receiver *Sync) Size(queue string) (int64, error) {
	return receiver.size, nil
}

func (receiver *Sync) Server(concurrent int, queue string) {
	// Sync driver does not need to run server.
}
