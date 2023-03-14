package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gookit/color"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	databasegorm "github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/facades"
)

type Orm struct {
	ctx       context.Context
	instance  contractsorm.Query
	instances map[string]contractsorm.Query
}

func NewOrm(ctx context.Context) *Orm {
	defaultConnection := facades.Config.GetString("database.default")
	gormQuery, err := databasegorm.NewQuery(ctx, defaultConnection)
	if err != nil {
		color.Redln(fmt.Sprintf("[Orm] Init %s connection error: %v", defaultConnection, err))

		return nil
	}
	if gormQuery == nil {
		return nil
	}

	return &Orm{
		ctx:      ctx,
		instance: gormQuery,
		instances: map[string]contractsorm.Query{
			defaultConnection: gormQuery,
		},
	}
}

// DEPRECATED: use gorm.New()
func NewGormInstance(connection string) (*gorm.DB, error) {
	return databasegorm.New(connection)
}

func (r *Orm) Connection(name string) contractsorm.Orm {
	if name == "" {
		name = facades.Config.GetString("database.default")
	}
	if instance, exist := r.instances[name]; exist {
		return &Orm{
			ctx:       r.ctx,
			instance:  instance,
			instances: r.instances,
		}
	}

	gormDB, err := databasegorm.NewQuery(r.ctx, name)
	if err != nil || gormDB == nil {
		color.Redln(fmt.Sprintf("[Orm] Init %s connection error: %v", name, err))

		return nil
	}

	r.instances[name] = gormDB

	return &Orm{
		ctx:       r.ctx,
		instance:  gormDB,
		instances: r.instances,
	}
}

func (r *Orm) DB() (*sql.DB, error) {
	db := r.Query().(*databasegorm.Query)

	return db.Instance().DB()
}

func (r *Orm) Query() contractsorm.Query {
	return r.instance
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
	return NewOrm(ctx)
}
