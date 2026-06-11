package gorm

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	dbcontract "github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/errors"
)

// SaveRelation inserts or updates child as a member of parent's relation. Sets child's foreign
// key (and morph_type for MorphOne/MorphMany) from parent's local key, then persists child via
// Query.Save. For BelongsToMany kinds (Many2Many, MorphToMany, MorphedByMany) persists child first,
// then writes a pivot row linking parent and child.
//
// Public Query-level helper used by Orm.Save. Named "SaveRelation" to avoid clashing with the
// existing single-arg Query.Save(value any) which persists a model directly.
//
// Supported kinds: HasOne, HasMany, MorphOne, MorphMany, Many2Many, MorphToMany, MorphedByMany.
// Other kinds error with OrmRelationKindNotSupported.
func (r *Query) SaveRelation(parent any, relation string, child any) error {
	if !isValidParent(parent) {
		return errors.OrmRelationParentNotPointer.Args(parent)
	}
	if !isValidParent(child) {
		return errors.OrmRelationParentNotPointer.Args(child)
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return err
	}
	switch desc.kind {
	case relKindHasOne, relKindHasMany, relKindMorphOne, relKindMorphMany:
		if err := r.setRelationFKOnChild(parent, child, desc); err != nil {
			return err
		}
		return r.wrap(r.freshSession()).Save(child)
	case relKindMany2Many, relKindMorphToMany:
		// Persist child first, then attach via pivot.
		if err := r.wrap(r.freshSession()).Save(child); err != nil {
			return err
		}
		childPK, err := readParentColumn(r, child, desc.pivotRelatedRef.primaryColumn)
		if err != nil {
			return err
		}
		return r.AttachRelation(parent, relation, []any{childPK})
	default:
		return errors.OrmRelationKindNotSupported.Args("Save", relation, kindName(desc.kind))
	}
}

// SaveManyRelation is the slice form of SaveRelation. children must be a slice or pointer-to-
// slice of either pointer-to-struct or struct elements. Iterates and bails on first error.
func (r *Query) SaveManyRelation(parent any, relation string, children any) error {
	rv := reflect.ValueOf(children)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return errors.OrmRelationKindNotSupported.Args("SaveMany", relation, fmt.Sprintf("children=%T (must be slice)", children))
	}
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		var elem any
		switch item.Kind() {
		case reflect.Pointer:
			elem = item.Interface()
		case reflect.Struct:
			if !item.CanAddr() {
				ptr := reflect.New(item.Type())
				ptr.Elem().Set(item)
				elem = ptr.Interface()
			} else {
				elem = item.Addr().Interface()
			}
		default:
			return errors.OrmRelationKindNotSupported.Args("SaveMany", relation, fmt.Sprintf("children element=%s", item.Kind()))
		}
		if err := r.SaveRelation(parent, relation, elem); err != nil {
			return err
		}
	}
	return nil
}

// SaveRelationWithPivot is SaveRelation with caller-supplied pivot column values for the
// BelongsToMany family. On HasOneOrMany kinds attrs is ignored (no pivot row).
func (r *Query) SaveRelationWithPivot(parent any, relation string, child any, attrs map[string]any) error {
	if !isValidParent(parent) {
		return errors.OrmRelationParentNotPointer.Args(parent)
	}
	if !isValidParent(child) {
		return errors.OrmRelationParentNotPointer.Args(child)
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return err
	}
	switch desc.kind {
	case relKindHasOne, relKindHasMany, relKindMorphOne, relKindMorphMany:
		// No pivot — just delegate to SaveRelation.
		return r.SaveRelation(parent, relation, child)
	case relKindMany2Many, relKindMorphToMany:
		// Persist child first, then attach via pivot with attrs.
		if err := r.wrap(r.freshSession()).Save(child); err != nil {
			return err
		}
		childPK, err := readParentColumn(r, child, desc.pivotRelatedRef.primaryColumn)
		if err != nil {
			return err
		}
		return r.AttachWithPivotRelation(parent, relation, map[any]map[string]any{childPK: attrs})
	default:
		return errors.OrmRelationKindNotSupported.Args("SaveWithPivot", relation, kindName(desc.kind))
	}
}

// SaveManyRelationWithPivot is the slice form of SaveRelationWithPivot. attrsPerChild is keyed by
// the related PK of each child; an entry may be nil to attach without extra columns.
func (r *Query) SaveManyRelationWithPivot(parent any, relation string, children any, attrsPerChild map[any]map[string]any) error {
	rv := reflect.ValueOf(children)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return errors.OrmRelationKindNotSupported.Args("SaveManyWithPivot", relation, fmt.Sprintf("children=%T (must be slice)", children))
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return err
	}
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		var elem any
		switch item.Kind() {
		case reflect.Pointer:
			elem = item.Interface()
		case reflect.Struct:
			if !item.CanAddr() {
				ptr := reflect.New(item.Type())
				ptr.Elem().Set(item)
				elem = ptr.Interface()
			} else {
				elem = item.Addr().Interface()
			}
		default:
			return errors.OrmRelationKindNotSupported.Args("SaveManyWithPivot", relation, fmt.Sprintf("children element=%s", item.Kind()))
		}
		// Read child's PK to look up attrs.
		childPK, err := readParentColumn(r, elem, desc.pivotRelatedRef.primaryColumn)
		if err != nil {
			return err
		}
		attrs := attrsPerChild[childPK]
		if err := r.SaveRelationWithPivot(parent, relation, elem, attrs); err != nil {
			return err
		}
	}
	return nil
}

