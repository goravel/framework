// Package gorm contains the GORM-backed implementation of the framework's database/orm contracts.
//
// Where the upstream framework has first-class Relation objects with `getRelationExistenceQuery` /
// `getRelationExistenceCountQuery` methods, GORM models its relationships through struct-tag
// metadata. The two are bridged here by relation.go's resolver: it inspects gorm.Schema and a
// model's optional ModelWithThroughRelations declaration, then produces a relationDescriptor
// that knows how to emit a correlated EXISTS / count subquery for any of HasOne, HasMany,
// BelongsTo, BelongsToMany, MorphOne, MorphMany, HasOneThrough or HasManyThrough.
package gorm

import (
	"fmt"
	"strings"

	gormio "gorm.io/gorm"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/deep"
	"github.com/goravel/framework/support/str"
)

// ---------------------------------------------------------------------------
// Public API: existence (has / whereHas / doesntHave / orWhereDoesntHave / ...)
//
// Each method appends a relationExistence to conditions.relations; the actual subquery is built
// when buildConditions runs (so the parent model from .Model() / dest can be resolved).
// ---------------------------------------------------------------------------

// Has adds a relationship count / exists condition to the query.
//
// args may include any combination of a RelationCallback (or func(Query) Query) for scoping the
// inner subquery, a string operator (defaults to ">="), and an int count (defaults to 1). For
// nested relations dot-notation is honoured: `Has("Books.Author")` is shorthand for
// `WhereHas("Books", q -> q.Has("Author"))`.
func (r *Query) Has(relation string, args ...any) contractsorm.Query {
	return r.queueRelationExistence(relation, args, "and", false)
}

// OrHas adds a relationship count / exists condition to the query with an "or" conjunction.
func (r *Query) OrHas(relation string, args ...any) contractsorm.Query {
	return r.queueRelationExistence(relation, args, "or", false)
}

// DoesntHave adds a relationship absence condition to the query.
// Equivalent to Has(rel, "<", 1).
func (r *Query) DoesntHave(relation string, args ...any) contractsorm.Query {
	return r.queueRelationExistence(relation, args, "and", true)
}

// OrDoesntHave adds a relationship absence condition with an "or" conjunction.
func (r *Query) OrDoesntHave(relation string, args ...any) contractsorm.Query {
	return r.queueRelationExistence(relation, args, "or", true)
}

// WhereHas adds a relationship count / exists condition to the query with where clauses.
// Functionally identical to Has - the rename merely makes call sites that always pass a callback
// read more naturally.
func (r *Query) WhereHas(relation string, args ...any) contractsorm.Query {
	return r.queueRelationExistence(relation, args, "and", false)
}

// OrWhereHas adds a relationship count / exists condition with where clauses and an "or"
// conjunction.
func (r *Query) OrWhereHas(relation string, args ...any) contractsorm.Query {
	return r.queueRelationExistence(relation, args, "or", false)
}

// WhereDoesntHave adds a relationship absence condition to the query with where clauses.
func (r *Query) WhereDoesntHave(relation string, args ...any) contractsorm.Query {
	return r.queueRelationExistence(relation, args, "and", true)
}

// OrWhereDoesntHave adds a relationship absence condition with where clauses and an "or"
// conjunction.
func (r *Query) OrWhereDoesntHave(relation string, args ...any) contractsorm.Query {
	return r.queueRelationExistence(relation, args, "or", true)
}

// HasMorph adds a polymorphic relationship count / exists condition to the query.
// types is a slice of model instances; the morph value used in the polymorphic type column is
// derived from each model's GORM-resolved table name.
//
// Note: auto-discovery of distinct morph values via `types = ['*']` is not supported; an explicit
// list of model instances is required.
func (r *Query) HasMorph(relation string, types []any, args ...any) contractsorm.Query {
	return r.queueMorphExistence(relation, types, args, "and", false)
}

// OrHasMorph adds a polymorphic relationship count / exists condition with an "or" conjunction.
func (r *Query) OrHasMorph(relation string, types []any, args ...any) contractsorm.Query {
	return r.queueMorphExistence(relation, types, args, "or", false)
}

// DoesntHaveMorph adds a polymorphic relationship absence condition.
func (r *Query) DoesntHaveMorph(relation string, types []any, args ...any) contractsorm.Query {
	return r.queueMorphExistence(relation, types, args, "and", true)
}

// OrDoesntHaveMorph adds a polymorphic relationship absence condition with an "or" conjunction.
func (r *Query) OrDoesntHaveMorph(relation string, types []any, args ...any) contractsorm.Query {
	return r.queueMorphExistence(relation, types, args, "or", true)
}

