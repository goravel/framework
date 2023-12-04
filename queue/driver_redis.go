package queue

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	connection string
	client     *redis.Client
	failedJobs orm.Query
	ctx        context.Context
}

func NewRedis(connection string, client *redis.Client, failedJobs orm.Query) *Redis {
	return &Redis{
		connection: connection,
		client:     client,
		failedJobs: failedJobs,
		ctx:        context.Background(),
	}
}

func (receiver *Redis) ConnectionName() string {
	return receiver.connection
}

func (receiver *Redis) Push(job queue.Job, args []queue.Payloads, queue string) error {
	return receiver.client.RPush(receiver.ctx, queue, job).Err()
}

func (receiver *Redis) Bulk(jobs []queue.Jobs, queue string) error {
	for _, job := range jobs {
		err := receiver.client.RPush(receiver.ctx, queue, job).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (receiver *Redis) Later(delay int, job queue.Job, args []queue.Payloads, queue string) error {
	return receiver.client.ZAdd(receiver.ctx, queue, redis.Z{
		Score:  float64(time.Now().Add(time.Duration(delay) * time.Second).Unix()),
		Member: job,
	}).Err()
}

func (receiver *Redis) Pop(queue string) (queue.Job, []queue.Payloads, error) {
	job, err := receiver.client.XRead(receiver.ctx, queue).Result()
	if err != nil {
		return nil, nil, err
	}

	return job, nil, nil
}

func (receiver *Redis) Delete(queue string, job queue.Job) error {
	// 实现从队列删除任务的逻辑
	// 这里只是一个示例，具体实现需要根据你的需求来
	return receiver.client.LRem(receiver.ctx, queue, 0, job).Err()
}

func (receiver *Redis) Release(queue string, job queue.Job, delay int) error {
	// 实现释放任务回到队列的逻辑
	// 这里只是一个示例，具体实现需要根据你的需求来
	return receiver.client.ZAdd(receiver.ctx, queue, redis.Z{
		Score:  float64(time.Now().Add(time.Duration(delay) * time.Second).Unix()),
		Member: job,
	}).Err()
}

func (receiver *Redis) Clear(queue string) error {
	// 实现清空队列的逻辑
	// 这里只是一个示例，具体实现需要根据你的需求来
	return receiver.client.Del(receiver.ctx, queue).Err()
}

func (receiver *Redis) Size(queue string) (int64, error) {
	// 实现获取队列大小的逻辑
	// 这里只是一个示例，具体实现需要根据你的需求来
	return receiver.client.LLen(receiver.ctx, queue).Result()
}

func (receiver *Redis) Server(concurrent int, queue string) {
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
					job, args, _ := receiver.Pop(queue)
					err := Call(job.Signature(), args)
					if err != nil {
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
