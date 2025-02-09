package db

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/db"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/errors"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	instance *sqlx.DB
}

func NewDB(instance *sqlx.DB) db.DB {
	return &DB{instance: instance}
}

func BuildDB(config config.Config, connection string) (db.DB, error) {
	driverCallback, exist := config.Get(fmt.Sprintf("database.connections.%s.via", connection)).(func() (contractsdriver.Driver, error))
	if !exist {
		return nil, errors.DatabaseConfigNotFound
	}

	driver, err := driverCallback()
	if err != nil {
		return nil, err
	}

	instance, err := driver.DB()
	if err != nil {
		return nil, err
	}

	return &DB{instance: instance}, nil
}

func (r *DB) Table(name string) db.Query {
	return NewQuery(r.instance, name)
}
