package driver

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/logger"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
)

func BuildGorm(config config.Config, log log.Log, pool database.Pool) (*gorm.DB, error) {
	if len(pool.Writers) == 0 {
		return nil, errors.DatabaseConfigNotFound
	}

	logger := logger.NewLogger(config, log).ToGorm()
	gormConfig := &gorm.Config{
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
		return nil, err
	}

	maxIdleConns := config.GetInt("database.pool.max_idle_conns", 10)
	maxOpenConns := config.GetInt("database.pool.max_open_conns", 100)
	connMaxIdleTime := config.GetDuration("database.pool.conn_max_idletime", 3600)
	connMaxLifetime := config.GetDuration("database.pool.conn_max_lifetime", 3600)

	if len(pool.Writers) == 1 && len(pool.Readers) == 0 {
		db, err := instance.DB()
		if err != nil {
			return nil, err
		}

		db.SetMaxIdleConns(maxIdleConns)
		db.SetMaxOpenConns(maxOpenConns)
		db.SetConnMaxIdleTime(connMaxIdleTime * time.Second)
		db.SetConnMaxLifetime(connMaxLifetime * time.Second)

		return instance, nil
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
		return nil, err
	}

	return instance, nil
}
