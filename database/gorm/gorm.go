package gorm

import (
	"log"
	"os"
	"time"

	gormio "gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
)

type Builder struct {
	config        config.Config
	configBuilder database.ConfigBuilder
	instance      *gormio.DB
}

func NewGorm(config config.Config, configBuilder database.ConfigBuilder) (*gormio.DB, error) {
	builder := &Builder{
		config:        config,
		configBuilder: configBuilder,
	}

	return builder.Build()
}

func (r *Builder) Build() (*gormio.DB, error) {
	readConfigs := r.configBuilder.Reads()
	writeConfigs := r.configBuilder.Writes()
	if len(writeConfigs) == 0 {
		return nil, errors.OrmDatabaseConfigNotFound
	}

	if err := r.init(writeConfigs[0]); err != nil {
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

func (r *Builder) configurePool() error {
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

func (r *Builder) configureReadWriteSeparate(readConfigs, writeConfigs []database.FullConfig) error {
	if len(readConfigs) == 0 || len(writeConfigs) == 0 {
		return nil
	}

	readDialectors, err := getDialectors(readConfigs)
	if err != nil {
		return err
	}

	writeDialectors, err := getDialectors(writeConfigs)
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

func (r *Builder) init(fullConfig database.FullConfig) error {
	dialectors, err := getDialectors([]database.FullConfig{fullConfig})
	if err != nil {
		return err
	}
	if len(dialectors) == 0 {
		return errors.OrmNoDialectorsFound
	}

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
	instance, err := gormio.Open(dialectors[0], &gormio.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger.LogMode(logLevel),
		NowFunc: func() time.Time {
			return carbon.Now().StdTime()
		},
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   fullConfig.Prefix,
			SingularTable: fullConfig.Singular,
		},
	})
	if err != nil {
		return err
	}

	r.instance = instance

	return nil
}
