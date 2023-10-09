package queue

import (
	"github.com/goravel/framework/contracts/queue"
)

type ASync struct {
	connection string
}

func NewASync(connection string) *ASync {
	return &ASync{
		connection: connection,
	}
}

func (receiver *ASync) ConnectionName() string {
	return receiver.connection
}

func (receiver *ASync) Push(job queue.Job, args []queue.Arg, queue string) error {
	return nil
}

func (receiver *ASync) Bulk(jobs []queue.Jobs, queue string) error {
	return nil
}

func (receiver *ASync) Later(delay int, job queue.Job, args []queue.Arg, queue string) error {
	return nil
}

func (receiver *ASync) Pop(queue string) (queue.Job, []queue.Arg, error) {
	return nil, nil, nil
}

func (receiver *ASync) Delete(queue string, job queue.Job) error {
	return nil
}

func (receiver *ASync) Release(queue string, job queue.Job, delay int) error {
	return nil
}

func (receiver *ASync) Clear(queue string) error {
	return nil
}

func (receiver *ASync) Size(queue string) (int64, error) {
	return 0, nil
}

func (receiver *ASync) Server(concurrent int, queue string) {

}