// WhereHasMorph adds a polymorphic existence condition with where clauses; callbacks may be
// MorphRelationCallback for per-type scoping.
func (r *Query) WhereHasMorph(relation string, types []any, args ...any) contractsorm.Query {
	return r.queueMorphExistence(relation, types, args, "and", false)
}

// OrWhereHasMorph adds a polymorphic existence condition with where clauses and an "or"
// conjunction.
func (r *Query) OrWhereHasMorph(relation string, types []any, args ...any) contractsorm.Query {
	return r.queueMorphExistence(relation, types, args, "or", false)
}

// WhereDoesntHaveMorph adds a polymorphic absence condition with where clauses.
func (r *Query) WhereDoesntHaveMorph(relation string, types []any, args ...any) contractsorm.Query {
	return r.queueMorphExistence(relation, types, args, "and", true)
}

// OrWhereDoesntHaveMorph adds a polymorphic absence condition with where clauses and an "or"
// conjunction.
func (r *Query) OrWhereDoesntHaveMorph(relation string, types []any, args ...any) contractsorm.Query {
	return r.queueMorphExistence(relation, types, args, "or", true)
}

// ---------------------------------------------------------------------------
// Public API: aggregate sub-selects (withCount / withMax / withSum / ...)
//
// Each method appends a selectSub to conditions.selectSubs; the actual sub-select column is
// emitted when buildConditions runs (so the parent model can be resolved and the alias derived).
// ---------------------------------------------------------------------------

// WithAggregate adds a sub-select to include an aggregate value for a relationship.
// fn must be one of: count, max, min, sum, avg, exists.
func (r *Query) WithAggregate(relation, column, fn string, args ...any) contractsorm.Query {
	if !validAggregateFn(fn) {
		query := r.new(r.instance.Session(&gormio.Session{}))
		_ = query.instance.AddError(errors.OrmRelationInvalidAggregate.Args(fn))
		return query
	}
	cb, _, _, err := parseRelationArgs(args)
	if err != nil {
		query := r.new(r.instance.Session(&gormio.Session{}))
		_ = query.instance.AddError(err)
		return query
	}
	conditions := r.conditions
	conditions.selectSubs = deep.Append(conditions.selectSubs, selectSub{
		relation: relation,
		column:   column,
		function: fn,
		alias:    aggregateAlias(relation, fn, column),
		callback: cb,
	})
	return r.setConditions(conditions)
}

// WithCount adds sub-select queries to count the relations. Each entry may be either:
//   - a string ("Books") - emits `(SELECT COUNT(*) FROM ...) AS books_count`
//   - a contractsorm.RelationCount struct for scoped counts and/or custom alias
func (r *Query) WithCount(relations ...any) contractsorm.Query {
	current := r
	for _, raw := range relations {
		switch v := raw.(type) {
		case string:
			next, ok := current.WithAggregate(v, "*", "count").(*Query)
			if !ok {
				return current
			}
			current = next
		case contractsorm.RelationCount:
			args := []any{}
			if v.Callback != nil {
				args = append(args, v.Callback)
			}
			next, ok := current.WithAggregate(v.Name, "*", "count", args...).(*Query)
			if !ok {
				return current
			}
			if v.Alias != "" {
				if n := len(next.conditions.selectSubs); n > 0 {
					next.conditions.selectSubs[n-1].alias = v.Alias
				}
			}
			current = next
		default:
			impl := r.new(r.instance.Session(&gormio.Session{}))
			_ = impl.instance.AddError(errors.OrmRelationInvalidArgument.Args(raw))
			return impl
		}
	}
	return current
}

// WithMax adds sub-select queries to include the max of the relation's column.
func (r *Query) WithMax(relation, column string, args ...any) contractsorm.Query {
	return r.WithAggregate(relation, column, "max", args...)
}

// WithMin adds sub-select queries to include the min of the relation's column.
func (r *Query) WithMin(relation, column string, args ...any) contractsorm.Query {
	return r.WithAggregate(relation, column, "min", args...)
}

// WithSum adds sub-select queries to include the sum of the relation's column.
func (r *Query) WithSum(relation, column string, args ...any) contractsorm.Query {
	return r.WithAggregate(relation, column, "sum", args...)
}

// WithAvg adds sub-select queries to include the average of the relation's column.
func (r *Query) WithAvg(relation, column string, args ...any) contractsorm.Query {
	return r.WithAggregate(relation, column, "avg", args...)
}