// AssociateRelation sets parent's foreign key (and morph_type for MorphTo) to point at owner,
// then persists parent. Supported kinds: BelongsTo, MorphTo. owner must be a non-nil pointer to
// a struct.
//
// Public Query-level helper used by Orm.Associate.
func (r *Query) AssociateRelation(parent any, relation string, owner any) error {
	if !isValidParent(parent) {
		return errors.OrmRelationParentNotPointer.Args(parent)
	}
	if !isValidParent(owner) {
		return errors.OrmRelationParentNotPointer.Args(owner)
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return err
	}
	switch desc.kind {
	case relKindBelongsTo:
		return r.applyAssociate(parent, owner, desc, false)
	case relKindMorphTo:
		return r.applyAssociate(parent, owner, desc, true)
	default:
		return errors.OrmRelationKindNotSupported.Args("Associate", relation, kindName(desc.kind))
	}
}

// DissociateRelation clears parent's foreign key (and morph_type for MorphTo) and persists
// parent. Supported kinds: BelongsTo, MorphTo.
func (r *Query) DissociateRelation(parent any, relation string) error {
	if !isValidParent(parent) {
		return errors.OrmRelationParentNotPointer.Args(parent)
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return err
	}
	switch desc.kind {
	case relKindBelongsTo:
		return r.applyDissociate(parent, desc, false)
	case relKindMorphTo:
		return r.applyDissociate(parent, desc, true)
	default:
		return errors.OrmRelationKindNotSupported.Args("Dissociate", relation, kindName(desc.kind))
	}
}

// applyAssociate writes owner's PK into parent's FK column (and the morph_type column for
// MorphTo, resolved from the morph map / MorphClass()), then persists parent.
func (r *Query) applyAssociate(parent, owner any, desc *relationDescriptor, isMorph bool) error {
	if err := r.mutateAssociate(parent, owner, desc, isMorph); err != nil {
		return err
	}
	return r.wrap(r.freshSession()).Save(parent)
}

// mutateAssociate is the pure-mutation half of applyAssociate. Writes owner's PK into parent's
// FK column (and the morph_type column for MorphTo). No persistence.
func (r *Query) mutateAssociate(parent, owner any, desc *relationDescriptor, isMorph bool) error {
	parentSchema, err := parseGormSchema(r.instance, parent)
	if err != nil {
		return err
	}
	parentRV := reflect.ValueOf(parent).Elem()

	var fkColumn string
	if isMorph {
		fkColumn = desc.morphIDColumn
	} else {
		if len(desc.references) == 0 {
			return errors.OrmRelationUnsupported.Args(desc.name, desc.parentTable, "no references")
		}
		fkColumn = desc.references[0].foreignColumn
	}
	fkField, ok := parentSchema.FieldsByDBName[fkColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(desc.name, parentSchema.Name, "no FK field "+fkColumn)
	}

	ownerPKColumn := "id"
	if !isMorph && len(desc.references) > 0 {
		ownerPKColumn = desc.references[0].primaryColumn
	} else if isMorph && desc.morphOwnerKey != "" {
		ownerPKColumn = desc.morphOwnerKey
	}
	ownerPK, err := readParentColumn(r, owner, ownerPKColumn)
	if err != nil {
		return err
	}
	if err := fkField.Set(r.ctx, parentRV, ownerPK); err != nil {
		return err
	}

	if isMorph {
		typeField, ok := parentSchema.FieldsByDBName[desc.morphTypeColumn]
		if !ok {
			return errors.OrmRelationUnsupported.Args(desc.name, parentSchema.Name, "no morph type field "+desc.morphTypeColumn)
		}
		alias, ok := resolveMorphAlias(owner)
		if !ok {
			tbl, terr := tableNameFor(r.instance, owner)
			if terr != nil {
				return terr
			}
			alias = tbl
		}
		if err := typeField.Set(r.ctx, parentRV, alias); err != nil {
			return err
		}
	}
	return nil
}

// applyDissociate sets parent's FK to the zero value (and morph_type to "" for MorphTo), then
// persists parent.
func (r *Query) applyDissociate(parent any, desc *relationDescriptor, isMorph bool) error {
	if err := r.mutateDissociate(parent, desc, isMorph); err != nil {
		return err
	}
	return r.wrap(r.freshSession()).Save(parent)
}

// mutateDissociate is the pure-mutation half of applyDissociate.
func (r *Query) mutateDissociate(parent any, desc *relationDescriptor, isMorph bool) error {
	parentSchema, err := parseGormSchema(r.instance, parent)
	if err != nil {
		return err
	}
	parentRV := reflect.ValueOf(parent).Elem()

	var fkColumn string
	if isMorph {
		fkColumn = desc.morphIDColumn
	} else {
		if len(desc.references) == 0 {
			return errors.OrmRelationUnsupported.Args(desc.name, desc.parentTable, "no references")
		}
		fkColumn = desc.references[0].foreignColumn
	}
	fkField, ok := parentSchema.FieldsByDBName[fkColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(desc.name, parentSchema.Name, "no FK field "+fkColumn)
	}
	zero := reflect.Zero(fkField.FieldType).Interface()
	if err := fkField.Set(r.ctx, parentRV, zero); err != nil {
		return err
	}

	if isMorph {
		typeField, ok := parentSchema.FieldsByDBName[desc.morphTypeColumn]
		if !ok {
			return errors.OrmRelationUnsupported.Args(desc.name, parentSchema.Name, "no morph type field "+desc.morphTypeColumn)
		}
		zeroType := reflect.Zero(typeField.FieldType).Interface()
		if err := typeField.Set(r.ctx, parentRV, zeroType); err != nil {
			return err
		}
	}
	return nil
}

