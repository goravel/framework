package driver

import (
	"github.com/goravel/framework/contracts/queue"
)

type Sync struct {
	connection string
}

func NewSync(connection string) *Sync {
	return &Sync{
		connection: connection,
	}
}

func (receiver *Sync) ConnectionName() string {
	return receiver.connection
}

func (receiver *Sync) Push(job queue.Job, args []queue.Arg) error {
	return nil
}

func (receiver *Sync) Bulk(jobs []queue.Job) error {
	return nil
}

func (receiver *Sync) Later(job queue.Job, delay int) error {
	return nil
}

func (receiver *Sync) Pop() (queue.Job, error) {
	return nil, nil
}

func (receiver *Sync) Delete(job queue.Job) error {
	return nil
}

func (receiver *Sync) Release(job queue.Job, delay int) error {
	return nil
}

func (receiver *Sync) Clear() error {
	return nil
}

func (receiver *Sync) Size() (int, error) {
	return 0, nil
}

func (receiver *Sync) Server(concurrent int) {
}