// WithExists adds sub-select queries to include the existence of related models. The result is
// emitted as `CASE WHEN EXISTS (...) THEN 1 ELSE 0 END` for cross-dialect portability (SQL Server
// has no boolean literal). The dest field may be either `bool` or an integer type - Go's
// database/sql layer converts 0/1 ints to bool automatically.
func (r *Query) WithExists(relations ...string) contractsorm.Query {
	current := r
	for _, rel := range relations {
		next, ok := current.WithAggregate(rel, "*", "exists").(*Query)
		if !ok {
			return current
		}
		current = next
	}
	return current
}

// ---------------------------------------------------------------------------
// Public API: eager loading (With / Without / WithOnly)
//
// These methods build up conditions.eagerLoad. The actual loader runs after the main query
// returns; see eager_loader.go for the execution side.
// ---------------------------------------------------------------------------

// With eagerly loads the given relationships using Goravel's own loader. Accepts the
// union of fedaco's with(...) shapes; see the orm.Query interface comment for the full grammar.
func (r *Query) With(args ...any) contractsorm.Query {
	entries, err := parseEagerLoad(args)
	if err != nil {
		query := r.new(r.instance.Session(&gormio.Session{}))
		_ = query.instance.AddError(err)
		return query
	}
	conditions := r.conditions
	for _, e := range entries {
		conditions.eagerLoad = upsertEagerLoadEntry(conditions.eagerLoad, e, e.callback != nil || e.columns != nil)
	}
	return r.setConditions(conditions)
}

// Without removes the named relations from the eager-load list. Mirrors fedaco's
// without(). Names must match exactly (including dot-paths, e.g. "Books.Author").
func (r *Query) Without(relations ...string) contractsorm.Query {
	if len(relations) == 0 || len(r.conditions.eagerLoad) == 0 {
		return r
	}
	drop := make(map[string]struct{}, len(relations))
	for _, n := range relations {
		drop[n] = struct{}{}
	}
	conditions := r.conditions
	filtered := make([]eagerLoadEntry, 0, len(conditions.eagerLoad))
	for _, e := range conditions.eagerLoad {
		if _, omit := drop[e.relation]; omit {
			continue
		}
		filtered = append(filtered, e)
	}
	conditions.eagerLoad = filtered
	return r.setConditions(conditions)
}

// WithOnly clears the eager-load list, then adds the given relations. Mirrors fedaco's
// withOnly(). Useful when a default-scoped query has eager loads you want to override.
func (r *Query) WithOnly(args ...any) contractsorm.Query {
	conditions := r.conditions
	conditions.eagerLoad = nil
	return r.setConditions(conditions).With(args...)
}

// ---------------------------------------------------------------------------
// Internal: queueing
// ---------------------------------------------------------------------------

func (r *Query) queueRelationExistence(relation string, args []any, conjunction string, doesntHave bool) contractsorm.Query {
	cb, op, count, err := parseRelationArgs(args)
	if err != nil {
		query := r.new(r.instance.Session(&gormio.Session{}))
		_ = query.instance.AddError(err)
		return query
	}
	if doesntHave {
		op = "<"
		count = 1
	}
	conditions := r.conditions
	conditions.relations = deep.Append(conditions.relations, relationExistence{
		relation:    relation,
		operator:    op,
		count:       count,
		conjunction: conjunction,
		callback:    cb,
	})
	return r.setConditions(conditions)
}

func (r *Query) queueMorphExistence(relation string, types []any, args []any, conjunction string, doesntHave bool) contractsorm.Query {
	if len(types) == 0 {
		query := r.new(r.instance.Session(&gormio.Session{}))
		_ = query.instance.AddError(errors.OrmRelationMorphTypesEmpty)
		return query
	}
	cb, mcb, op, count, err := parseMorphRelationArgs(args)
	if err != nil {
		query := r.new(r.instance.Session(&gormio.Session{}))
		_ = query.instance.AddError(err)
		return query
	}
	if doesntHave {
		op = "<"
		count = 1
	}
	conditions := r.conditions
	conditions.relations = deep.Append(conditions.relations, relationExistence{
		relation:      relation,
		operator:      op,
		count:         count,
		conjunction:   conjunction,
		callback:      cb,
		morphTypes:    types,
		morphCallback: mcb,
	})
	return r.setConditions(conditions)
}

// ---------------------------------------------------------------------------
// Internal: build phase (called from buildConditions)
// ---------------------------------------------------------------------------

