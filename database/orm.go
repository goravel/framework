package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gookit/color"
	"github.com/pkg/errors"

	"github.com/goravel/framework/contracts/config"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	databasegorm "github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/orm"
)

type OrmImpl struct {
	ctx        context.Context
	config     config.Config
	connection string
	query      ormcontract.Query
	queries    map[string]ormcontract.Query
}

func NewOrmImpl(ctx context.Context, config config.Config, connection string, query ormcontract.Query) (*OrmImpl, error) {
	return &OrmImpl{
		ctx:        ctx,
		config:     config,
		connection: connection,
		query:      query,
		queries: map[string]ormcontract.Query{
			connection: query,
		},
	}, nil
}

func (r *OrmImpl) Connection(name string) ormcontract.Orm {
	if name == "" {
		name = r.config.GetString("database.default")
	}
	if instance, exist := r.queries[name]; exist {
		return &OrmImpl{
			ctx:     r.ctx,
			query:   instance,
			queries: r.queries,
		}
	}

	queue, err := databasegorm.InitializeQuery(r.ctx, r.config, name)
	if err != nil || queue == nil {
		color.Redln(fmt.Sprintf("[Orm] Init %s connection error: %v", name, err))

		return nil
	}

	r.queries[name] = queue

	return &OrmImpl{
		ctx:     r.ctx,
		query:   queue,
		queries: r.queries,
	}
}

func (r *OrmImpl) DB() (*sql.DB, error) {
	db := r.Query().(*databasegorm.QueryImpl)

	return db.Instance().DB()
}

func (r *OrmImpl) Query() ormcontract.Query {
	return r.query
}

func (r *OrmImpl) Factory() ormcontract.Factory {
	return NewFactoryImpl(r.Query())
}

func (r *OrmImpl) Observe(model any, observer ormcontract.Observer) {
	orm.Observers = append(orm.Observers, orm.Observer{
		Model:    model,
		Observer: observer,
	})
}

func (r *OrmImpl) Transaction(txFunc func(tx ormcontract.Transaction) error) error {
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

func (r *OrmImpl) WithContext(ctx context.Context) ormcontract.Orm {
	instance, _ := NewOrmImpl(ctx, r.config, r.connection, r.query)

	return instance
}
