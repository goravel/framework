package driver

import (
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
	instrumentationdatabase "github.com/goravel/framework/telemetry/instrumentation/database"
)

type cachedConnection struct {
	db         *gorm.DB
	instrument *instrumentationdatabase.Instrument
}

var (
	connectionToDB     sync.Map
	connectionToDBLock sync.Mutex
	pingWarning        sync.Once
)

func BuildGorm(config config.Config, logger logger.Interface, pool database.Pool, connection string, telemetryResolver contractstelemetry.Resolver) (*gorm.DB, *instrumentationdatabase.Instrument, error) {
	if cached, ok := connectionToDB.Load(connection); ok {
		c := cached.(cachedConnection)
		return c.db, c.instrument, nil
	}

	if len(pool.Writers) == 0 {
		return nil, nil, errors.DatabaseConfigNotFound
	}

	// If the database is empty, it means the database is not configured, we don't want to return an error or print a warning here.
	if pool.Writers[0].Database == "" {
		return nil, nil, nil
	}

	connectionToDBLock.Lock()
	defer connectionToDBLock.Unlock()

	if cached, ok := connectionToDB.Load(connection); ok {
		c := cached.(cachedConnection)
		return c.db, c.instrument, nil
	}

	gormConfig := &gorm.Config{
		DisableAutomaticPing:                     true,
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger,
		NowFunc: func() time.Time {
			return carbon.Now().StdTime()
		},
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   pool.Writers[0].Prefix,
			SingularTable: pool.Writers[0].Singular,
			NoLowerCase:   pool.Writers[0].NoLowerCase,
			NameReplacer:  pool.Writers[0].NameReplacer,
		},
	}

	instance, err := gorm.Open(pool.Writers[0].Dialector, gormConfig)
	if err != nil {
		return nil, nil, err
	}
	if pinger, ok := instance.ConnPool.(interface{ Ping() error }); ok {
		if err = pinger.Ping(); err != nil {
			pingWarning.Do(func() {
				color.Warningln(err.Error())
			})
		}
	}

	var instrument *instrumentationdatabase.Instrument
	if telemetryResolver != nil && instrumentationdatabase.Enabled(config) {
		instrument = instrumentationdatabase.NewInstrument(pool, connection, telemetryResolver)
		if pluginErr := instance.Use(instrumentationdatabase.NewGormPlugin(instrument)); pluginErr != nil {
			color.Warningln("database telemetry: " + pluginErr.Error())
		}
	}

	maxIdleConns := config.GetInt("database.pool.max_idle_conns", 10)
	maxOpenConns := config.GetInt("database.pool.max_open_conns", 100)
	connMaxIdleTime := config.GetDuration("database.pool.conn_max_idletime", 3600)
	connMaxLifetime := config.GetDuration("database.pool.conn_max_lifetime", 3600)

	if len(pool.Writers) == 1 && len(pool.Readers) == 0 {
		db, err := instance.DB()
		if err != nil {
			return nil, nil, err
		}

		db.SetMaxIdleConns(maxIdleConns)
		db.SetMaxOpenConns(maxOpenConns)
		db.SetConnMaxIdleTime(connMaxIdleTime * time.Second)
		db.SetConnMaxLifetime(connMaxLifetime * time.Second)

		if instrument != nil {
			instrument.SetDB(db)
		}

		connectionToDB.Store(connection, cachedConnection{db: instance, instrument: instrument})

		return instance, instrument, nil
	}

	var (
		writeDialectors []gorm.Dialector
		readDialectors  []gorm.Dialector
	)

	for _, writer := range pool.Writers {
		writeDialectors = append(writeDialectors, writer.Dialector)
	}

	for _, reader := range pool.Readers {
		readDialectors = append(readDialectors, reader.Dialector)
	}

	if err := instance.Use(dbresolver.Register(dbresolver.Config{
		Sources:           writeDialectors,
		Replicas:          readDialectors,
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}).SetMaxIdleConns(maxIdleConns).
		SetMaxOpenConns(maxOpenConns).
		SetConnMaxLifetime(connMaxLifetime * time.Second).
		SetConnMaxIdleTime(connMaxIdleTime * time.Second)); err != nil {
		return nil, nil, err
	}

	if instrument != nil {
		if sqlDB, err := instance.DB(); err == nil {
			instrument.SetDB(sqlDB)
		}
	}

	connectionToDB.Store(connection, cachedConnection{db: instance, instrument: instrument})

	return instance, instrument, nil
}

func drainConnections() []cachedConnection {
	connectionToDBLock.Lock()
	defer connectionToDBLock.Unlock()

	var stale []cachedConnection
	connectionToDB.Range(func(key, value any) bool {
		if cached, ok := value.(cachedConnection); ok {
			stale = append(stale, cached)
		}
		connectionToDB.Delete(key)
		return true
	})

	return stale
}

// ResetConnections drops the cached connections so a later build reconnects with
// the current configuration, and unregisters their pool-metrics callbacks. The
// pools are left open because a caller may still hold the *gorm.DB; use
// CloseConnections when the connections are known to be idle.
func ResetConnections() {
	for _, cached := range drainConnections() {
		cached.instrument.Shutdown()
	}
}

// CloseConnections drops the cached connections, unregisters their callbacks and
// closes the pools. Call it only when the connections are idle (e.g. the database
// provider's Register, which runs after app.Restart has stopped the runners).
func CloseConnections() {
	for _, cached := range drainConnections() {
		cached.instrument.Shutdown()
		if cached.db == nil {
			continue
		}
		if sqlDB, err := cached.db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				color.Warningln("close database connection: " + err.Error())
			}
		}
	}
}