// CreateRelation persists a new related row. For HasOneOrMany kinds (HasOne, HasMany, MorphOne,
// MorphMany) the framework first sets the FK (and morph type column) on dest from parent, then
// inserts. dest must be a non-nil pointer to a struct of the related type.
//
// Public Query-level helper used by Orm.Create.
func (r *Query) CreateRelation(parent any, relation string, dest any) error {
	if !isValidParent(parent) {
		return errors.OrmRelationParentNotPointer.Args(parent)
	}
	if !isValidParent(dest) {
		return errors.OrmRelationParentNotPointer.Args(dest)
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return err
	}
	switch desc.kind {
	case relKindHasOne, relKindHasMany, relKindMorphOne, relKindMorphMany:
		if err := r.setRelationFKOnChild(parent, dest, desc); err != nil {
			return err
		}
		return r.wrap(r.freshSession()).Create(dest)
	case relKindMany2Many, relKindMorphToMany:
		// Create dest first, then attach via pivot.
		if err := r.wrap(r.freshSession()).Create(dest); err != nil {
			return err
		}
		childPK, err := readParentColumn(r, dest, desc.pivotRelatedRef.primaryColumn)
		if err != nil {
			return err
		}
		return r.AttachRelation(parent, relation, []any{childPK})
	default:
		return errors.OrmRelationKindNotSupported.Args("Create", relation, kindName(desc.kind))
	}
}

// CreateManyRelation is the slice form of CreateRelation. dests must be a slice or a pointer to a
// slice; iterates and calls CreateRelation per element, bailing on the first error.
func (r *Query) CreateManyRelation(parent any, relation string, dests any) error {
	rv := reflect.ValueOf(dests)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return errors.OrmRelationUnsupported.Args(relation, fmt.Sprintf("%T", parent), "CreateMany requires a slice")
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		if elem.Kind() != reflect.Pointer {
			elem = elem.Addr()
		}
		if err := r.CreateRelation(parent, relation, elem.Interface()); err != nil {
			return err
		}
	}
	return nil
}

// FindOrNewRelation finds the related row with primary key id. If absent, fills dest with a new
// instance of the related model and pre-sets the FK (and morph type) — but does NOT persist.
// dest must be a pointer to a struct.
func (r *Query) FindOrNewRelation(parent any, relation string, id any, dest any) error {
	if !isValidParent(parent) {
		return errors.OrmRelationParentNotPointer.Args(parent)
	}
	if !isValidParent(dest) {
		return errors.OrmRelationParentNotPointer.Args(dest)
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return err
	}
	switch desc.kind {
	case relKindHasOne, relKindHasMany, relKindMorphOne, relKindMorphMany:
		q := r.newRelationQuery(parent, relation)
		if err := q.Find(dest, id); err != nil {
			return err
		}
		// Check if Find actually populated dest by inspecting the PK field.
		schema, err := parseGormSchema(r.instance, dest)
		if err != nil {
			return err
		}
		if len(schema.PrimaryFields) == 0 {
			return errors.OrmRelationUnsupported.Args(relation, schema.Name, "no primary key")
		}
		pkField := schema.PrimaryFields[0]
		rv := reflect.ValueOf(dest).Elem()
		pkVal, isZero := pkField.ValueOf(r.ctx, rv)
		_ = pkVal
		if isZero {
			// Not found — set FK on the zero-valued dest.
			return r.setRelationFKOnChild(parent, dest, desc)
		}
		return nil
	default:
		return errors.OrmRelationKindNotSupported.Args("FindOrNew", relation, kindName(desc.kind))
	}
}

// FirstOrNewRelation finds the first related row matching attrs. If absent, fills dest with a new
// instance carrying attrs+values and pre-set FK — does NOT persist.
func (r *Query) FirstOrNewRelation(parent any, relation string, attrs map[string]any, values map[string]any, dest any) error {
	if !isValidParent(parent) {
		return errors.OrmRelationParentNotPointer.Args(parent)
	}
	if !isValidParent(dest) {
		return errors.OrmRelationParentNotPointer.Args(dest)
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return err
	}
	switch desc.kind {
	case relKindHasOne, relKindHasMany, relKindMorphOne, relKindMorphMany:
		q := r.newRelationQuery(parent, relation)
		for col, val := range attrs {
			q = q.Where(col, val)
		}
		if err := q.First(dest); err != nil {
			// Not found — overlay attrs+values and set FK.
			if err := r.applyAttrMap(dest, attrs); err != nil {
				return err
			}
			if err := r.applyAttrMap(dest, values); err != nil {
				return err
			}
			return r.setRelationFKOnChild(parent, dest, desc)
		}
		return nil
	default:
		return errors.OrmRelationKindNotSupported.Args("FirstOrNew", relation, kindName(desc.kind))
	}
}

// FirstOrCreateRelation is FirstOrNewRelation that persists when no matching row exists.
func (r *Query) FirstOrCreateRelation(parent any, relation string, attrs map[string]any, values map[string]any, dest any) error {
	if !isValidParent(parent) {
		return errors.OrmRelationParentNotPointer.Args(parent)
	}
	if !isValidParent(dest) {
		return errors.OrmRelationParentNotPointer.Args(dest)
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return err
	}
	switch desc.kind {
	case relKindHasOne, relKindHasMany, relKindMorphOne, relKindMorphMany:
		q := r.newRelationQuery(parent, relation)
		for col, val := range attrs {
			q = q.Where(col, val)
		}
		if err := q.First(dest); err != nil {
			// Not found — overlay attrs+values, set FK, persist.
			if err := r.applyAttrMap(dest, attrs); err != nil {
				return err
			}
			if err := r.applyAttrMap(dest, values); err != nil {
				return err
			}
			if err := r.setRelationFKOnChild(parent, dest, desc); err != nil {
				return err
			}
			return r.wrap(r.freshSession()).Create(dest)
		}
		return nil
	case relKindMany2Many, relKindMorphToMany:
		// For m2m: search the related table directly (no FK constraint), create+attach if missing.
		q := r.freshSession().Model(desc.relatedModel)
		for col, val := range attrs {
			q = q.Where(col, val)
		}
		if err := r.wrap(q).First(dest); err != nil {
			// Not found — overlay attrs+values, create, attach.
			if err := r.applyAttrMap(dest, attrs); err != nil {
				return err
			}
			if err := r.applyAttrMap(dest, values); err != nil {
				return err
			}
			if err := r.wrap(r.freshSession()).Create(dest); err != nil {
				return err
			}
			childPK, err := readParentColumn(r, dest, desc.pivotRelatedRef.primaryColumn)
			if err != nil {
				return err
			}
			return r.AttachRelation(parent, relation, []any{childPK})
		}
		return nil
	default:
		return errors.OrmRelationKindNotSupported.Args("FirstOrCreate", relation, kindName(desc.kind))
	}
}

