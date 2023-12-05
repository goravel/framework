package queue

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

type Worker struct {
	concurrent int
	driver     queue.Driver
	failedJobs orm.Query
	queue      string
}

func NewWorker(config *Config, concurrent int, connection string, queue string) *Worker {
	return &Worker{
		concurrent: concurrent,
		driver:     NewDriver(connection, config),
		failedJobs: config.FailedJobsDatabase(),
		queue:      queue,
	}
}

func (r *Worker) Run() error {
	if r.driver.DriverName() == DriverSync {
		return fmt.Errorf("queue %s driver not need run", r.queue)
	}

	failedJobChan := make(chan FailedJob)
	sigChan := make(chan os.Signal, 1)
	quitChan := make(chan struct{})
	var wg sync.WaitGroup
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	for i := 0; i < r.concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-quitChan:
					return
				default:
					job, args, err := r.driver.Pop(r.queue)
					if err != nil {
						// This error not need to be reported.
						// It is usually caused by the queue being empty.
						continue
					}
					err = Call(job.Signature(), args)
					if err != nil {
						failedJobChan <- FailedJob{
							Queue:     r.queue,
							Job:       job.Signature(),
							Payloads:  args,
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
		fmt.Printf("Received signal: %s, shutting down queue %s...\n", sig, r.queue)
		close(quitChan)
		wg.Wait()
		close(failedJobChan)
	}()

	for job := range failedJobChan {
		_ = r.failedJobs.Create(&job)
	}

	return nil
}
