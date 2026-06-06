package database

import (
	"context"
	"database/sql"
	"sync"
	"time"

	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/goravel/framework/support/color"
)

type contextWrapper struct {
	context.Context
	parent context.Context
	start  time.Time
}

type GormPlugin struct {
	sqlDB      *sql.DB
	driverName string
	poolName   string
	poolOnce   sync.Once
}

func NewGormPlugin() *GormPlugin {
	return &GormPlugin{}
}

func NewGormPluginWithPool(sqlDB *sql.DB, driverName, poolName string) *GormPlugin {
	return &GormPlugin{sqlDB: sqlDB, driverName: driverName, poolName: poolName}
}

func (r *GormPlugin) Name() string {
	return "goravel:telemetry"
}

func (r *GormPlugin) Initialize(db *gorm.DB) error {
	if err := db.Callback().Create().Before("gorm:create").Register("goravel:telemetry:before_create", r.before("gorm.Create")); err != nil {
		return err
	}
	if err := db.Callback().Create().After("gorm:create").Register("goravel:telemetry:after_create", r.after); err != nil {
		return err
	}
	if err := db.Callback().Query().Before("gorm:query").Register("goravel:telemetry:before_query", r.before("gorm.Query")); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gorm:query").Register("goravel:telemetry:after_query", r.after); err != nil {
		return err
	}
	if err := db.Callback().Update().Before("gorm:update").Register("goravel:telemetry:before_update", r.before("gorm.Update")); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register("goravel:telemetry:after_update", r.after); err != nil {
		return err
	}
	if err := db.Callback().Delete().Before("gorm:delete").Register("goravel:telemetry:before_delete", r.before("gorm.Delete")); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").Register("goravel:telemetry:after_delete", r.after); err != nil {
		return err
	}
	if err := db.Callback().Row().Before("gorm:row").Register("goravel:telemetry:before_row", r.before("gorm.Row")); err != nil {
		return err
	}
	if err := db.Callback().Row().After("gorm:row").Register("goravel:telemetry:after_row", r.after); err != nil {
		return err
	}
	if err := db.Callback().Raw().Before("gorm:raw").Register("goravel:telemetry:before_raw", r.before("gorm.Raw")); err != nil {
		return err
	}

	return db.Callback().Raw().After("gorm:raw").Register("goravel:telemetry:after_raw", r.after)
}

func (r *GormPlugin) before(spanName string) func(*gorm.DB) {
	return func(tx *gorm.DB) {
		parent := tx.Statement.Context
		spanCtx, _, ok := startSpan(parent, spanName)
		if !ok {
			return
		}

		if r.sqlDB != nil {
			r.poolOnce.Do(func() {
				if err := RegisterPoolMetrics(r.sqlDB, r.driverName, r.poolName); err != nil {
					color.Warningln("failed to register database pool metrics:", err)
				}
			})
		}

		tx.Statement.Context = contextWrapper{Context: spanCtx, parent: parent, start: time.Now()}
	}
}

func (r *GormPlugin) after(tx *gorm.DB) {
	wrapper, ok := tx.Statement.Context.(contextWrapper)
	if !ok {
		return
	}
	tx.Statement.Context = wrapper.parent

	span := oteltrace.SpanFromContext(wrapper.Context)

	endSpan(wrapper.Context, span, wrapper.start, dbSystem(tx.Dialector.Name()), tx.Statement.SQL.String(), tx.Statement.Table, tx.Statement.RowsAffected, tx.Error)
}
