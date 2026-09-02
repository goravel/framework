package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/driver"
	"github.com/goravel/framework/database/factory"
	"github.com/goravel/framework/database/gorm"
)

// QueriesCache is the shared, concurrency-safe registry of connections built
// from the same database config. Every Orm instance created from the same
// connection pool shares one QueriesCache, so concurrent Connection(name) calls
// from different instances (e.g. lazily building the same non-default
// connection from multiple goroutines) are serialized by the internal mutex.
type QueriesCache struct {
	mu          sync.RWMutex
	connections map[string]cachedConnection
}

// cachedConnection holds a built connection: the Query handle and the database
// config it was built from.
type cachedConnection struct {
	query    contractsorm.Query
	dbConfig database.Config
}

// NewQueriesCache creates a QueriesCache from per-connection queries and
// configs. queries and dbConfigs must be keyed identically by connection name.
func NewQueriesCache(queries map[string]contractsorm.Query, dbConfigs map[string]database.Config) *QueriesCache {
	cache := &QueriesCache{connections: make(map[string]cachedConnection, len(queries))}
	for name, query := range queries {
		cache.connections[name] = cachedConnection{query: query, dbConfig: dbConfigs[name]}
	}

	return cache
}

type Orm struct {
	ctx             context.Context
	config          config.Config
	log             log.Log
	query           contractsorm.Query
	queries         *QueriesCache
	fresh           func(key ...any)
	connection      string
	modelToObserver []contractsorm.ModelToObserver
	dbConfig        database.Config
	mutex           sync.Mutex
}

// NewOrm creates a new Orm instance for the given connection.
//
// queries is the shared, concurrency-safe registry of built connections and
// must be non-nil (construct it via NewQueriesCache). It should contain an
// entry for the connection (extended lazily by Connection()).
func NewOrm(
	ctx context.Context,
	config config.Config,
	connection string,
	dbConfig database.Config,
	query contractsorm.Query,
	queries *QueriesCache,
	log log.Log,
	modelToObserver []contractsorm.ModelToObserver,
	fresh func(key ...any),
) *Orm {
	return &Orm{
		ctx:             ctx,
		config:          config,
		connection:      connection,
		dbConfig:        dbConfig,
		log:             log,
		modelToObserver: modelToObserver,
		query:           query,
		queries:         queries,
		fresh:           fresh,
	}
}

func BuildOrm(ctx context.Context, config config.Config, connection string, log log.Log, fresh func(key ...any)) (*Orm, error) {
	query, dbConfig, err := gorm.BuildQuery(ctx, config, connection, log, nil)
	if err != nil {
		return NewOrm(ctx, config, connection, dbConfig, nil, NewQueriesCache(nil, nil), log, nil, fresh), err
	}

	return NewOrm(ctx, config, connection, dbConfig, query, NewQueriesCache(map[string]contractsorm.Query{connection: query}, map[string]database.Config{connection: dbConfig}), log, nil, fresh), nil
}

func (r *Orm) Config() database.Config {
	return r.dbConfig
}

func (r *Orm) Connection(name string) contractsorm.Orm {
	if name == "" {
		name = r.config.GetString("database.default")
	}

	r.queries.mu.RLock()
	instance, exist := r.queries.connections[name]
	r.queries.mu.RUnlock()
	if exist {
		return NewOrm(r.ctx, r.config, name, instance.dbConfig, instance.query, r.queries, r.log, r.modelToObserver, r.fresh)
	}

	query, dbConfig, err := gorm.BuildQuery(r.ctx, r.config, name, r.log, r.modelToObserver)
	if err != nil || query == nil {
		r.log.Errorf("[Orm] Init %s connection error: %v", name, err)

		return NewOrm(r.ctx, r.config, name, dbConfig, nil, r.queries, r.log, r.modelToObserver, r.fresh)
	}

	r.queries.mu.Lock()
	defer r.queries.mu.Unlock()
	// Double-check: another goroutine may have built this connection first.
	if instance, exist := r.queries.connections[name]; exist {
		return NewOrm(r.ctx, r.config, name, instance.dbConfig, instance.query, r.queries, r.log, r.modelToObserver, r.fresh)
	}

	r.queries.connections[name] = cachedConnection{query: query, dbConfig: dbConfig}

	return NewOrm(r.ctx, r.config, name, dbConfig, query, r.queries, r.log, r.modelToObserver, r.fresh)
}

func (r *Orm) DB() (*sql.DB, error) {
	return r.query.DB()
}

func (r *Orm) Factory() contractsorm.Factory {
	return factory.NewFactoryImpl(r.Query())
}

func (r *Orm) DatabaseName() string {
	return r.dbConfig.Database
}

func (r *Orm) Name() string {
	return r.connection
}

func (r *Orm) Observe(model any, observer contractsorm.Observer) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.modelToObserver = append(r.modelToObserver, contractsorm.ModelToObserver{
		Model:    model,
		Observer: observer,
	})

	r.queries.mu.RLock()
	queries := make([]contractsorm.Query, 0, len(r.queries.connections))
	for _, connection := range r.queries.connections {
		queries = append(queries, connection.query)
	}
	r.queries.mu.RUnlock()

	for _, query := range queries {
		if queryWithObserver, ok := query.(contractsorm.QueryWithObserver); ok {
			queryWithObserver.Observe(model, observer)
		}
	}

	if queryWithObserver, ok := r.query.(contractsorm.QueryWithObserver); ok {
		queryWithObserver.Observe(model, observer)
	}
}

func (r *Orm) Query() contractsorm.Query {
	if r.ctx != context.Background() {
		if queryWithContext, ok := r.query.(contractsorm.QueryWithContext); ok {
			return queryWithContext.WithContext(r.ctx)
		}
	}

	return r.query
}

func (r *Orm) SetQuery(query contractsorm.Query) {
	r.query = query
}

// TODO: The fresh logic needs to be optimized, it's a bit unclear now.
// https://github.com/goravel/goravel/issues/848
func (r *Orm) Fresh() {
	r.fresh(binding.Orm)
	driver.ResetConnections()
}

func (r *Orm) Transaction(txFunc func(tx contractsorm.Query) error) (err error) {
	tx, err := r.Query().Begin()
	if err != nil {
		return err
	}

	defer func() {
		if re := recover(); re != nil {
			err = fmt.Errorf("panic: %v", re)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Join(err, rollbackErr)
			}
		}
	}()

	if err := txFunc(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}

		return err
	} else {
		return tx.Commit()
	}
}

func (r *Orm) WithContext(ctx context.Context) contractsorm.Orm {
	return NewOrm(ctx, r.config, r.connection, r.dbConfig, r.query, r.queries, r.log, r.modelToObserver, r.fresh)
}
