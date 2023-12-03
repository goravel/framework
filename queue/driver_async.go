package queue

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

type ASync struct {
	connection string
	size       int64
	jobs       []contractsqueue.Jobs
	failedJobs orm.Query
	mu         sync.Mutex
}

func NewASync(connection string, failedJobs orm.Query) *ASync {
	return &ASync{
		connection: connection,
		size:       0,
		jobs:       make([]contractsqueue.Jobs, 0),
		failedJobs: failedJobs,
	}
}

func (receiver *ASync) ConnectionName() string {
	return receiver.connection
}

func (receiver *ASync) Push(job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	receiver.size++
	receiver.jobs = append(receiver.jobs, contractsqueue.Jobs{Job: job, Args: args})

	return nil
}

func (receiver *ASync) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	receiver.size += int64(len(jobs))
	receiver.jobs = append(receiver.jobs, jobs...)

	return nil
}

func (receiver *ASync) Later(delay int, job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	receiver.size++
	receiver.jobs = append(receiver.jobs, contractsqueue.Jobs{Job: job, Args: args, Delay: int64(delay)})

	return nil
}

func (receiver *ASync) Pop(queue string) (contractsqueue.Job, []contractsqueue.Arg, error) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	if len(receiver.jobs) == 0 {
		return nil, nil, nil
	}

	job := receiver.jobs[0]
	receiver.jobs = receiver.jobs[1:]

	return job.Job, job.Args, nil
}

func (receiver *ASync) Delete(queue string, job contractsqueue.Job) error {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	flag := true
	for i, j := range receiver.jobs {
		if j.Job == job {
			receiver.jobs = append(receiver.jobs[:i], receiver.jobs[i+1:]...)
			receiver.size--
			flag = false
			break
		}
	}

	if flag {
		return fmt.Errorf("job %s not found", job.Signature())
	}

	return nil
}

func (receiver *ASync) Release(queue string, job contractsqueue.Job, delay int) error {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	receiver.size++
	receiver.jobs = append(receiver.jobs, contractsqueue.Jobs{Job: job, Delay: int64(delay)})
	return nil
}

func (receiver *ASync) Clear(queue string) error {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	receiver.jobs = make([]contractsqueue.Jobs, 0)
	receiver.size = 0
	return nil
}

func (receiver *ASync) Size(queue string) (int64, error) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	return receiver.size, nil
}

func (receiver *ASync) Server(concurrent int, queue string) {
	failedJobChan := make(chan FailedJob)
	sigChan := make(chan os.Signal)
	quitChan := make(chan struct{})
	var wg sync.WaitGroup
	signal.Notify(sigChan)

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-quitChan:
					return
				default:
					if len(receiver.jobs) == 0 {
						time.Sleep(time.Second)
						continue
					}

					job, args, _ := receiver.Pop(queue)
					err := Call(job.Signature(), args)
					if err != nil {
						receiver.size--
						failedJobChan <- FailedJob{
							Queue:     queue,
							Job:       job.Signature(),
							Arg:       args,
							Exception: err.Error(),
							FailedAt:  carbon.DateTime{Carbon: carbon.Now()},
						}
					}
				}
			}
		}()
	}

	go func() {
		sig := <-sigChan
		fmt.Printf("Received signal: %s, shutting down...\n", sig)
		close(quitChan)
		close(failedJobChan)
		wg.Wait()
	}()

	for {
		job := <-failedJobChan
		_ = receiver.failedJobs.Create(&job)
	}
}
