package queue

import (
	"context"

	"github.com/redis/go-redis/v9"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/json"
)

type Redis struct {
	connection   string
	defaultQueue string
	retryAfter   uint
	client       *redis.Client
	lua          RedisLua
	ctx          context.Context
}

func NewRedis(connection string, client *redis.Client) *Redis {
	return &Redis{
		connection:   connection,
		defaultQueue: "default",
		retryAfter:   60,
		client:       client,
		ctx:          context.Background(),
	}
}

func (r *Redis) ConnectionName() string {
	return r.connection
}

func (r *Redis) DriverName() string {
	return DriverRedis
}

func (r *Redis) Push(job contractsqueue.Job, payloads []contractsqueue.Payloads, queue string) error {
	queue = r.getQueue(queue)
	data, err := json.Marshal(contractsqueue.Jobs{
		Job:      job,
		Payloads: payloads,
	})
	if err != nil {
		return err
	}

	return r.lua.Push().Run(r.ctx, r.client, []string{queue, queue + ":notify"}, data).Err()
}

func (r *Redis) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	queue = r.getQueue(queue)
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

func (r *Redis) Later(delay uint, job contractsqueue.Job, payloads []contractsqueue.Payloads, queue string) error {
	data, err := json.Marshal(contractsqueue.Jobs{
		Job:      job,
		Payloads: payloads,
	})
	if err != nil {
		return err
	}

	return r.client.ZAdd(r.ctx, r.getQueue(queue)+":delayed", redis.Z{
		Score:  float64(carbon.Now().AddSeconds(int(delay)).Timestamp()),
		Member: data,
	}).Err()
}

func (r *Redis) Pop(queue string) (contractsqueue.Job, []contractsqueue.Payloads, error) {
	prefixed := r.getQueue(queue)
	r.migrate(prefixed)

	job, err := r.retrieveNextJob(prefixed)
	if err != nil {
		return nil, nil, err
	}

	return job.Job, job.Payloads, nil
}

func (r *Redis) Delete(queue string, job contractsqueue.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return r.client.ZRem(r.ctx, r.getQueue(queue)+":reserved", data).Err()
}

func (r *Redis) Release(queue string, job contractsqueue.Job, delay uint) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	queue = r.getQueue(queue)
	return r.lua.Release().Run(r.ctx, r.client, []string{queue + ":delayed", queue + ":delayed"}, data, carbon.Now().AddSeconds(int(delay)).Timestamp()).Err()
}

func (r *Redis) Clear(queue string) error {
	queue = r.getQueue(queue)
	return r.lua.Clear().Run(r.ctx, r.client, []string{queue, queue + ":delayed", queue + ":reserved", queue + ":notify"}).Err()
}

func (r *Redis) getQueue(queue string) string {
	if len(queue) == 0 {
		return "queues:" + r.defaultQueue
	}

	return "queues:" + queue
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
	r.lua.MigrateExpiredJobs().Run(r.ctx, r.client, []string{from, to, to + ":notify"}, carbon.Now().Timestamp()).Result()
}

func (r *Redis) retrieveNextJob(queue string, block ...bool) (contractsqueue.Jobs, error) {
	if len(block) == 0 {
		block = []bool{true}
	}

	raw, err := r.lua.Pop().Run(r.ctx, r.client, []string{queue, queue + ":reserved", queue + ":notify"}, carbon.Now().Timestamp()).Result()
	if err != nil {
		return contractsqueue.Jobs{}, err
	}

	var job contractsqueue.Jobs
	if err = json.Unmarshal([]byte(raw.([]interface{})[0].(string)), &job); err != nil {
		return contractsqueue.Jobs{}, err
	}

	// If there is no job, we will block the worker until there is a job.
	if job.Job == nil && job.Payloads == nil && block[0] {
		err = r.client.BRPop(r.ctx, 0, queue+":notify").Err()
		if err != nil {
			return contractsqueue.Jobs{}, err
		}

		return r.retrieveNextJob(queue, false)
	}

	return job, nil
}
