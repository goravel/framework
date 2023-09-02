package gorm

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/wire"
	gormio "gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	"github.com/goravel/framework/contracts/config"
	databasecontract "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/database/db"
	"github.com/goravel/framework/support/carbon"
)

var GormSet = wire.NewSet(NewGormImpl, wire.Bind(new(Gorm), new(*GormImpl)))
var _ Gorm = &GormImpl{}

type Gorm interface {
	Make() (*gormio.DB, error)
}

type GormImpl struct {
	config     config.Config
	connection string
	dbConfig   db.Config
	dialector  Dialector
	instance   *gormio.DB
}

func NewGormImpl(config config.Config, connection string, dbConfig db.Config, dialector Dialector) *GormImpl {
	return &GormImpl{
		config:     config,
		connection: connection,
		dbConfig:   dbConfig,
		dialector:  dialector,
	}
}

func (r *GormImpl) Make() (*gormio.DB, error) {
	readConfigs := r.dbConfig.Reads()
	writeConfigs := r.dbConfig.Writes()
	if len(writeConfigs) == 0 {
		return nil, errors.New("not found database configuration")
	}

	writeDialectors, err := r.dialector.Make([]databasecontract.Config{writeConfigs[0]})
	if err != nil {
		return nil, fmt.Errorf("init gorm dialector error: %v", err)
	}

	if err := r.init(writeDialectors[0]); err != nil {
		return nil, err
	}

	if err := r.configurePool(); err != nil {
		return nil, err
	}

	if err := r.configureReadWriteSeparate(readConfigs, writeConfigs); err != nil {
		return nil, err
	}

	return r.instance, nil
}

func (r *GormImpl) configurePool() error {
	sqlDB, err := r.instance.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(r.config.GetInt("database.pool.max_idle_conns", 10))
	sqlDB.SetMaxOpenConns(r.config.GetInt("database.pool.max_open_conns", 100))
	sqlDB.SetConnMaxIdleTime(time.Duration(r.config.GetInt("database.pool.conn_max_idletime", 3600)) * time.Second)
	sqlDB.SetConnMaxLifetime(time.Duration(r.config.GetInt("database.pool.conn_max_lifetime", 3600)) * time.Second)

	return nil
}

func (r *GormImpl) configureReadWriteSeparate(readConfigs, writeConfigs []databasecontract.Config) error {
	if len(readConfigs) == 0 || len(writeConfigs) == 0 {
		return nil
	}

	readDialectors, err := r.dialector.Make(readConfigs)
	if err != nil {
		return err
	}

	writeDialectors, err := r.dialector.Make(writeConfigs)
	if err != nil {
		return err
	}

	return r.instance.Use(dbresolver.Register(dbresolver.Config{
		Sources:           writeDialectors,
		Replicas:          readDialectors,
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}))
}

func (r *GormImpl) init(dialector gormio.Dialector) error {
	var logLevel gormlogger.LogLevel
	if r.config.GetBool("app.debug") {
		logLevel = gormlogger.Info
	} else {
		logLevel = gormlogger.Error
	}

	logger := NewLogger(log.New(os.Stdout, "\r\n", log.LstdFlags), gormlogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormlogger.Info,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})
	instance, err := gormio.Open(dialector, &gormio.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger.LogMode(logLevel),
		NowFunc: func() time.Time {
			return carbon.Now().ToStdTime()
		},
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   r.config.GetString(fmt.Sprintf("database.connections.%s.prefix", r.connection)),
			SingularTable: r.config.GetBool(fmt.Sprintf("database.connections.%s.singular", r.connection)),
		},
	})
	if err != nil {
		return err
	}

	r.instance = instance

	return nil
}
