package driver

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/queue"
)

type Database struct {
	connection string
	db         orm.Orm
}

func NewDatabase(connection string, db orm.Orm) *Database {
	return &Database{
		connection: connection,
		db:         db,
	}
}

func (receiver *Database) ConnectionName() string {
	return receiver.connection
}

func (receiver *Database) Push(job queue.Job, args []queue.Arg) error {
	return nil
}

func (receiver *Database) Bulk(jobs []queue.Job) error {
	return nil
}

func (receiver *Database) Later(job queue.Job, delay int) error {
	return nil
}

func (receiver *Database) Pop() (queue.Job, error) {
	return nil, nil
}

func (receiver *Database) Delete(job queue.Job) error {
	return nil
}

func (receiver *Database) Release(job queue.Job, delay int) error {
	return nil
}

func (receiver *Database) Clear() error {
	return nil
}

func (receiver *Database) Size() (int, error) {
	return 0, nil
}

func (receiver *Database) Server(concurrent int) {

}
