package queue

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/json"
)

type redisData struct {
	Signature string `json:"signature"`
	Args      []any  `json:"args"`
	Attempts  uint   `json:"attempts"`
}

type Redis struct {
	connection string
	retryAfter uint
	client     *redis.Client
	lua        RedisLua
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

func (r *Redis) Push(job contractsqueue.Job, payloads []any, queue string) error {
	data, err := json.MarshalString(&redisData{
		Signature: job.Signature(),
		Args:      payloads,
	})
	if err != nil {
		return err
	}

	return r.lua.Push().Run(r.ctx, r.client, []string{queue, queue + ":notify"}, data).Err()
}

func (r *Redis) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	for _, job := range jobs {
		if job.Delay > 0 {
			if err := r.Later(job.Delay, job.Job, job.Payloads, queue); err != nil {
				return err
			}
		} else {
			if err := r.Push(job.Job, job.Payloads, queue); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Redis) Later(delay uint, job contractsqueue.Job, payloads []any, queue string) error {
	data, err := json.MarshalString(&redisData{
		Signature: job.Signature(),
		Args:      payloads,
	})
	if err != nil {
		return err
	}

	return r.client.ZAdd(r.ctx, queue+":delayed", redis.Z{
		Score:  float64(carbon.Now().AddSeconds(int(delay)).Timestamp()),
		Member: data,
	}).Err()
}

func (r *Redis) Pop(queue string) (contractsqueue.Job, []any, error) {
	r.migrate(queue)

	job, err := r.retrieveNextJob(queue)
	if err != nil {
		return nil, nil, err
	}

	return job.Job, job.Payloads, nil
}

func (r *Redis) Delete(queue string, job contractsqueue.Jobs) error {
	payload, err := json.MarshalString(&redisData{
		Signature: job.Job.Signature(),
		Args:      job.Payloads,
	})
	if err != nil {
		return err
	}

	return r.client.ZRem(r.ctx, queue+":reserved", payload).Err()
}

func (r *Redis) Release(queue string, job contractsqueue.Jobs, delay uint) error {
	payload, err := json.MarshalString(&redisData{
		Signature: job.Job.Signature(),
		Args:      job.Payloads,
	})
	if err != nil {
		return err
	}

	return r.lua.Release().Run(r.ctx, r.client, []string{queue + ":delayed", queue + ":delayed"}, payload, carbon.Now().AddSeconds(int(delay)).Timestamp()).Err()
}

func (r *Redis) Clear(queue string) error {
	return r.lua.Clear().Run(r.ctx, r.client, []string{queue, queue + ":delayed", queue + ":reserved", queue + ":notify"}).Err()
}

func (r *Redis) Size(queue string) (uint64, error) {
	return r.lua.Size().Run(r.ctx, r.client, []string{queue, queue + ":delayed", queue + ":reserved"}).Uint64()
}

func (r *Redis) migrate(queue string) {
	r.migrateExpiredJobs(queue+":delayed", queue)
	if r.retryAfter > 0 {
		r.migrateExpiredJobs(queue+":reserved", queue)
	}
}

func (r *Redis) migrateExpiredJobs(from, to string) {
	_ = r.lua.MigrateExpiredJobs().Run(r.ctx, r.client, []string{from, to, to + ":notify"}, carbon.Now().Timestamp()).Err()
}

func (r *Redis) retrieveNextJob(queue string, block ...bool) (contractsqueue.Jobs, error) {
	if len(block) == 0 {
		block = []bool{true}
	}

	raw, err := r.lua.Pop().Run(r.ctx, r.client, []string{queue, queue + ":reserved", queue + ":notify"}, carbon.Now().Timestamp()).Result()
	if err != nil {
		return contractsqueue.Jobs{}, err
	}
	value, ok := raw.([]any)
	if !ok {
		return contractsqueue.Jobs{}, fmt.Errorf("invalid return type: %T", raw)
	}
	if len(value) != 2 {
		return contractsqueue.Jobs{}, fmt.Errorf("invalid return value: %v", raw)
	}

	// If there is no job, we will block the worker until there is a job.
	if value[0] == nil || len(value[0].(string)) == 0 {
		if !block[0] {
			return contractsqueue.Jobs{}, errors.New("no job in queue")
		}
		err = r.client.BRPop(r.ctx, 0, queue+":notify").Err()
		if err != nil {
			return contractsqueue.Jobs{}, err
		}

		return r.retrieveNextJob(queue, false)
	}

	var job redisData
	if err = json.UnmarshalString(value[0].(string), &job); err != nil {
		return contractsqueue.Jobs{}, err
	}

	return contractsqueue.Jobs{
		Job:      JobRegistry[job.Signature],
		Payloads: job.Args,
	}, nil
}