// UpdateOrCreateRelation finds the first related row matching attrs (or creates one), then overlays
// values onto it and persists. Always saves dest.
func (r *Query) UpdateOrCreateRelation(parent any, relation string, attrs map[string]any, values map[string]any, dest any) error {
	if !isValidParent(parent) {
		return errors.OrmRelationParentNotPointer.Args(parent)
	}
	if !isValidParent(dest) {
		return errors.OrmRelationParentNotPointer.Args(dest)
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return err
	}
	switch desc.kind {
	case relKindHasOne, relKindHasMany, relKindMorphOne, relKindMorphMany:
		// FirstOrNew logic.
		q := r.newRelationQuery(parent, relation)
		for col, val := range attrs {
			q = q.Where(col, val)
		}
		if err := q.First(dest); err != nil {
			// Not found — overlay attrs+values and set FK.
			if err := r.applyAttrMap(dest, attrs); err != nil {
				return err
			}
			if err := r.applyAttrMap(dest, values); err != nil {
				return err
			}
			if err := r.setRelationFKOnChild(parent, dest, desc); err != nil {
				return err
			}
		} else {
			// Found — overlay values only.
			if err := r.applyAttrMap(dest, values); err != nil {
				return err
			}
		}
		return r.wrap(r.freshSession()).Save(dest)
	case relKindMany2Many, relKindMorphToMany:
		// For m2m: search the related table directly, create+attach if missing, otherwise update.
		q := r.freshSession().Model(desc.relatedModel)
		for col, val := range attrs {
			q = q.Where(col, val)
		}
		var freshlyCreated bool
		if err := r.wrap(q).First(dest); err != nil {
			// Not found — overlay attrs+values, create.
			if err := r.applyAttrMap(dest, attrs); err != nil {
				return err
			}
			if err := r.applyAttrMap(dest, values); err != nil {
				return err
			}
			if err := r.wrap(r.freshSession()).Create(dest); err != nil {
				return err
			}
			freshlyCreated = true
		} else {
			// Found — overlay values only.
			if err := r.applyAttrMap(dest, values); err != nil {
				return err
			}
			if err := r.wrap(r.freshSession()).Save(dest); err != nil {
				return err
			}
		}
		// Attach if freshly created.
		if freshlyCreated {
			childPK, err := readParentColumn(r, dest, desc.pivotRelatedRef.primaryColumn)
			if err != nil {
				return err
			}
			return r.AttachRelation(parent, relation, []any{childPK})
		}
		return nil
	default:
		return errors.OrmRelationKindNotSupported.Args("UpdateOrCreate", relation, kindName(desc.kind))
	}
}

// AttachRelation inserts pivot rows linking parent to each id in ids. Skips ids that already
// have a pivot row. Supported kinds: Many2Many, MorphToMany, MorphedByMany.
//
// Public Query-level helper used by Orm.Attach.
func (r *Query) AttachRelation(parent any, relation string, ids []any) error {
	desc, parentVal, err := r.resolvePivot(parent, relation, "Attach")
	if err != nil {
		return err
	}
	affected, err := r.doAttach(desc, parentVal, ids, nil)
	if err != nil {
		return err
	}
	if affected > 0 {
		return r.touchIfTouching(desc, parent, parentVal)
	}
	return nil
}

// AttachWithPivotRelation is Attach with per-row pivot column values.
func (r *Query) AttachWithPivotRelation(parent any, relation string, idsWithAttrs map[any]map[string]any) error {
	desc, parentVal, err := r.resolvePivot(parent, relation, "AttachWithPivot")
	if err != nil {
		return err
	}
	affected, err := r.doAttachWithPivot(desc, parentVal, idsWithAttrs)
	if err != nil {
		return err
	}
	if affected > 0 {
		return r.touchIfTouching(desc, parent, parentVal)
	}
	return nil
}

// doAttach is the shared work-horse for AttachRelation / AttachWithPivotRelation. It does not
// trigger touchIfTouching — callers (the public methods, and syncCore) own when to touch. Returns
// the number of newly-inserted pivot rows; zero when every id was already attached.
//
// idsWithAttrs is optional: when non-nil, attrs from this map are merged into each new row;
// when nil, ids alone drive the insert.
func (r *Query) doAttach(desc *relationDescriptor, parentVal any, ids []any, idsWithAttrs map[any]map[string]any) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	existing, err := r.existingPivotIDs(desc, parentVal, ids)
	if err != nil {
		return 0, err
	}
	rows := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		if _, dup := existing[dictKey(id)]; dup {
			continue
		}
		var attrs map[string]any
		if idsWithAttrs != nil {
			attrs = idsWithAttrs[id]
		}
		rows = append(rows, r.basePivotRow(desc, parentVal, id, attrs))
	}
	if len(rows) == 0 {
		return 0, nil
	}
	if err := r.freshSession().Table(desc.pivotTable).Create(rows).Error; err != nil {
		return 0, err
	}
	return int64(len(rows)), nil
}

// doAttachWithPivot is doAttach for the with-attrs entry shape (map keyed by id). Equivalent in
// outcome to doAttach with the same attrs; kept separate to avoid materialising an ids slice
// just to satisfy the doAttach signature when callers already have a map.
func (r *Query) doAttachWithPivot(desc *relationDescriptor, parentVal any, idsWithAttrs map[any]map[string]any) (int64, error) {
	if len(idsWithAttrs) == 0 {
		return 0, nil
	}
	ids := make([]any, 0, len(idsWithAttrs))
	for id := range idsWithAttrs {
		ids = append(ids, id)
	}
	existing, err := r.existingPivotIDs(desc, parentVal, ids)
	if err != nil {
		return 0, err
	}
	rows := make([]map[string]any, 0, len(ids))
	for id, attrs := range idsWithAttrs {
		if _, dup := existing[dictKey(id)]; dup {
			continue
		}
		rows = append(rows, r.basePivotRow(desc, parentVal, id, attrs))
	}
	if len(rows) == 0 {
		return 0, nil
	}
	if err := r.freshSession().Table(desc.pivotTable).Create(rows).Error; err != nil {
		return 0, err
	}
	return int64(len(rows)), nil
}

