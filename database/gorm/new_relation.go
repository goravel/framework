package gorm

import (
	"fmt"
	"reflect"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm/morphmap"
	"github.com/goravel/framework/errors"
)

// Related is the public Query-level entry point. The Orm-level entry (Orm.Related) delegates
// here. Returns a fresh Query pre-scoped to the related rows for parent.relation.
func (r *Query) Related(parent any, relation string) contractsorm.Query {
	return r.newRelationQuery(parent, relation)
}

// newRelationQuery returns a fresh Query pre-scoped to the related rows for parent.relation.
// Mirrors fedaco's model.NewRelation('foo') for the read path. Caller can chain Where / OrderBy
// / Get / etc. on the returned Query. Write operations live on RelationWriter (see Query.Relation).
//
// Public entry: Orm.Related delegates here on a fresh-session *Query.
func (r *Query) newRelationQuery(parent any, relation string) contractsorm.Query {
	if !isValidParent(parent) {
		return r.guardedQuery(errors.OrmRelationParentNotPointer.Args(parent))
	}

	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return r.guardedQuery(err)
	}

	var built contractsorm.Query
	switch desc.kind {
	case relKindHasOne, relKindHasMany:
		built = r.newHasOneOrManyQuery(parent, desc)
	case relKindBelongsTo:
		built = r.newBelongsToQuery(parent, desc)
	case relKindMorphOne, relKindMorphMany:
		built = r.newMorphOneOrManyQuery(parent, desc)
	case relKindMorphTo:
		built = r.newMorphToQuery(parent, desc)
	case relKindMany2Many:
		built = r.newMany2ManyQuery(parent, desc, false)
	case relKindMorphToMany:
		built = r.newMany2ManyQuery(parent, desc, true)
	case relKindHasOneThrough, relKindHasManyThrough:
		built = r.newThroughQuery(parent, desc)
	default:
		return r.guardedQuery(errors.OrmRelationUnsupported.Args(relation, fmt.Sprintf("%T", parent), fmt.Sprintf("kind=%d", desc.kind)))
	}
	// Apply the relation's default scope on top of the per-kind constraints. Caller still gets
	// a chainable Query and can layer further conditions via Where / OrderBy / etc.
	if desc.onQuery != nil {
		built = desc.onQuery(built)
	}
	return built
}

// --- per-kind builders -----------------------------------------------------

// newHasOneOrManyQuery: SELECT * FROM <related> WHERE <related>.<id_col> = parent.<local_key>.
func (r *Query) newHasOneOrManyQuery(parent any, desc *relationDescriptor) contractsorm.Query {
	if len(desc.references) == 0 {
		return r.guardedQuery(errors.OrmRelationUnsupported.Args(desc.name, desc.parentTable, "no references"))
	}
	ref := desc.references[0]
	parentVal, err := readParentColumn(r, parent, ref.primaryColumn)
	if err != nil {
		return r.guardedQuery(err)
	}
	return r.relatedQuery(desc.relatedModel).Where(ref.foreignColumn, parentVal)
}

// newBelongsToQuery: SELECT * FROM <related> WHERE <related>.<owner_key> = parent.<fk_col>.
func (r *Query) newBelongsToQuery(parent any, desc *relationDescriptor) contractsorm.Query {
	if len(desc.references) == 0 {
		return r.guardedQuery(errors.OrmRelationUnsupported.Args(desc.name, desc.parentTable, "no references"))
	}
	ref := desc.references[0]
	parentVal, err := readParentColumn(r, parent, ref.foreignColumn)
	if err != nil {
		return r.guardedQuery(err)
	}
	return r.relatedQuery(desc.relatedModel).Where(ref.primaryColumn, parentVal)
}

// newMorphOneOrManyQuery: HasMany shape + WHERE <related>.<type_col> = desc.morphValue.
func (r *Query) newMorphOneOrManyQuery(parent any, desc *relationDescriptor) contractsorm.Query {
	if len(desc.references) == 0 {
		return r.guardedQuery(errors.OrmRelationUnsupported.Args(desc.name, desc.parentTable, "no references"))
	}
	ref := desc.references[0]
	parentVal, err := readParentColumn(r, parent, ref.primaryColumn)
	if err != nil {
		return r.guardedQuery(err)
	}
	return r.relatedQuery(desc.relatedModel).
		Where(ref.foreignColumn, parentVal).
		Where(desc.morphTypeColumn, desc.morphValue)
}

// newMorphToQuery resolves the related model per-row from the parent's <type_col> via the morph
// map, then issues SELECT * FROM <resolved> WHERE <resolved>.<owner_key> = parent.<id_col>.
//
// If the parent's *_type is empty or unregistered, the returned Query is guarded so subsequent
// Get / First yields no rows without a database round-trip.
func (r *Query) newMorphToQuery(parent any, desc *relationDescriptor) contractsorm.Query {
	morphType, err := readParentColumn(r, parent, desc.morphTypeColumn)
	if err != nil {
		return r.guardedQuery(err)
	}
	morphID, err := readParentColumn(r, parent, desc.morphIDColumn)
	if err != nil {
		return r.guardedQuery(err)
	}
	typeStr := morphTypeToString(morphType)
	if typeStr == "" {
		return r.zeroRowQuery()
	}
	sample := morphmap.Find(typeStr)
	if sample == nil {
		return r.guardedQuery(errors.OrmMorphTypeUnknown.Args(typeStr))
	}
	ownerKey := desc.morphOwnerKey
	if ownerKey == "" {
		ownerKey = "id"
	}
	return r.relatedQuery(sample).Where(ownerKey, morphID)
}

