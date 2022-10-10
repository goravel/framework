package database

import (
	"context"

	"github.com/pkg/errors"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

type Orm struct {
	ctx             context.Context
	connection      string
	defaultInstance contractsorm.DB
	instances       map[string]contractsorm.DB
}

func NewOrm() contractsorm.Orm {
	orm := &Orm{}

	return orm.Connection("")
}

func (r *Orm) Connection(name string) contractsorm.Orm {
	defaultConnection := facades.Config.GetString("database.default")
	if name == "" {
		name = defaultConnection
	}

	r.connection = name
	if r.instances == nil {
		r.instances = make(map[string]contractsorm.DB)
	}

	if _, exist := r.instances[name]; exist {
		return r
	}

	gorm, err := NewGormDB(r.ctx, name)
	if err != nil {
		facades.Log.Errorf("init connection error: %v", err)

		return r
	}

	r.instances[name] = gorm

	if name == defaultConnection {
		r.defaultInstance = gorm
	}

	return r
}

func (r *Orm) Query() contractsorm.DB {
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

func (r *Orm) Transaction(txFunc func(tx contractsorm.Transaction) error) error {
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

func (r *Orm) WithContext(ctx context.Context) contractsorm.Orm {
	r.ctx = ctx

	return r
}
