package queue

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type Worker struct {
	concurrent    int
	driver        *DriverImpl
	job           *JobImpl
	failedJobs    orm.Query
	queue         string
	failedJobChan chan FailedJob
	sigChan       chan os.Signal
	quitChan      chan struct{}
}

func NewWorker(config *Config, concurrent int, connection string, queue string, job *JobImpl) *Worker {
	return &Worker{
		concurrent:    concurrent,
		driver:        NewDriverImpl(connection, config),
		job:           job,
		failedJobs:    config.FailedJobsQuery(),
		queue:         queue,
		failedJobChan: make(chan FailedJob),
		sigChan:       make(chan os.Signal, 1),
		quitChan:      make(chan struct{}),
	}
}

func (r *Worker) Run() error {
	driver, err := r.driver.New()
	if err != nil {
		return err
	}
	if driver.Driver() == DriverSync {
		return fmt.Errorf("queue %s driver not need run", r.queue)
	}

	var wg sync.WaitGroup
	signal.Notify(r.sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	for i := 0; i < r.concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-r.quitChan:
					return
				default:
					job, args, err := driver.Pop(r.queue)
					if err != nil {
						// This error not need to be reported.
						// It is usually caused by the queue being empty.
						time.Sleep(1 * time.Second)
						continue
					}

					err = r.job.Call(job.Signature(), args)
					if err != nil {
						r.failedJobChan <- FailedJob{
							Queue:     r.queue,
							Signature: job.Signature(),
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
		sig := <-r.sigChan
		fmt.Printf("Received signal: %s, shutting down queue %s...\n", sig, r.queue)
		close(r.quitChan)
		wg.Wait()
		close(r.failedJobChan)
	}()

	go func() {
		for job := range r.failedJobChan {
			_ = r.failedJobs.Create(&job)
		}
	}()

	return nil
}

func (r *Worker) Shutdown() error {
	r.sigChan <- syscall.SIGTERM
	return nil
}