// newMany2ManyQuery: SELECT <related>.* FROM <related> INNER JOIN <pivot>
//
//	ON <related>.<rel_pk> = <pivot>.<rel_fk>
//	WHERE <pivot>.<parent_fk> = parent.<pk>  [AND <pivot>.<type_col> = morphValue]
//
// Pivot columns are not surfaced — call sites that need them should add Select / a future
// WithPivot helper.
func (r *Query) newMany2ManyQuery(parent any, desc *relationDescriptor, isMorph bool) contractsorm.Query {
	parentVal, err := readParentColumn(r, parent, desc.pivotParentRef.primaryColumn)
	if err != nil {
		return r.guardedQuery(err)
	}
	relatedTable := desc.relatedTable
	pivotTable := desc.pivotTable

	q := r.relatedQuery(desc.relatedModel).
		Join(fmt.Sprintf("INNER JOIN %s ON %s.%s = %s.%s",
			quoteIdent(pivotTable),
			quoteIdent(pivotTable), quoteIdent(desc.pivotRelatedRef.foreignColumn),
			quoteIdent(relatedTable), quoteIdent(desc.pivotRelatedRef.primaryColumn))).
		Where(fmt.Sprintf("%s.%s = ?", quoteIdent(pivotTable), quoteIdent(desc.pivotParentRef.foreignColumn)), parentVal)
	if isMorph {
		q = q.Where(fmt.Sprintf("%s.%s = ?", quoteIdent(pivotTable), quoteIdent(desc.morphTypeColumn)), desc.morphValue)
	}
	return applyOnPivotQueryToQuery(q, desc)
}

// newThroughQuery:
//
//	SELECT <related>.* FROM <related>
//	INNER JOIN <through> ON <related>.<second_key> = <through>.<second_local_key>
//	WHERE <through>.<first_key> = parent.<local_key>
func (r *Query) newThroughQuery(parent any, desc *relationDescriptor) contractsorm.Query {
	parentVal, err := readParentColumn(r, parent, desc.localKey)
	if err != nil {
		return r.guardedQuery(err)
	}
	return r.relatedQuery(desc.relatedModel).
		Join(fmt.Sprintf("INNER JOIN %s ON %s.%s = %s.%s",
			quoteIdent(desc.throughTable),
			quoteIdent(desc.relatedTable), quoteIdent(desc.secondKey),
			quoteIdent(desc.throughTable), quoteIdent(desc.secondLocalKey))).
		Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.throughTable), quoteIdent(desc.firstKey)), parentVal)
}

// --- helpers ---------------------------------------------------------------

// relatedQuery returns a fresh Query bound to the given related model, sharing connection /
// config / context with the receiver. Subsequent unqualified Where("col", v) calls default to
// the related table's column space.
func (r *Query) relatedQuery(related any) contractsorm.Query {
	q := r.wrap(r.freshSession())
	return q.Model(related)
}

// guardedQuery returns a Query guarded so subsequent terminals (Get/First/Count/...) immediately
// surface err to the caller.
func (r *Query) guardedQuery(err error) contractsorm.Query {
	q := r.wrap(r.freshSession())
	_ = q.instance.AddError(err)
	return q
}

// zeroRowQuery returns a Query whose Get/First yields no rows without a database round-trip —
// the WHERE evaluates to a guaranteed-false condition.
func (r *Query) zeroRowQuery() contractsorm.Query {
	return r.wrap(r.freshSession()).Where("1 = 0")
}

// readParentColumn reads the value of the given DB column from parent via GORM's parsed schema.
func readParentColumn(r *Query, parent any, dbColumn string) (any, error) {
	parentSchema, err := parseGormSchema(r.instance, parent)
	if err != nil {
		return nil, err
	}
	field, ok := parentSchema.FieldsByDBName[dbColumn]
	if !ok {
		return nil, errors.OrmRelationUnsupported.Args("", parentSchema.Name, "no parent field for "+dbColumn)
	}
	rv := reflect.ValueOf(parent)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	val, _ := field.ValueOf(r.ctx, rv)
	return val, nil
}

// morphTypeToString coerces a morph_type column value (which the driver may return as string,
// []byte, or sql.NullString) into a plain string.
func morphTypeToString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case []byte:
		return string(x)
	}
	return fmt.Sprint(v)
}

// isValidParent returns true if v is a non-nil pointer to a struct.
func isValidParent(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return false
	}
	return rv.Elem().Kind() == reflect.Struct
}
