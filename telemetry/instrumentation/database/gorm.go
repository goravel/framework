package database

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/support/color"
)

// PluginName is the gorm plugin name this instrumentation registers under.
const PluginName = "goravel:telemetry"

// contextWrapper carries the span context through gorm's Statement.Context for
// the duration of one operation while retaining the original parent, so after()
// can restore it and keep sequential queries as siblings rather than nesting.
type contextWrapper struct {
	context.Context
	parent context.Context
	start  time.Time
}

type callbackRegistrar interface {
	Register(name string, fn func(*gorm.DB)) error
}

type GormPlugin struct {
	instrument *Instrument
	sqlDB      *sql.DB
	poolOnce   sync.Once
}

// NewGormPlugin returns the plugin. It is always registered: telemetry is
// resolved lazily, so the callbacks no-op until it is available and enabled
// rather than deciding at connection-build time when it may not be ready.
func NewGormPlugin(pool contractsdatabase.Pool, connection string) *GormPlugin {
	return &GormPlugin{instrument: NewInstrument(pool, connection)}
}

func (r *GormPlugin) Name() string {
	return PluginName
}

func (r *GormPlugin) Initialize(db *gorm.DB) error {
	// Each row holds three intentionally distinct strings: key is the
	// registration suffix, spanName is the fallback span name (endSpan renames
	// it to "<op> <table>" once known), and "gorm:<op>" is gorm's built-in hook
	// the before/after callbacks anchor to.
	registrations := []struct {
		key      string
		spanName string
		before   callbackRegistrar
		after    callbackRegistrar
	}{
		{"create", "gorm.Create", db.Callback().Create().Before("gorm:create"), db.Callback().Create().After("gorm:create")},
		{"query", "gorm.Query", db.Callback().Query().Before("gorm:query"), db.Callback().Query().After("gorm:query")},
		{"update", "gorm.Update", db.Callback().Update().Before("gorm:update"), db.Callback().Update().After("gorm:update")},
		{"delete", "gorm.Delete", db.Callback().Delete().Before("gorm:delete"), db.Callback().Delete().After("gorm:delete")},
		{"row", "gorm.Row", db.Callback().Row().Before("gorm:row"), db.Callback().Row().After("gorm:row")},
		{"raw", "gorm.Raw", db.Callback().Raw().Before("gorm:raw"), db.Callback().Raw().After("gorm:raw")},
	}

	for _, registration := range registrations {
		if err := registration.before.Register(PluginName+":before_"+registration.key, r.before(registration.spanName)); err != nil {
			return err
		}
		if err := registration.after.Register(PluginName+":after_"+registration.key, r.after); err != nil {
			return err
		}
	}

	// Capture the *sql.DB for pool metrics, which are registered lazily once
	// telemetry is active (see before). A dialector without one has no stats to
	// observe, so skip it while keeping the tracing callbacks.
	if sqlDB, err := db.DB(); err == nil {
		r.sqlDB = sqlDB
	}

	return nil
}

func (r *GormPlugin) before(spanName string) func(*gorm.DB) {
	return func(tx *gorm.DB) {
		if !r.instrument.active() {
			return
		}

		r.poolOnce.Do(r.registerPoolMetrics)

		parent := tx.Statement.Context
		spanCtx, _ := r.instrument.startSpan(parent, spanName)
		tx.Statement.Context = contextWrapper{Context: spanCtx, parent: parent, start: time.Now()}
	}
}

func (r *GormPlugin) after(tx *gorm.DB) {
	wrapper, ok := tx.Statement.Context.(contextWrapper)
	if !ok {
		return
	}
	tx.Statement.Context = wrapper.parent

	span := trace.SpanFromContext(wrapper.Context)
	r.instrument.endSpan(wrapper.Context, span, wrapper.start, tx.Statement.SQL.String(), tx.Statement.Table, tx.Statement.RowsAffected, tx.Error)
}

// registerPoolMetrics is best-effort: a registration failure warns but must not
// disrupt the query that triggered it.
func (r *GormPlugin) registerPoolMetrics() {
	if r.sqlDB == nil {
		return
	}

	if err := r.instrument.registerPoolMetrics(r.sqlDB); err != nil {
		color.Warningln(err.Error())
	}
}