// DetachRelation removes pivot rows linking parent to the given ids. With nil/empty ids, removes
// all pivot rows for parent (and morph type, for polymorphic). Returns the number of rows
// removed.
func (r *Query) DetachRelation(parent any, relation string, ids []any) (int64, error) {
	desc, parentVal, err := r.resolvePivot(parent, relation, "Detach")
	if err != nil {
		return 0, err
	}
	affected, err := r.doDetach(desc, parentVal, ids)
	if err != nil {
		return 0, err
	}
	if affected > 0 {
		if err := r.touchIfTouching(desc, parent, parentVal); err != nil {
			return affected, err
		}
	}
	return affected, nil
}

// doDetach is the no-touch worker behind DetachRelation. Returns rows affected by the DELETE.
func (r *Query) doDetach(desc *relationDescriptor, parentVal any, ids []any) (int64, error) {
	q := r.freshSession().Table(desc.pivotTable).
		Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.pivotParentRef.foreignColumn)), parentVal)
	if desc.kind == relKindMorphToMany {
		q = q.Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.morphTypeColumn)), desc.morphValue)
	}
	if len(ids) > 0 {
		q = q.Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.pivotTable), quoteIdent(desc.pivotRelatedRef.foreignColumn)), ids)
	}
	q = applyOnPivotQuery(q, desc)
	res := q.Delete(nil)
	return res.RowsAffected, res.Error
}

// resolvePivot is the shared front-half for all pivot operations: validates parent, resolves the
// descriptor, asserts a pivot-friendly kind, and reads the parent's PK that anchors every pivot
// row. Returns the descriptor + parent's PK value.
func (r *Query) resolvePivot(parent any, relation, op string) (*relationDescriptor, any, error) {
	if !isValidParent(parent) {
		return nil, nil, errors.OrmRelationParentNotPointer.Args(parent)
	}
	desc, err := resolveRelation(r.instance, parent, relation)
	if err != nil {
		return nil, nil, err
	}
	if desc.kind != relKindMany2Many && desc.kind != relKindMorphToMany {
		return nil, nil, errors.OrmRelationKindNotSupported.Args(op, relation, kindName(desc.kind))
	}
	parentVal, err := readParentColumn(r, parent, desc.pivotParentRef.primaryColumn)
	if err != nil {
		return nil, nil, err
	}
	return desc, parentVal, nil
}

// basePivotRow builds the column map for one pivot INSERT row. Always includes the parent FK and
// the related FK; for MorphToMany also includes the morph_type column. When the descriptor's
// resolved created_at / updated_at columns are non-empty, stamps those columns with time.Now().
// Caller-supplied attrs are merged on top — the caller wins on column-name conflicts.
func (r *Query) basePivotRow(desc *relationDescriptor, parentVal, relatedID any, attrs map[string]any) map[string]any {
	row := map[string]any{
		desc.pivotParentRef.foreignColumn:  parentVal,
		desc.pivotRelatedRef.foreignColumn: relatedID,
	}
	if desc.kind == relKindMorphToMany {
		row[desc.morphTypeColumn] = desc.morphValue
	}
	now := time.Now()
	if desc.pivotCreatedAtColumn != "" {
		row[desc.pivotCreatedAtColumn] = now
	}
	if desc.pivotUpdatedAtColumn != "" {
		row[desc.pivotUpdatedAtColumn] = now
	}
	for k, v := range attrs {
		row[k] = v
	}
	return row
}

// existingPivotIDs returns the set of already-attached related ids among ids. Used by Attach to
// skip duplicates.
func (r *Query) existingPivotIDs(desc *relationDescriptor, parentVal any, ids []any) (map[string]struct{}, error) {
	q := r.freshSession().
		Table(desc.pivotTable).
		Select(desc.pivotRelatedRef.foreignColumn).
		Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.pivotParentRef.foreignColumn)), parentVal).
		Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.pivotTable), quoteIdent(desc.pivotRelatedRef.foreignColumn)), ids)
	if desc.kind == relKindMorphToMany {
		q = q.Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.morphTypeColumn)), desc.morphValue)
	}
	q = applyOnPivotQuery(q, desc)
	var rows []map[string]any
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		out[dictKey(row[desc.pivotRelatedRef.foreignColumn])] = struct{}{}
	}
	return out, nil
}

// allPivotIDs returns the set of all currently-attached related ids for parent. Used by Sync /
// Toggle to compute the diff.
func (r *Query) allPivotIDs(desc *relationDescriptor, parentVal any) ([]any, error) {
	q := r.freshSession().
		Table(desc.pivotTable).
		Select(desc.pivotRelatedRef.foreignColumn).
		Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.pivotParentRef.foreignColumn)), parentVal)
	if desc.kind == relKindMorphToMany {
		q = q.Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.morphTypeColumn)), desc.morphValue)
	}
	q = applyOnPivotQuery(q, desc)
	var rows []map[string]any
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, row[desc.pivotRelatedRef.foreignColumn])
	}
	return out, nil
}

// SyncRelation replaces parent's pivot rows so they exactly match ids: detaches missing entries,
// attaches new ones, leaves existing untouched. Returns the per-id outcome.
//
// Public Query-level helper used by Orm.Sync.
func (r *Query) SyncRelation(parent any, relation string, ids []any) (*dbcontract.SyncResult, error) {
	return r.syncCore(parent, relation, ids, true /*detach*/, false /*toggle*/, "Sync")
}

