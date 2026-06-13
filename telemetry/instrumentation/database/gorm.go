package database

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	contractsdatabase "github.com/goravel/framework/contracts/database"
)

type contextWrapper struct {
	context.Context
	parent context.Context
	start  time.Time
}

type gormCallback interface {
	Register(name string, fn func(*gorm.DB)) error
}

type GormPlugin struct {
	instrument *instrument
}

func NewGormPlugin(pool contractsdatabase.Pool, connection string) *GormPlugin {
	inst := newInstrument(pool, connection)
	if inst == nil {
		return nil
	}

	return &GormPlugin{instrument: inst}
}

func (r *GormPlugin) Name() string {
	return "goravel:telemetry"
}

func (r *GormPlugin) Initialize(db *gorm.DB) error {
	registrations := []struct {
		key      string
		spanName string
		before   gormCallback
		after    gormCallback
	}{
		{"create", "gorm.Create", db.Callback().Create().Before("gorm:create"), db.Callback().Create().After("gorm:create")},
		{"query", "gorm.Query", db.Callback().Query().Before("gorm:query"), db.Callback().Query().After("gorm:query")},
		{"update", "gorm.Update", db.Callback().Update().Before("gorm:update"), db.Callback().Update().After("gorm:update")},
		{"delete", "gorm.Delete", db.Callback().Delete().Before("gorm:delete"), db.Callback().Delete().After("gorm:delete")},
		{"row", "gorm.Row", db.Callback().Row().Before("gorm:row"), db.Callback().Row().After("gorm:row")},
		{"raw", "gorm.Raw", db.Callback().Raw().Before("gorm:raw"), db.Callback().Raw().After("gorm:raw")},
	}

	for _, registration := range registrations {
		if err := registration.before.Register("goravel:telemetry:before_"+registration.key, r.before(registration.spanName)); err != nil {
			return err
		}
		if err := registration.after.Register("goravel:telemetry:after_"+registration.key, r.after); err != nil {
			return err
		}
	}

	return nil
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
