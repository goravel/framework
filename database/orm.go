package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gookit/color"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	databasegorm "github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/facades"
)

type Orm struct {
	ctx             context.Context
	connection      string
	defaultInstance ormcontract.DB
	instances       map[string]ormcontract.DB
}

func NewOrm(ctx context.Context) *Orm {
	return &Orm{ctx: ctx}
}

// DEPRECATED: use gorm.New()
func NewGormInstance(connection string) (*gorm.DB, error) {
	return databasegorm.New(connection)
}

func (r *Orm) Connection(name string) ormcontract.Orm {
	defaultConnection := facades.Config.GetString("database.default")
	if name == "" {
		name = defaultConnection
	}

	r.connection = name
	if r.instances == nil {
		r.instances = make(map[string]ormcontract.DB)
	}

	if instance, exist := r.instances[name]; exist {
		if name == defaultConnection && r.defaultInstance == nil {
			r.defaultInstance = instance
		}

		return r
	}

	gormDB, err := databasegorm.NewDB(r.ctx, name)
	if err != nil {
		color.Redln(fmt.Sprintf("[Orm] Init connection error, %v", err))

		return nil
	}
	if gormDB == nil {
		return nil
	}

	r.instances[name] = gormDB

	if name == defaultConnection {
		r.defaultInstance = gormDB
	}

	return r
}

func (r *Orm) DB() (*sql.DB, error) {
	db := r.Query().(*databasegorm.DB)

	return db.Instance().DB()
}

func (r *Orm) Query() ormcontract.DB {
	if r.connection == "" {
		if r.defaultInstance == nil {
			r.Connection("")
		}

		return r.defaultInstance
	}

	instance, exist := r.instances[r.connection]
	if !exist {
		return nil
	}

	r.connection = ""

	return instance
}

func (r *Orm) Transaction(txFunc func(tx ormcontract.Transaction) error) error {
	tx, err := r.Query().Begin()
	if err != nil {
		return err
	}

	if err := txFunc(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "rollback error: %v", err)
		}

		return err
	} else {
		return tx.Commit()
	}
}

func (r *Orm) WithContext(ctx context.Context) ormcontract.Orm {
	return NewOrm(ctx)
}
