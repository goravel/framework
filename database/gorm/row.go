package gorm

import (
	"github.com/goravel/framework/database/db"
)

type Row struct {
	err   error
	query *Query
	row   map[string]any
}

func (r *Row) Err() error {
	return r.err
}

func (r *Row) Scan(value any) error {
	row := db.NewRow(r.row, r.err)
	if err := row.Scan(value); err != nil {
		return err
	}

	if len(r.query.conditions.eagerLoad) == 0 {
		return nil
	}

	// Per-row eager loading for the Cursor() path. applyEagerLoads consumes its own slice
	// (sets it to nil after running), so we copy into a fresh query so subsequent rows in the
	// cursor still see the queued entries.
	query := r.query.new(r.query.instance)
	query.clearConditions()
	query.conditions.eagerLoad = append([]eagerLoadEntry(nil), r.query.conditions.eagerLoad...)

	return query.applyEagerLoads(value)
}
