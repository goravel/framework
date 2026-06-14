package database

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	contractsdatabase "github.com/goravel/framework/contracts/database"
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
}

func NewGormPlugin(pool contractsdatabase.Pool, connection string) *GormPlugin {
	inst := NewInstrument(pool, connection)
	if inst == nil {
		return nil
	}

	return &GormPlugin{instrument: inst}
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

	// Pool metrics are best-effort: a dialector without a *sql.DB ConnPool has
	// no stats to observe, so skip them while keeping the tracing callbacks.
	sqlDB, err := db.DB()
	if err != nil {
		return nil
	}

	return r.instrument.registerPoolMetrics(sqlDB)
}

func (r *GormPlugin) before(spanName string) func(*gorm.DB) {
	return func(tx *gorm.DB) {
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
