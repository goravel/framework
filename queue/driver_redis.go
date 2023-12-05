package queue

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/json"
)

type redisData struct {
	Signature string               `json:"signature"`
	Args      []contractsqueue.Arg `json:"args"`
	Attempts  uint                 `json:"attempts"`
}

type Redis struct {
	connection string
	retryAfter uint
	client     *redis.Client
	ctx        context.Context
}

func NewRedis(connection string, client *redis.Client) *Redis {
	return &Redis{
		connection: connection,
		retryAfter: 60,
		client:     client,
		ctx:        context.Background(),
	}
}

func (r *Redis) ConnectionName() string {
	return r.connection
}

func (r *Redis) DriverName() string {
	return DriverRedis
}

func (r *Redis) Push(job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	payload, err := r.jobToJSON(job.Signature(), args)
	if err != nil {
		return err
	}

	return r.client.RPush(r.ctx, queue, payload).Err()
}

func (r *Redis) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	pipe := r.client.Pipeline()

	for _, job := range jobs {
		payload, err := r.jobToJSON(job.Job.Signature(), job.Args)
		if err != nil {
			return err
		}

		if job.Delay > 0 {
			delayDuration := time.Duration(job.Delay) * time.Second
			pipe.ZAdd(r.ctx, queue+":delayed", redis.Z{
				Score:  float64(time.Now().Add(delayDuration).Unix()),
				Member: payload,
			})
		} else {
			pipe.RPush(r.ctx, queue, payload)
		}
	}

	_, err := pipe.Exec(r.ctx)
	return err
}

func (r *Redis) Later(delay uint, job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	payload, err := r.jobToJSON(job.Signature(), args)
	if err != nil {
		return err
	}

	delayDuration := time.Duration(delay) * time.Second
	return r.client.ZAdd(r.ctx, queue+":delayed", redis.Z{
		Score:  float64(time.Now().Add(delayDuration).Unix()),
		Member: payload,
	}).Err()
}

func (r *Redis) Pop(queue string) (contractsqueue.Job, []contractsqueue.Arg, error) {
	if err := r.migrateDelayedJobs(queue); err != nil {
		return nil, nil, err
	}

	result, err := r.client.BLPop(r.ctx, 1*time.Second, queue).Result()
	if err == redis.Nil {
		return r.Pop(queue)
	} else if err != nil {
		return nil, nil, err
	}

	signature, args, err := r.jsonToJob(result[1])
	if err != nil {
		return nil, nil, err
	}
	job, err := Get(signature)
	if err != nil {
		return nil, nil, err
	}

	return job, args, nil
}

func (r *Redis) Delete(queue string, job contractsqueue.Jobs) error {
	payload, err := r.jobToJSON(job.Job.Signature(), job.Args)
	if err != nil {
		return err
	}

	return r.client.LRem(r.ctx, queue, 0, payload).Err()
}

func (r *Redis) Release(queue string, job contractsqueue.Jobs, delay uint) error {
	payload, err := r.jobToJSON(job.Job.Signature(), job.Args)
	if err != nil {
		return err
	}

	delayDuration := time.Duration(delay) * time.Second
	return r.client.ZAdd(r.ctx, queue+":delayed", redis.Z{
		Score:  float64(time.Now().Add(delayDuration).Unix()),
		Member: payload,
	}).Err()
}

func (r *Redis) Clear(queue string) error {
	return r.client.Del(r.ctx, queue).Err()
}

func (r *Redis) Size(queue string) (uint64, error) {
	size, err := r.client.LLen(r.ctx, queue).Result()
	return uint64(size), err
}

func (r *Redis) migrateDelayedJobs(queue string) error {
	jobs, err := r.client.ZRangeByScoreWithScores(r.ctx, queue+":delayed", &redis.ZRangeBy{
		Min:    "-inf",
		Max:    strconv.FormatFloat(float64(time.Now().Unix()), 'f', -1, 64),
		Offset: 0,
		Count:  -1,
	}).Result()
	if err != nil {
		return err
	}

	pipe := r.client.TxPipeline()
	for _, job := range jobs {
		// 将到期的任务转移到主队列
		pipe.RPush(r.ctx, queue, job.Member)
		// 从延迟队列中移除任务
		pipe.ZRem(r.ctx, queue+":delayed", job.Member)
	}
	_, err = pipe.Exec(r.ctx)
	if err != nil {
		return err
	}

	return nil
}

// jobToJSON convert signature and args to JSON
func (r *Redis) jobToJSON(signature string, args []contractsqueue.Arg) (string, error) {
	return json.MarshalString(&redisData{
		Signature: signature,
		Args:      args,
		Attempts:  0,
	})
}

// jsonToJob convert JSON to signature and args
func (r *Redis) jsonToJob(jsonString string) (string, []contractsqueue.Arg, error) {
	var data redisData
	err := json.UnmarshalString(jsonString, &data)
	if err != nil {
		return "", nil, err
	}

	return data.Signature, data.Args, nil
}
