package queue

import (
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
)

type ASync struct {
	connection string
	size       int64
	jobs       []contractsqueue.Jobs
}

func NewASync(connection string) *ASync {
	return &ASync{
		connection: connection,
		size:       0,
		jobs:       make([]contractsqueue.Jobs, 0),
	}
}

func (receiver *ASync) ConnectionName() string {
	return receiver.connection
}

func (receiver *ASync) Push(job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	receiver.size++
	receiver.jobs = append(receiver.jobs, contractsqueue.Jobs{Job: job, Args: args})

	return nil
}

func (receiver *ASync) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	receiver.size += int64(len(jobs))
	receiver.jobs = append(receiver.jobs, jobs...)

	return nil
}

func (receiver *ASync) Later(delay int, job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	receiver.size++
	receiver.jobs = append(receiver.jobs, contractsqueue.Jobs{Job: job, Args: args, Delay: int64(delay)})

	return nil
}

func (receiver *ASync) Pop(queue string) (contractsqueue.Job, []contractsqueue.Arg, error) {
	if len(receiver.jobs) == 0 {
		return nil, nil, nil
	}

	job := receiver.jobs[0]
	receiver.jobs = receiver.jobs[1:]

	return job.Job, job.Args, nil
}

func (receiver *ASync) Delete(queue string, job contractsqueue.Job) error {
	return nil
}

func (receiver *ASync) Release(queue string, job contractsqueue.Job, delay int) error {
	receiver.size++
	receiver.jobs = append(receiver.jobs, contractsqueue.Jobs{Job: job, Delay: int64(delay)})
	return nil
}

func (receiver *ASync) Clear(queue string) error {
	receiver.jobs = make([]contractsqueue.Jobs, 0)
	receiver.size = 0
	return nil
}

func (receiver *ASync) Size(queue string) (int64, error) {
	return receiver.size, nil
}

func (receiver *ASync) Server(concurrent int, queue string) {
	var errChan chan error

	go func() {
		for {
			if len(receiver.jobs) == 0 {
				time.Sleep(time.Second)
				continue
			}

			job, args, err := receiver.Pop(queue)
			if err != nil {
				receiver.size--
				time.Sleep(time.Second)
				continue
			}

			err = Call(job.Signature(), args)
			if err != nil {
				receiver.size--
				errChan <- err
			}

			time.Sleep(time.Second)
		}
	}()
}