// buildRelations compiles all queued relation existence conditions into the outer GORM query.
// Resolves each relation against the parent model now that one of conditions.model / conditions.dest
// is set.
func (r *Query) buildRelations(db *gormio.DB) *gormio.DB {
	if len(r.conditions.relations) == 0 {
		return db
	}
	parent := r.parentModel()
	if parent == nil {
		_ = db.AddError(errors.OrmQueryEmptyRelation)
		return db
	}

	for _, item := range r.conditions.relations {
		if len(item.morphTypes) > 0 {
			db = r.applyMorphExistence(db, parent, item)
		} else {
			db = r.applyExistence(db, parent, item)
		}
	}
	r.conditions.relations = nil
	return db
}

// buildSelectSubAggregates compiles WithCount / WithSum / WithExists / etc. as sub-select
// columns. GORM's Select() overwrites prior selections, so we coalesce the parent's existing
// columns (or a default `<parent>.*`) with all sub-select expressions and emit a single
// Select() containing the full list of column expressions plus the inner subqueries as bindings.
func (r *Query) buildSelectSubAggregates(db *gormio.DB) *gormio.DB {
	if len(r.conditions.selectSubs) == 0 {
		return db
	}
	parent := r.parentModel()
	if parent == nil {
		_ = db.AddError(errors.OrmQueryEmptyRelation)
		return db
	}
	parentTable := r.parentTable(parent)

	// Start with the columns the user already requested (via prior .Select() calls). If they
	// didn't request anything, we project the parent's wildcard so the row can still be scanned
	// into the dest model.
	existing := append([]string{}, db.Statement.Selects...)
	if len(existing) == 0 {
		existing = []string{fmt.Sprintf("%s.*", quoteIdent(parentTable))}
	}

	var subExprs []string
	var subArgs []any
	for _, sub := range r.conditions.selectSubs {
		desc, err := resolveRelation(r.instance, parent, sub.relation)
		if err != nil {
			_ = db.AddError(err)
			continue
		}
		inner := r.compileAggregateSubquery(desc, sub)
		if inner == nil {
			continue
		}
		alias := sub.alias
		if alias == "" {
			alias = aggregateAlias(sub.relation, sub.function, sub.column)
		}
		if sub.function == "exists" {
			// Use CASE WHEN EXISTS instead of bare `EXISTS (...) AS col`: PostgreSQL returns
			// EXISTS as a native bool which won't scan into integer struct fields, and SQL Server
			// rejects EXISTS as a column expression entirely. CASE WHEN yields a portable 0/1 int
			// across SQLite / MySQL / PostgreSQL / SQL Server.
			subExprs = append(subExprs, fmt.Sprintf("CASE WHEN EXISTS (?) THEN 1 ELSE 0 END AS %s", quoteIdent(alias)))
		} else {
			subExprs = append(subExprs, fmt.Sprintf("(?) AS %s", quoteIdent(alias)))
		}
		subArgs = append(subArgs, inner)
	}
	if len(subExprs) == 0 {
		r.conditions.selectSubs = nil
		return db
	}
	full := strings.Join(append(existing, subExprs...), ", ")
	db = db.Select(full, subArgs...)
	r.conditions.selectSubs = nil
	return db
}

// ---------------------------------------------------------------------------
// Internal: existence subquery construction
// ---------------------------------------------------------------------------

func (r *Query) applyExistence(db *gormio.DB, parent any, item relationExistence) *gormio.DB {
	desc, err := resolveRelation(r.instance, parent, item.relation)
	if err != nil {
		_ = db.AddError(err)
		return db
	}
	inner := r.compileExistenceSubquery(desc, item.callback)
	if inner == nil {
		return db
	}
	return r.attachHasWhere(db, inner, item.operator, item.count, item.conjunction)
}

func (r *Query) applyMorphExistence(db *gormio.DB, parent any, item relationExistence) *gormio.DB {
	desc, err := resolveRelation(r.instance, parent, item.relation)
	if err != nil {
		_ = db.AddError(err)
		return db
	}
	switch desc.kind {
	case relKindMorphOne, relKindMorphMany:
		return r.applyOutboundMorphExistence(db, desc, item)
	case relKindMorphTo:
		return r.applyMorphToExistence(db, desc, item)
	default:
		_ = db.AddError(errors.OrmRelationUnsupported.Args(item.relation, fmt.Sprintf("%T", parent), "morph (must be polymorphic relation)"))
		return db
	}
}

