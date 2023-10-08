package driver

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

func (receiver *ASync) Push(job queue.Job, args []queue.Arg) error {
	return nil
}

func (receiver *ASync) Bulk(jobs []queue.Job) error {
	return nil
}

func (receiver *ASync) Later(job queue.Job, delay int) error {
	return nil
}

func (receiver *ASync) Pop() (queue.Job, error) {
	return nil, nil
}

func (receiver *ASync) Delete(job queue.Job) error {
	return nil
}

func (receiver *ASync) Release(job queue.Job, delay int) error {
	return nil
}

func (receiver *ASync) Clear() error {
	return nil
}

func (receiver *ASync) Size() (int, error) {
	return 0, nil
}

func (receiver *ASync) Server(concurrent int) {
}