// SyncWithoutDetachingRelation is SyncRelation minus the detach step.
func (r *Query) SyncWithoutDetachingRelation(parent any, relation string, ids []any) (*dbcontract.SyncResult, error) {
	return r.syncCore(parent, relation, ids, false /*detach*/, false /*toggle*/, "SyncWithoutDetaching")
}

// ToggleRelation attaches missing entries and detaches existing ones.
func (r *Query) ToggleRelation(parent any, relation string, ids []any) (*dbcontract.SyncResult, error) {
	return r.syncCore(parent, relation, ids, false, true /*toggle*/, "Toggle")
}

// syncCore is the shared engine for Sync / SyncWithoutDetaching / Toggle.
func (r *Query) syncCore(parent any, relation string, ids []any, detachMissing bool, toggle bool, op string) (*dbcontract.SyncResult, error) {
	desc, parentVal, err := r.resolvePivot(parent, relation, op)
	if err != nil {
		return nil, err
	}
	current, err := r.allPivotIDs(desc, parentVal)
	if err != nil {
		return nil, err
	}
	currentSet := make(map[string]any, len(current))
	for _, id := range current {
		currentSet[dictKey(id)] = id
	}
	wantSet := make(map[string]any, len(ids))
	for _, id := range ids {
		wantSet[dictKey(id)] = id
	}

	out := &dbcontract.SyncResult{}
	switch {
	case toggle:
		// Anything in `ids` that exists -> detach; anything that doesn't -> attach.
		var attachIDs, detachIDs []any
		for k, v := range wantSet {
			if _, exists := currentSet[k]; exists {
				detachIDs = append(detachIDs, v)
			} else {
				attachIDs = append(attachIDs, v)
			}
		}
		if len(attachIDs) > 0 {
			if _, err := r.doAttach(desc, parentVal, attachIDs, nil); err != nil {
				return nil, err
			}
		}
		if len(detachIDs) > 0 {
			if _, err := r.doDetach(desc, parentVal, detachIDs); err != nil {
				return nil, err
			}
		}
		out.Attached = castKeys(attachIDs, desc.relatedKeyType)
		out.Detached = castKeys(detachIDs, desc.relatedKeyType)
	default:
		// Attach anything in `wantSet` that isn't yet attached.
		var attachIDs []any
		for k, v := range wantSet {
			if _, exists := currentSet[k]; !exists {
				attachIDs = append(attachIDs, v)
			}
		}
		if len(attachIDs) > 0 {
			if _, err := r.doAttach(desc, parentVal, attachIDs, nil); err != nil {
				return nil, err
			}
		}
		out.Attached = castKeys(attachIDs, desc.relatedKeyType)

		if detachMissing {
			// Detach anything in `currentSet` that isn't in `wantSet`.
			var detachIDs []any
			for k, v := range currentSet {
				if _, keep := wantSet[k]; !keep {
					detachIDs = append(detachIDs, v)
				}
			}
			if len(detachIDs) > 0 {
				if _, err := r.doDetach(desc, parentVal, detachIDs); err != nil {
					return nil, err
				}
			}
			out.Detached = castKeys(detachIDs, desc.relatedKeyType)
		}
	}

	if syncResultChanged(out) {
		if err := r.touchIfTouching(desc, parent, parentVal); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// SyncRelationWithPivot is SyncRelation with per-ID pivot column values. The map key is the
// related id; the map value is the column-name-to-value map applied to that pivot row. For
// existing pivot rows with non-empty attrs, updates the pivot columns (reported in
// SyncResult.Updated). Mirrors fedaco's sync(map).
func (r *Query) SyncRelationWithPivot(parent any, relation string, idsWithAttrs map[any]map[string]any) (*dbcontract.SyncResult, error) {
	return r.syncCoreWithPivot(parent, relation, idsWithAttrs, true /*detach*/, false /*toggle*/, "SyncWithPivot")
}

// SyncRelationWithPivotValues applies the same pivot column values to all ids. Mirrors fedaco's
// syncWithPivotValues.
func (r *Query) SyncRelationWithPivotValues(parent any, relation string, ids []any, pivotValues map[string]any) (*dbcontract.SyncResult, error) {
	idsWithAttrs := make(map[any]map[string]any, len(ids))
	for _, id := range ids {
		idsWithAttrs[id] = pivotValues
	}
	return r.syncCoreWithPivot(parent, relation, idsWithAttrs, true /*detach*/, false /*toggle*/, "SyncWithPivotValues")
}

// SyncWithoutDetachingRelationWithPivot is SyncRelationWithPivot minus the detach step.
func (r *Query) SyncWithoutDetachingRelationWithPivot(parent any, relation string, idsWithAttrs map[any]map[string]any) (*dbcontract.SyncResult, error) {
	return r.syncCoreWithPivot(parent, relation, idsWithAttrs, false /*detach*/, false /*toggle*/, "SyncWithoutDetachingWithPivot")
}

// ToggleRelationWithPivot is ToggleRelation with per-ID pivot column values for newly attached rows.
func (r *Query) ToggleRelationWithPivot(parent any, relation string, idsWithAttrs map[any]map[string]any) (*dbcontract.SyncResult, error) {
	return r.syncCoreWithPivot(parent, relation, idsWithAttrs, false, true /*toggle*/, "ToggleWithPivot")
}

// syncCoreWithPivot is the shared engine for SyncWithPivot / SyncWithPivotValues /
// SyncWithoutDetachingWithPivot / ToggleWithPivot. Similar to syncCore but accepts a map of IDs
// to pivot attributes and updates existing pivot rows when attrs are non-empty.
func (r *Query) syncCoreWithPivot(parent any, relation string, idsWithAttrs map[any]map[string]any, detachMissing bool, toggle bool, op string) (*dbcontract.SyncResult, error) {
	desc, parentVal, err := r.resolvePivot(parent, relation, op)
	if err != nil {
		return nil, err
	}
	current, err := r.allPivotIDs(desc, parentVal)
	if err != nil {
		return nil, err
	}
	currentSet := make(map[string]any, len(current))
	for _, id := range current {
		currentSet[dictKey(id)] = id
	}
	wantSet := make(map[string]any, len(idsWithAttrs))
	for id := range idsWithAttrs {
		wantSet[dictKey(id)] = id
	}

	out := &dbcontract.SyncResult{}
	switch {
	case toggle:
		// Anything in `idsWithAttrs` that exists -> detach; anything that doesn't -> attach with attrs.
		var detachIDs []any
		attachMap := make(map[any]map[string]any)
		for k, v := range wantSet {
			if _, exists := currentSet[k]; exists {
				detachIDs = append(detachIDs, v)
			} else {
				attachMap[v] = idsWithAttrs[v]
			}
		}
		if len(attachMap) > 0 {
			if _, err := r.doAttachWithPivot(desc, parentVal, attachMap); err != nil {
				return nil, err
			}
			for id := range attachMap {
				out.Attached = append(out.Attached, id)
			}
		}
		if len(detachIDs) > 0 {
			if _, err := r.doDetach(desc, parentVal, detachIDs); err != nil {
				return nil, err
			}
		}
		out.Attached = castKeys(out.Attached, desc.relatedKeyType)
		out.Detached = castKeys(detachIDs, desc.relatedKeyType)
	default:
		// Attach anything in `wantSet` that isn't yet attached; update existing if attrs non-empty.
		attachMap := make(map[any]map[string]any)
		var updateIDs []any
		for k, v := range wantSet {
			if _, exists := currentSet[k]; !exists {
				attachMap[v] = idsWithAttrs[v]
			} else {
				// Already attached — if attrs non-empty, update the pivot row.
				attrs := idsWithAttrs[v]
				if len(attrs) > 0 {
					if _, err := r.doUpdateExistingPivot(desc, parentVal, v, attrs); err != nil {
						return nil, err
					}
					updateIDs = append(updateIDs, v)
				}
			}
		}
		if len(attachMap) > 0 {
			if _, err := r.doAttachWithPivot(desc, parentVal, attachMap); err != nil {
				return nil, err
			}
			for id := range attachMap {
				out.Attached = append(out.Attached, id)
			}
		}
		out.Attached = castKeys(out.Attached, desc.relatedKeyType)
		out.Updated = castKeys(updateIDs, desc.relatedKeyType)

		if detachMissing {
			// Detach anything in `currentSet` that isn't in `wantSet`.
			var detachIDs []any
			for k, v := range currentSet {
				if _, keep := wantSet[k]; !keep {
					detachIDs = append(detachIDs, v)
				}
			}
			if len(detachIDs) > 0 {
				if _, err := r.doDetach(desc, parentVal, detachIDs); err != nil {
					return nil, err
				}
			}
			out.Detached = castKeys(detachIDs, desc.relatedKeyType)
		}
	}

	if syncResultChanged(out) {
		if err := r.touchIfTouching(desc, parent, parentVal); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// syncResultChanged reports whether out indicates any actual pivot-table mutation. Used by
// syncCore / syncCoreWithPivot to decide whether to call touchIfTouching at the end.
func syncResultChanged(out *dbcontract.SyncResult) bool {
	return len(out.Attached) > 0 || len(out.Detached) > 0 || len(out.Updated) > 0
}

// UpdateExistingPivotRelation updates pivot columns for an already-attached id. When
// pivotTimestamps is enabled and attrs doesn't already set updated_at, injects time.Now() into
// the update map. No-op (returns 0) if no matching pivot row exists.
func (r *Query) UpdateExistingPivotRelation(parent any, relation string, id any, attrs map[string]any) (int64, error) {
	desc, parentVal, err := r.resolvePivot(parent, relation, "UpdateExistingPivot")
	if err != nil {
		return 0, err
	}
	affected, err := r.doUpdateExistingPivot(desc, parentVal, id, attrs)
	if err != nil {
		return 0, err
	}
	if affected > 0 {
		if err := r.touchIfTouching(desc, parent, parentVal); err != nil {
			return affected, err
		}
	}
	return affected, nil
}

// doUpdateExistingPivot is the no-touch worker behind UpdateExistingPivotRelation. Returns the
// number of pivot rows actually updated. When the descriptor has a resolved updated_at column
// and attrs doesn't already set it, injects time.Now() into the UPDATE map.
func (r *Query) doUpdateExistingPivot(desc *relationDescriptor, parentVal any, id any, attrs map[string]any) (int64, error) {
	if len(attrs) == 0 && desc.pivotUpdatedAtColumn == "" {
		return 0, nil
	}
	updateMap := make(map[string]any, len(attrs)+1)
	for k, v := range attrs {
		updateMap[k] = v
	}
	if desc.pivotUpdatedAtColumn != "" {
		if _, hasUpdatedAt := updateMap[desc.pivotUpdatedAtColumn]; !hasUpdatedAt {
			updateMap[desc.pivotUpdatedAtColumn] = time.Now()
		}
	}
	q := r.freshSession().Table(desc.pivotTable).
		Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.pivotParentRef.foreignColumn)), parentVal).
		Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.pivotRelatedRef.foreignColumn)), id)
	if desc.kind == relKindMorphToMany {
		q = q.Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.morphTypeColumn)), desc.morphValue)
	}
	q = applyOnPivotQuery(q, desc)
	res := q.Updates(updateMap)
	return res.RowsAffected, res.Error
}

