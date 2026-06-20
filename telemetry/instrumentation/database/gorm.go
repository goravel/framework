package database

import (
	"database/sql"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/support/color"
)

const PluginName = "goravel:telemetry"

const spanSettingsKey = PluginName + ":span"

type spanState struct {
	span  trace.Span
	start time.Time
}

type GormPlugin struct {
	instrument *Instrument
	sqlDB      *sql.DB
	poolOnce   sync.Once
}

func NewGormPlugin(pool contractsdatabase.Pool, connection string, resolver contractstelemetry.Resolver) *GormPlugin {
	return &GormPlugin{instrument: NewInstrument(pool, connection, resolver)}
}

func (r *GormPlugin) Name() string {
	return PluginName
}

func (r *GormPlugin) Initialize(db *gorm.DB) error {
	cb := db.Callback()

	if err := cb.Create().Before("gorm:create").Register(PluginName+":before_create", r.before); err != nil {
		return err
	}
	if err := cb.Create().After("gorm:create").Register(PluginName+":after_create", r.after); err != nil {
		return err
	}
	if err := cb.Query().Before("gorm:query").Register(PluginName+":before_query", r.before); err != nil {
		return err
	}
	if err := cb.Query().After("gorm:query").Register(PluginName+":after_query", r.after); err != nil {
		return err
	}
	if err := cb.Update().Before("gorm:update").Register(PluginName+":before_update", r.before); err != nil {
		return err
	}
	if err := cb.Update().After("gorm:update").Register(PluginName+":after_update", r.after); err != nil {
		return err
	}
	if err := cb.Delete().Before("gorm:delete").Register(PluginName+":before_delete", r.before); err != nil {
		return err
	}
	if err := cb.Delete().After("gorm:delete").Register(PluginName+":after_delete", r.after); err != nil {
		return err
	}
	if err := cb.Row().Before("gorm:row").Register(PluginName+":before_row", r.before); err != nil {
		return err
	}
	if err := cb.Row().After("gorm:row").Register(PluginName+":after_row", r.after); err != nil {
		return err
	}
	if err := cb.Raw().Before("gorm:raw").Register(PluginName+":before_raw", r.before); err != nil {
		return err
	}
	if err := cb.Raw().After("gorm:raw").Register(PluginName+":after_raw", r.after); err != nil {
		return err
	}

	if sqlDB, err := db.DB(); err == nil {
		r.sqlDB = sqlDB
	}

	return nil
}

func (r *GormPlugin) before(tx *gorm.DB) {
	if !r.instrument.active() {
		return
	}

	r.poolOnce.Do(r.registerPoolMetrics)

	ctx, span := r.instrument.startSpan(tx.Statement.Context, "db")
	tx.Statement.Context = ctx
	tx.Statement.Settings.Store(spanSettingsKey, spanState{span: span, start: time.Now()})
}

func (r *GormPlugin) after(tx *gorm.DB) {
	val, ok := tx.Statement.Settings.Load(spanSettingsKey)
	if !ok {
		return
	}

	state := val.(spanState)
	r.instrument.endSpan(tx.Statement.Context, state.span, state.start, tx.Statement.SQL.String(), tx.Statement.Table, tx.Statement.RowsAffected, tx.Error)
}

func (r *GormPlugin) registerPoolMetrics() {
	if r.sqlDB == nil {
		return
	}

	if err := r.instrument.registerPoolMetrics(r.sqlDB); err != nil {
		color.Warningln(err.Error())
	}
}
