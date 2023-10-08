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

func (receiver *Redis) Push(job queue.Job, args []queue.Arg) error {
	return nil
}

func (receiver *Redis) Bulk(jobs []queue.Jobs) error {
	return nil
}

func (receiver *Redis) Later(job queue.Job, delay int) error {
	return nil
}

func (receiver *Redis) Pop() (queue.Job, error) {
	return nil, nil
}

func (receiver *Redis) Delete(job queue.Job) error {
	return nil
}

func (receiver *Redis) Release(job queue.Job, delay int) error {
	return nil
}

func (receiver *Redis) Clear() error {
	return nil
}

func (receiver *Redis) Size() (int, error) {
	return 0, nil
}

func (receiver *Redis) Server(concurrent int) {
}
