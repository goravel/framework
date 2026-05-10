package gorm

import (
	"fmt"

	gormio "gorm.io/gorm"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

// pivotQuery is the gorm-backed implementation of contractsorm.PivotQuery handed to user-supplied
// PivotCallback closures. It wraps a *gormio.DB chain so each Where* call appends a clause; the
// final *gormio.DB is read back via .db() and chained onto whatever query the caller is about to
// execute (SELECT for existingPivotIDs/allPivotIDs, DELETE for DetachRelation, UPDATE for
// UpdateExistingPivotRelation).
//
// Identifier qualification: the pivot table is always present in the surrounding query, so the
// callback is expected to pass bare column names (e.g. "active") — we prefix them with the pivot
// table name to avoid ambiguity when the surrounding query JOINs the related table.
type pivotQuery struct {
	db        *gormio.DB
	tableName string
}

func newPivotQuery(db *gormio.DB, tableName string) *pivotQuery {
	return &pivotQuery{db: db, tableName: tableName}
}

// qualified returns "<pivotTable>.<column>" with the project's quoteIdent treatment, mirroring
// how relation_writes.go formats its hand-written WHERE clauses.
func (p *pivotQuery) qualified(column string) string {
	return fmt.Sprintf("%s.%s", quoteIdent(p.tableName), quoteIdent(column))
}

func (p *pivotQuery) Where(column string, args ...any) contractsorm.PivotQuery {
	switch len(args) {
	case 1:
		// (column, value) — operator defaults to "=".
		p.db = p.db.Where(fmt.Sprintf("%s = ?", p.qualified(column)), args[0])
	case 2:
		// (column, operator, value).
		op, ok := args[0].(string)
		if !ok {
			// Defensive fallback: treat both args as values for an "=" + AND join. This shouldn't
			// happen with normal usage but avoids a silent panic on bad input.
			p.db = p.db.Where(fmt.Sprintf("%s = ?", p.qualified(column)), args[0]).
				Where(fmt.Sprintf("%s = ?", p.qualified(column)), args[1])
			return p
		}
		p.db = p.db.Where(fmt.Sprintf("%s %s ?", p.qualified(column), op), args[1])
	default:
		// 0 args or >2 args — no-op. The narrower interface should prevent this in practice.
	}
	return p
}

func (p *pivotQuery) WhereIn(column string, values []any) contractsorm.PivotQuery {
	p.db = p.db.Where(fmt.Sprintf("%s IN ?", p.qualified(column)), values)
	return p
}

func (p *pivotQuery) WhereNotIn(column string, values []any) contractsorm.PivotQuery {
	p.db = p.db.Where(fmt.Sprintf("%s NOT IN ?", p.qualified(column)), values)
	return p
}

func (p *pivotQuery) WhereNull(column string) contractsorm.PivotQuery {
	p.db = p.db.Where(fmt.Sprintf("%s IS NULL", p.qualified(column)))
	return p
}

func (p *pivotQuery) WhereNotNull(column string) contractsorm.PivotQuery {
	p.db = p.db.Where(fmt.Sprintf("%s IS NOT NULL", p.qualified(column)))
	return p
}

// applyOnPivotQuery threads the descriptor's OnPivotQuery callback through q. No-op if the
// callback is nil. Used by every pivot-table read/update/delete code path so the per-relation
// scope is honoured uniformly. INSERT paths (AttachRelation / AttachWithPivotRelation) skip this
// — see the doc on contractsorm.PivotQuery for rationale.
func applyOnPivotQuery(q *gormio.DB, desc *relationDescriptor) *gormio.DB {
	if desc.onPivotQuery == nil {
		return q
	}
	pq := newPivotQuery(q, desc.pivotTable)
	desc.onPivotQuery(pq)
	return pq.db
}
