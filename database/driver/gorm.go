package postgres

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

type Gorm struct {
	config config.Config
	log    log.Log
	pool   database.Pool
}

func BuildGorm(config config.Config, log log.Log, pool database.Pool) (*gorm.DB, error) {
	g := &Gorm{config: config, log: log, pool: pool}
	instance, err := g.instance()
	if err != nil {
		return nil, err
	}
	if err := g.configurePool(instance); err != nil {
		return nil, err
	}
	if err := g.configureReadWriteSeparate(instance); err != nil {
		return nil, err
	}

	return instance, nil
}

func (r *Gorm) configurePool(instance *gorm.DB) error {
	db, err := instance.DB()
	if err != nil {
		return err
	}

	config := r.pool.Config()
	db.SetMaxIdleConns(config.GetInt("database.pool.max_idle_conns", 10))
	db.SetMaxOpenConns(config.GetInt("database.pool.max_open_conns", 100))
	db.SetConnMaxIdleTime(time.Duration(config.GetInt("database.pool.conn_max_idletime", 3600)) * time.Second)
	db.SetConnMaxLifetime(time.Duration(config.GetInt("database.pool.conn_max_lifetime", 3600)) * time.Second)

	return nil
}

func (r *Gorm) configureReadWriteSeparate(instance *gorm.DB) error {
	writers, readers, err := r.writeAndReadDialectors()
	if err != nil {
		return err
	}

	return instance.Use(dbresolver.Register(dbresolver.Config{
		Sources:           writers,
		Replicas:          readers,
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}))
}

func (r *Gorm) gormConfig() *gorm.Config {
	logger := logger.NewLogger(r.config, r.log).ToGorm()
	writeConfigs := r.pool.Writes()
	if len(writeConfigs) == 0 {
		return nil
	}

	return &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger,
		NowFunc: func() time.Time {
			return carbon.Now().StdTime()
		},
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   writeConfigs[0].Prefix,
			SingularTable: writeConfigs[0].Singular,
			NoLowerCase:   writeConfigs[0].NoLowerCase,
			NameReplacer:  writeConfigs[0].NameReplacer,
		},
	}
}

func (r *Gorm) instance() (*gorm.DB, error) {
	if len(r.pool.Writers) == 0 {
		return nil, errors.DatabaseConfigNotFound
	}

	logger := logger.NewLogger(r.pool.Config(), r.log).ToGorm()
	config := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger,
		NowFunc: func() time.Time {
			return carbon.Now().StdTime()
		},
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   writeConfigs[0].Prefix,
			SingularTable: writeConfigs[0].Singular,
			NoLowerCase:   writeConfigs[0].NoLowerCase,
			NameReplacer:  writeConfigs[0].NameReplacer,
		},
	}

	instance, err := gorm.Open(r.pool.Writers[0].Dialector, r.gormConfig())
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (r *Gorm) writeAndReadDialectors() (writers []gorm.Dialector, readers []gorm.Dialector, err error) {
	writeConfigs := r.pool.Writes()
	readConfigs := r.pool.Reads()

	writers, err = r.configsToDialectors(writeConfigs)
	if err != nil {
		return nil, nil, err
	}
	readers, err = r.configsToDialectors(readConfigs)
	if err != nil {
		return nil, nil, err
	}

	return writers, readers, nil
}