// setRelationFKOnChild reads parent's local key, then writes that value into child's FK column
// (and the morph_type column for MorphOne/MorphMany). Mutates child in place; child must be a
// pointer to a struct.
func (r *Query) setRelationFKOnChild(parent, child any, desc *relationDescriptor) error {
	if len(desc.references) == 0 {
		return errors.OrmRelationUnsupported.Args(desc.name, desc.parentTable, "no references")
	}
	ref := desc.references[0]
	parentVal, err := readParentColumn(r, parent, ref.primaryColumn)
	if err != nil {
		return err
	}
	childSchema, err := parseGormSchema(r.instance, child)
	if err != nil {
		return err
	}
	fkField, ok := childSchema.FieldsByDBName[ref.foreignColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(desc.name, childSchema.Name, "no FK field "+ref.foreignColumn)
	}
	rv := reflect.ValueOf(child).Elem()
	if err := fkField.Set(r.ctx, rv, parentVal); err != nil {
		return err
	}
	if desc.kind == relKindMorphOne || desc.kind == relKindMorphMany {
		typeField, ok := childSchema.FieldsByDBName[desc.morphTypeColumn]
		if !ok {
			return errors.OrmRelationUnsupported.Args(desc.name, childSchema.Name, "no morph type field "+desc.morphTypeColumn)
		}
		if err := typeField.Set(r.ctx, rv, desc.morphValue); err != nil {
			return err
		}
	}
	return nil
}