// applyOutboundMorphExistence builds the morph-existence clauses for outbound MorphOne /
// MorphMany. The morph_type column lives on the *related* table (e.g. houses.houseable_type), so
// for each requested type we build a correlated EXISTS subquery whose inner WHERE pins
// houses.houseable_type to that type's morph value. Multiple types are joined with OR.
func (r *Query) applyOutboundMorphExistence(db *gormio.DB, desc *relationDescriptor, item relationExistence) *gormio.DB {
	sub := r.freshSession()
	first := true
	for _, typeModel := range item.morphTypes {
		morphValue, terr := tableNameFor(r.instance, typeModel)
		if terr != nil {
			_ = db.AddError(terr)
			continue
		}
		morphValue = resolveMorphValue(typeModel, morphValue)

		var perTypeCallback contractsorm.RelationCallback
		if item.morphCallback != nil {
			cb := item.morphCallback
			captured := morphValue
			perTypeCallback = func(q contractsorm.Query) contractsorm.Query {
				return cb(q, captured)
			}
		} else {
			perTypeCallback = item.callback
		}
		inner := r.compileMorphExistenceSubquery(desc, morphValue, perTypeCallback)
		if inner == nil {
			continue
		}

		var clauseSQL string
		var clauseArgs []any
		if shouldUseExists(item.operator, item.count) {
			negate := item.operator == "<" && item.count == 1
			if negate {
				clauseSQL = "NOT EXISTS (?)"
			} else {
				clauseSQL = "EXISTS (?)"
			}
			clauseArgs = []any{inner}
		} else {
			countInner := inner.Select("COUNT(*)")
			clauseSQL = fmt.Sprintf("(?) %s ?", item.operator)
			clauseArgs = []any{countInner, item.count}
		}

		if first {
			sub = sub.Where(clauseSQL, clauseArgs...)
			first = false
		} else {
			sub = sub.Or(clauseSQL, clauseArgs...)
		}
	}
	if first {
		return db
	}

	if item.conjunction == "or" {
		return db.Or(sub)
	}
	return db.Where(sub)
}

// applyMorphToExistence builds the inverse-polymorphic existence clauses. Mirrors fedaco's
// hasMorph at libs/fedaco/src/fedaco/mixins/queries-relationships.ts:320-378: for each requested
// type, we synthesise a BelongsTo-shaped subquery against that type's table, and AND it with a
// type filter on the parent's morph_type column. The per-type clauses are OR-ed together.
//
// Generated SQL pattern:
//
//	WHERE (
//	     (parents.imageable_type = 'post'  AND ((SELECT count(*) FROM posts  WHERE posts.id  = parents.imageable_id) >= N))
//	  OR (parents.imageable_type = 'video' AND ((SELECT count(*) FROM videos WHERE videos.id = parents.imageable_id) >= N))
//	)
func (r *Query) applyMorphToExistence(db *gormio.DB, desc *relationDescriptor, item relationExistence) *gormio.DB {
	sub := r.freshSession()
	first := true
	ownerKey := desc.morphOwnerKey
	if ownerKey == "" {
		ownerKey = "id"
	}

	for _, typeModel := range item.morphTypes {
		relatedTable, terr := tableNameFor(r.instance, typeModel)
		if terr != nil {
			_ = db.AddError(terr)
			continue
		}
		morphValue := resolveMorphValue(typeModel, relatedTable)

		// Per-type callback resolution (same shape as outbound morph existence).
		var perTypeCallback contractsorm.RelationCallback
		if item.morphCallback != nil {
			cb := item.morphCallback
			captured := morphValue
			perTypeCallback = func(q contractsorm.Query) contractsorm.Query {
				return cb(q, captured)
			}
		} else {
			perTypeCallback = item.callback
		}

		// Build the BelongsTo-shaped inner: SELECT * FROM <relatedTable> WHERE
		// <relatedTable>.<ownerKey> = <parentTable>.<morphIDColumn>.
		inner := r.freshSession().Table(relatedTable).Where(fmt.Sprintf("%s.%s = %s.%s",
			quoteIdent(relatedTable), quoteIdent(ownerKey),
			quoteIdent(desc.parentTable), quoteIdent(desc.morphIDColumn)))
		if desc.onQuery != nil {
			wrapper := r.wrap(inner)
			result := desc.onQuery(wrapper)
			if w, ok := result.(*Query); ok {
				inner = w.buildConditions().instance
			}
		}
		if perTypeCallback != nil {
			wrapper := r.wrap(inner)
			result := perTypeCallback(wrapper)
			if w, ok := result.(*Query); ok {
				inner = w.buildConditions().instance
			}
		}

		// Type filter applied at the outer level (parent table).
		typeClause := fmt.Sprintf("%s.%s = ?", quoteIdent(desc.parentTable), quoteIdent(desc.morphTypeColumn))
		typeArgs := []any{morphValue}

		var perType *gormio.DB
		if shouldUseExists(item.operator, item.count) {
			negate := item.operator == "<" && item.count == 1
			if negate {
				perType = r.freshSession().Where(typeClause, typeArgs...).Where("NOT EXISTS (?)", inner)
			} else {
				perType = r.freshSession().Where(typeClause, typeArgs...).Where("EXISTS (?)", inner)
			}
		} else {
			countInner := inner.Select("COUNT(*)")
			perType = r.freshSession().Where(typeClause, typeArgs...).Where(fmt.Sprintf("(?) %s ?", item.operator), countInner, item.count)
		}

		if first {
			sub = sub.Where(perType)
			first = false
		} else {
			sub = sub.Or(perType)
		}
	}
	if first {
		return db
	}

	if item.conjunction == "or" {
		return db.Or(sub)
	}
	return db.Where(sub)
}

