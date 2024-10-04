package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/goravel/framework/contracts/config"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/orm"
)

type Orm struct {
	ctx        context.Context
	config     config.Config
	connection string
	log        log.Log
	query      contractsorm.Query
	queries    map[string]contractsorm.Query
	refresh    func(key any)
}

func NewOrm(
	ctx context.Context,
	config config.Config,
	connection string,
	query contractsorm.Query,
	queries map[string]contractsorm.Query,
	log log.Log,
	refresh func(key any),
) *Orm {
	return &Orm{
		ctx:        ctx,
		config:     config,
		connection: connection,
		log:        log,
		query:      query,
		queries:    queries,
		refresh:    refresh,
	}
}

func BuildOrm(ctx context.Context, config config.Config, connection string, log log.Log, refresh func(key any)) (*Orm, error) {
	query, err := gorm.BuildQuery(ctx, config, connection, log)
	if err != nil {
		return nil, fmt.Errorf("[Orm] Build query for %s connection error: %v", connection, err)
	}

	queries := map[string]contractsorm.Query{
		connection: query,
	}

	return NewOrm(ctx, config, connection, query, queries, log, refresh), nil
}

func (r *Orm) Connection(name string) contractsorm.Orm {
	if name == "" {
		name = r.config.GetString("database.default")
	}
	if instance, exist := r.queries[name]; exist {
		return NewOrm(r.ctx, r.config, name, instance, r.queries, r.log, r.refresh)
	}

	query, err := gorm.BuildQuery(r.ctx, r.config, name, r.log)
	if err != nil || query == nil {
		r.log.Errorf("[Orm] Init %s connection error: %v", name, err)

		return NewOrm(r.ctx, r.config, name, nil, r.queries, r.log, r.refresh)
	}

	r.queries[name] = query

	return NewOrm(r.ctx, r.config, name, query, r.queries, r.log, r.refresh)
}

func (r *Orm) DB() (*sql.DB, error) {
	query, ok := r.Query().(*gorm.Query)
	if !ok {
		return nil, fmt.Errorf("unexpected Query type %T, expected *gorm.Query", r.Query())
	}

	return query.Instance().DB()
}

func (r *Orm) Query() contractsorm.Query {
	return r.query
}

func (r *Orm) Factory() contractsorm.Factory {
	return NewFactoryImpl(r.Query())
}

func (r *Orm) Observe(model any, observer contractsorm.Observer) {
	orm.Observers = append(orm.Observers, orm.Observer{
		Model:    model,
		Observer: observer,
	})
}

func (r *Orm) Refresh() {
	r.refresh(BindingOrm)
}

func (r *Orm) Transaction(txFunc func(tx contractsorm.Query) error) error {
	tx, err := r.Query().Begin()
	if err != nil {
		return err
	}

	if err := txFunc(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("rollback error: %v", err)
		}

		return err
	} else {
		return tx.Commit()
	}
}

func (r *Orm) WithContext(ctx context.Context) contractsorm.Orm {
	for _, query := range r.queries {
		if gormQuery, ok := query.(*gorm.Query); ok {
			gormQuery.SetContext(ctx)
		}
	}

	if gormQuery, ok := r.query.(*gorm.Query); ok {
		gormQuery.SetContext(ctx)
	}

	return &Orm{
		ctx:        ctx,
		config:     r.config,
		connection: r.connection,
		query:      r.query,
		queries:    r.queries,
	}
}