// applyAttrMap overlays attrs onto dest using GORM's parsed schema to map column names to struct
// fields. dest must be a pointer to a struct. Skips columns that don't map to a field.
func (r *Query) applyAttrMap(dest any, attrs map[string]any) error {
	if len(attrs) == 0 {
		return nil
	}
	schema, err := parseGormSchema(r.instance, dest)
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(dest).Elem()
	for col, val := range attrs {
		field, ok := schema.FieldsByDBName[col]
		if !ok {
			continue
		}
		if err := field.Set(r.ctx, rv, val); err != nil {
			return err
		}
	}
	return nil
}

// touchIfTouching bumps parent's updated_at column when desc.touches is true. Silently no-ops
// when desc.touches is false, when the parent's schema doesn't expose an updated_at field, or
// when the parent has no primary-key column to anchor the WHERE clause. Mirrors fedaco's
// touchIfTouching on BelongsToMany.
//
// Called at the tail end of public Sync / Attach / Detach / Toggle / UpdateExistingPivot methods,
// and only when the operation actually affected pivot rows. The internal doAttach / doDetach /
// doUpdateExistingPivot helpers do NOT touch — sync* paths chain multiple internal calls and
// touch at most once at the end via this helper.
func (r *Query) touchIfTouching(desc *relationDescriptor, parent any, parentVal any) error {
	if !desc.touches {
		return nil
	}
	parentSchema, err := parseGormSchema(r.instance, parent)
	if err != nil {
		return err
	}
	field, ok := parentSchema.FieldsByDBName["updated_at"]
	if !ok {
		// Parent doesn't have an updated_at column — silently skip.
		return nil
	}
	if len(parentSchema.PrimaryFields) == 0 {
		return nil
	}
	pkColumn := parentSchema.PrimaryFields[0].DBName
	now := time.Now()
	res := r.freshSession().Table(parentSchema.Table).
		Where(fmt.Sprintf("%s = ?", quoteIdent(pkColumn)), parentVal).
		Update(field.DBName, now)
	if res.Error != nil {
		return res.Error
	}
	// Mirror the change into the in-memory parent struct so subsequent reads see the bump.
	parentRV := reflect.ValueOf(parent).Elem()
	return field.Set(r.ctx, parentRV, now)
}

// kindName returns a human-friendly name for a relationKind, used in error messages.
func kindName(k relationKind) string {
	switch k {
	case relKindHasOne:
		return "hasOne"
	case relKindHasMany:
		return "hasMany"
	case relKindBelongsTo:
		return "belongsTo"
	case relKindMany2Many:
		return "many2Many"
	case relKindMorphOne:
		return "morphOne"
	case relKindMorphMany:
		return "morphMany"
	case relKindMorphTo:
		return "morphTo"
	case relKindMorphToMany:
		return "morphToMany"
	case relKindHasOneThrough:
		return "hasOneThrough"
	case relKindHasManyThrough:
		return "hasManyThrough"
	}
	return fmt.Sprintf("kind=%d", k)
}

// castKeys returns a copy of ids with each value normalised to keyType (the related model's PK
// type). Used by Sync* / Toggle* to ensure SyncResult elements carry a stable Go type regardless
// of what the caller passed in or what GORM scanned out of the pivot table.
//
// Mirrors fedaco's _castKeys / _getTypeSwapValue. Returns nil for a nil input slice (preserving
// the "no rows touched" signal).
func castKeys(ids []any, keyType reflect.Type) []any {
	if ids == nil {
		return nil
	}
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = castKey(id, keyType)
	}
	return out
}

// castKey converts v to the Go type t, handling the common cross-type cases (int/uint/float
// numeric widening + narrowing, string ↔ numeric). Returns v unchanged when t is nil, when v is
// already the right type, or when conversion isn't safely representable.
func castKey(v any, t reflect.Type) any {
	if v == nil || t == nil {
		return v
	}
	rv := reflect.ValueOf(v)
	if rv.Type() == t {
		return v
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return rv.Convert(t).Interface()
		case reflect.String:
			if i, err := strconv.ParseInt(rv.String(), 10, 64); err == nil {
				return reflect.ValueOf(i).Convert(t).Interface()
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return rv.Convert(t).Interface()
		case reflect.String:
			if u, err := strconv.ParseUint(rv.String(), 10, 64); err == nil {
				return reflect.ValueOf(u).Convert(t).Interface()
			}
		}
	case reflect.Float32, reflect.Float64:
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return rv.Convert(t).Interface()
		case reflect.String:
			if f, err := strconv.ParseFloat(rv.String(), 64); err == nil {
				return reflect.ValueOf(f).Convert(t).Interface()
			}
		}
	case reflect.String:
		// Numeric / []byte → string. Avoid reflect.Convert here because int→string interprets the
		// int as a Unicode code point (e.g. 65 → "A"), not a decimal digit string.
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Sprint(v)
		case reflect.Slice:
			if rv.Type().Elem().Kind() == reflect.Uint8 {
				return string(rv.Bytes())
			}
		}
	}
	return v
}