// compileExistenceSubquery returns a *gormio.DB representing the inner SELECT correlated to the
// parent table. The returned DB is intended to be passed as a value bound to a "?" placeholder
// in the outer query (GORM will inline it as a subquery and merge bindings).
func (r *Query) compileExistenceSubquery(desc *relationDescriptor, callback contractsorm.RelationCallback) *gormio.DB {
	inner := r.freshSession().Table(desc.relatedTable)

	switch desc.kind {
	case relKindHasOne, relKindHasMany:
		for _, ref := range desc.references {
			inner = inner.Where(fmt.Sprintf("%s.%s = %s.%s",
				quoteIdent(ref.foreignTable), quoteIdent(ref.foreignColumn),
				quoteIdent(ref.primaryTable), quoteIdent(ref.primaryColumn)))
		}
	case relKindBelongsTo:
		for _, ref := range desc.references {
			inner = inner.Where(fmt.Sprintf("%s.%s = %s.%s",
				quoteIdent(ref.primaryTable), quoteIdent(ref.primaryColumn),
				quoteIdent(ref.foreignTable), quoteIdent(ref.foreignColumn)))
		}
	case relKindMany2Many:
		inner = inner.Joins(fmt.Sprintf("INNER JOIN %s ON %s.%s = %s.%s",
			quoteIdent(desc.pivotTable),
			quoteIdent(desc.pivotTable), quoteIdent(desc.pivotRelatedRef.foreignColumn),
			quoteIdent(desc.relatedTable), quoteIdent(desc.pivotRelatedRef.primaryColumn)))
		inner = inner.Where(fmt.Sprintf("%s.%s = %s.%s",
			quoteIdent(desc.pivotTable), quoteIdent(desc.pivotParentRef.foreignColumn),
			quoteIdent(desc.pivotParentRef.primaryTable), quoteIdent(desc.pivotParentRef.primaryColumn)))
	case relKindMorphToMany:
		inner = inner.Joins(fmt.Sprintf("INNER JOIN %s ON %s.%s = %s.%s",
			quoteIdent(desc.pivotTable),
			quoteIdent(desc.pivotTable), quoteIdent(desc.pivotRelatedRef.foreignColumn),
			quoteIdent(desc.relatedTable), quoteIdent(desc.pivotRelatedRef.primaryColumn)))
		inner = inner.Where(fmt.Sprintf("%s.%s = %s.%s",
			quoteIdent(desc.pivotTable), quoteIdent(desc.pivotParentRef.foreignColumn),
			quoteIdent(desc.pivotParentRef.primaryTable), quoteIdent(desc.pivotParentRef.primaryColumn)))
		inner = inner.Where(fmt.Sprintf("%s.%s = ?",
			quoteIdent(desc.pivotTable), quoteIdent(desc.morphTypeColumn)), desc.morphValue)
	case relKindMorphOne, relKindMorphMany:
		for _, ref := range desc.references {
			inner = inner.Where(fmt.Sprintf("%s.%s = %s.%s",
				quoteIdent(ref.foreignTable), quoteIdent(ref.foreignColumn),
				quoteIdent(ref.primaryTable), quoteIdent(ref.primaryColumn)))
		}
		inner = inner.Where(fmt.Sprintf("%s.%s = ?",
			quoteIdent(desc.relatedTable), quoteIdent(desc.morphTypeColumn)), desc.morphValue)
	case relKindHasOneThrough, relKindHasManyThrough:
		inner = inner.Joins(fmt.Sprintf("INNER JOIN %s ON %s.%s = %s.%s",
			quoteIdent(desc.throughTable),
			quoteIdent(desc.relatedTable), quoteIdent(desc.secondKey),
			quoteIdent(desc.throughTable), quoteIdent(desc.secondLocalKey)))
		inner = inner.Where(fmt.Sprintf("%s.%s = %s.%s",
			quoteIdent(desc.throughTable), quoteIdent(desc.firstKey),
			quoteIdent(desc.parentTable), quoteIdent(desc.localKey)))
	default:
		_ = inner.AddError(errors.OrmRelationUnsupported.Args(desc.name, desc.parentTable, fmt.Sprintf("kind=%d", desc.kind)))
		return inner
	}

	// Apply the relation's default scope before the caller's callback so the user's WhereHas
	// callback can layer on top of the always-on filter.
	if desc.onQuery != nil {
		wrapper := r.wrap(inner)
		wrapped := desc.onQuery(wrapper)
		if w, ok := wrapped.(*Query); ok {
			inner = w.buildConditions().instance
		}
	}

	if callback != nil {
		wrapper := r.wrap(inner)
		wrapped := callback(wrapper)
		if w, ok := wrapped.(*Query); ok {
			inner = w.buildConditions().instance
		}
	}

	if desc.nested != nil {
		nested := r.compileExistenceSubquery(desc.nested, nil)
		if nested != nil {
			inner = inner.Where("EXISTS (?)", nested)
		}
	}

	return inner.Select("1")
}

