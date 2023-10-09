package driver

import (
	"github.com/goravel/framework/contracts/queue"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	connection string
	client     *redis.Client
}

func NewRedis(connection string, client *redis.Client) *Redis {
	return &Redis{
		connection: connection,
		client:     client,
	}
}

func (receiver *Redis) ConnectionName() string {
	return receiver.connection
}

func (receiver *Redis) Push(job queue.Job, args []queue.Arg, queue string) error {
	return nil
}

func (receiver *Redis) Bulk(jobs []queue.Jobs, queue string) error {
	return nil
}

func (receiver *Redis) Later(delay int, job queue.Job, args []queue.Arg, queue string) error {
	return nil
}

func (receiver *Redis) Pop(queue string) (queue.Job, []queue.Arg, error) {
	return nil, nil, nil
}

func (receiver *Redis) Delete(queue string, job queue.Job) error {
	return nil
}

func (receiver *Redis) Release(queue string, job queue.Job, delay int) error {
	return nil
}

func (receiver *Redis) Clear(queue string) error {
	return nil
}

func (receiver *Redis) Size(queue string) (int64, error) {
	return 0, nil
}

func (receiver *Redis) Server(concurrent int, queue string) {

}
