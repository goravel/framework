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

func (d *Database) ConnectionName() string {
	return d.connection
}

func (d *Database) Push(job queue.Job) error {
	return nil
}

func (d *Database) Bulk(jobs []queue.Job) error {
	return nil
}

func (d *Database) Later(job queue.Job, delay int) error {
	return nil
}

func (d *Database) Pop() (queue.Job, error) {
	return nil, nil
}

func (d *Database) Delete(job queue.Job) error {
	return nil
}

func (d *Database) Release(job queue.Job, delay int) error {
	return nil
}

func (d *Database) Clear() error {
	return nil
}

func (d *Database) Size() (int, error) {
	return 0, nil
}