// compileMorphExistenceSubquery is the morph-specific variant. The morph_type column is included
// in the *outer* group, not the inner subquery, because it lives on the parent's side.
func (r *Query) compileMorphExistenceSubquery(desc *relationDescriptor, morphValue string, callback contractsorm.RelationCallback) *gormio.DB {
	inner := r.freshSession().Table(desc.relatedTable)
	for _, ref := range desc.references {
		inner = inner.Where(fmt.Sprintf("%s.%s = %s.%s",
			quoteIdent(ref.foreignTable), quoteIdent(ref.foreignColumn),
			quoteIdent(ref.primaryTable), quoteIdent(ref.primaryColumn)))
	}
	inner = inner.Where(fmt.Sprintf("%s.%s = ?",
		quoteIdent(desc.relatedTable), quoteIdent(desc.morphTypeColumn)), morphValue)
	if desc.onQuery != nil {
		wrapper := r.wrap(inner)
		wrapped := desc.onQuery(wrapper)
		if w, ok := wrapped.(*Query); ok {
			inner = w.buildConditions().instance
		}
	}
	if callback != nil {
		wrapper := r.wrap(inner)
		wrapped := callback(wrapper)
		if w, ok := wrapped.(*Query); ok {
			inner = w.buildConditions().instance
		}
	}
	return inner.Select("1")
}

// compileAggregateSubquery returns a *gormio.DB whose compiled SQL is the inner SELECT for an
// aggregate. The select expression is dictated by sub.function.
func (r *Query) compileAggregateSubquery(desc *relationDescriptor, sub selectSub) *gormio.DB {
	inner := r.compileExistenceSubquery(desc, sub.callback)
	if inner == nil {
		return nil
	}
	var selectExpr string
	switch sub.function {
	case "count":
		selectExpr = "COUNT(*)"
	case "exists":
		selectExpr = "1"
	default:
		col := "*"
		if sub.column != "*" && sub.column != "" {
			col = fmt.Sprintf("%s.%s", quoteIdent(desc.relatedTable), quoteIdent(sub.column))
		}
		selectExpr = fmt.Sprintf("%s(%s)", strings.ToUpper(sub.function), col)
	}
	return inner.Select(selectExpr)
}

// attachHasWhere appends the inner subquery to the outer query as either WHERE [NOT] EXISTS
// (preferred, when operator/count match the EXISTS optimisation) or as a (SELECT count(*)) op ?
// comparison.
func (r *Query) attachHasWhere(db *gormio.DB, inner *gormio.DB, operator string, count int, conjunction string) *gormio.DB {
	if shouldUseExists(operator, count) {
		negate := operator == "<" && count == 1
		var clause string
		if negate {
			clause = "NOT EXISTS (?)"
		} else {
			clause = "EXISTS (?)"
		}
		if conjunction == "or" {
			return db.Or(clause, inner)
		}
		return db.Where(clause, inner)
	}

	countInner := inner.Select("COUNT(*)")
	clause := fmt.Sprintf("(?) %s ?", operator)
	if conjunction == "or" {
		return db.Or(clause, countInner, count)
	}
	return db.Where(clause, countInner, count)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *Query) freshSession() *gormio.DB {
	return r.instance.Session(&gormio.Session{NewDB: true, Initialized: true})
}

