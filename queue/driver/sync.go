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

func (receiver *Sync) Push(job queue.Job, args []queue.Arg, queue string) error {
	return nil
}

func (receiver *Sync) Bulk(jobs []queue.Jobs, queue string) error {
	return nil
}

func (receiver *Sync) Later(delay int, job queue.Job, args []queue.Arg, queue string) error {
	return nil
}

func (receiver *Sync) Pop(queue string) (queue.Job, []queue.Arg, error) {
	return nil, nil, nil
}

func (receiver *Sync) Delete(queue string, job queue.Job) error {
	return nil
}

func (receiver *Sync) Release(queue string, job queue.Job, delay int) error {
	return nil
}

func (receiver *Sync) Clear(queue string) error {
	return nil
}

func (receiver *Sync) Size(queue string) (int64, error) {
	return 0, nil
}

func (receiver *Sync) Server(concurrent int, queue string) {

}