// wrap turns a fresh GORM DB into a Goravel Query wrapper so user callbacks can use the
// familiar Where/OrWhere/etc. surface.
func (r *Query) wrap(db *gormio.DB) *Query {
	return NewQuery(r.ctx, r.config, r.dbConfig, db, r.grammar, r.log, r.modelToObserver, nil)
}

func (r *Query) parentModel() any {
	if r.conditions.model != nil {
		return r.conditions.model
	}
	if r.conditions.dest != nil {
		if m, err := modelToStruct(r.conditions.dest); err == nil && m != nil {
			return m
		}
	}
	if r.instance != nil && r.instance.Statement != nil {
		if r.instance.Statement.Model != nil {
			return r.instance.Statement.Model
		}
	}
	return nil
}

func (r *Query) parentTable(parent any) string {
	if t, err := tableNameFor(r.instance, parent); err == nil {
		return t
	}
	return ""
}

// parseRelationArgs unpacks variadic args of unknown order/length into (callback, op, count).
//
// Acceptable shapes (any combination):
//   - (callback)
//   - (callback, op)
//   - (callback, op, count)
//   - (op)
//   - (op, count)
//
// Defaults: op = ">=", count = 1.
func parseRelationArgs(args []any) (contractsorm.RelationCallback, string, int, error) {
	op := ">="
	count := 1
	var cb contractsorm.RelationCallback

	for _, arg := range args {
		switch v := arg.(type) {
		case nil:
			continue
		case contractsorm.RelationCallback:
			cb = v
		case func(contractsorm.Query) contractsorm.Query:
			cb = contractsorm.RelationCallback(v)
		case string:
			op = v
		case int:
			count = v
		case int64:
			count = int(v)
		default:
			return nil, op, count, errors.OrmRelationInvalidArgument.Args(v)
		}
	}
	return cb, op, count, nil
}

// parseMorphRelationArgs is a variant that also accepts MorphRelationCallback.
func parseMorphRelationArgs(args []any) (contractsorm.RelationCallback, contractsorm.MorphRelationCallback, string, int, error) {
	op := ">="
	count := 1
	var cb contractsorm.RelationCallback
	var mcb contractsorm.MorphRelationCallback

	for _, arg := range args {
		switch v := arg.(type) {
		case nil:
			continue
		case contractsorm.MorphRelationCallback:
			mcb = v
		case func(contractsorm.Query, string) contractsorm.Query:
			mcb = contractsorm.MorphRelationCallback(v)
		case contractsorm.RelationCallback:
			cb = v
		case func(contractsorm.Query) contractsorm.Query:
			cb = contractsorm.RelationCallback(v)
		case string:
			op = v
		case int:
			count = v
		case int64:
			count = int(v)
		default:
			return nil, nil, op, count, errors.OrmRelationInvalidArgument.Args(v)
		}
	}
	return cb, mcb, op, count, nil
}

// shouldUseExists is the standard optimisation: comparisons of the form ">= 1" or "< 1" can be
// emitted as cheaper EXISTS / NOT EXISTS instead of a full COUNT(*) sub-select.
func shouldUseExists(op string, count int) bool {
	return (op == ">=" || op == "<") && count == 1
}

func validAggregateFn(fn string) bool {
	switch fn {
	case "count", "max", "min", "sum", "avg", "exists":
		return true
	}
	return false
}

func aggregateAlias(relation, fn, column string) string {
	rel := str.Of(strings.ReplaceAll(relation, ".", "_")).Snake().String()
	if column == "*" || column == "" {
		return fmt.Sprintf("%s_%s", rel, fn)
	}
	return fmt.Sprintf("%s_%s_%s", rel, fn, column)
}

// quoteIdent applies a portable identifier quote. We use backticks - GORM rewrites identifiers
// per dialect during compilation when the value passes through `clause.Column`, but raw SQL
// fragments (which is what we emit for correlation) are left alone. Backticks work on MySQL and
// SQLite by default; PostgreSQL and SQL Server only accept double quotes. To stay portable we
// emit the bare identifier and let the dialects accept it - all tested dialects parse unquoted
// identifiers when the names are simple snake_case.
func quoteIdent(name string) string {
	return strings.NewReplacer("`", "", "'", "", `"`, "").Replace(name)
}

// Sanity: ensure *Query satisfies contractsorm.QueryWithRelations at compile time.
var _ contractsorm.QueryWithRelations = (*Query)(nil)
