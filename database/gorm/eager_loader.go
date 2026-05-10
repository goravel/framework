package gorm

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/samber/lo"
	gormio "gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"

	"github.com/goravel/framework/database/orm/morphmap"
	"github.com/goravel/framework/errors"
)

// defaultEagerLoadChunkSize is the WHERE IN list size at which we split a single eager-load
// query into multiple round-trips. 1000 covers the strictest mainstream limits:
// Oracle's hard cap of 1000 expressions and SQLite's default SQLITE_MAX_VARIABLE_NUMBER of 999.
// PostgreSQL/MySQL/SQL Server have higher limits but their planners also slow down dramatically
// past a few thousand entries, so chunking is a net win even where it isn't strictly required.
//
// The size can be overridden per-app via the `database.eager_load_chunk_size` config key. A value
// <= 0 disables chunking entirely (single IN clause regardless of length).
const defaultEagerLoadChunkSize = 1000

// applyEagerLoads runs all queued With entries against the just-loaded dest. It must be
// called by terminal methods (Get / Find / First / FirstOrFail / FirstOr / Cursor) after the main
// query has populated dest, and only when conditions.eagerLoad is non-empty.
func (r *Query) applyEagerLoads(dest any) error {
	if len(r.conditions.eagerLoad) == 0 {
		return nil
	}
	parents, err := collectEagerParents(dest)
	if err != nil {
		return err
	}
	entries := r.conditions.eagerLoad
	r.conditions.eagerLoad = nil
	if len(parents) == 0 {
		return nil
	}
	return r.runEagerLoads(parents, entries)
}

// runEagerLoads iterates the eager-load entries and dispatches each top-level relation to its
// kind-specific loader. Nested entries (those whose name contains a dot) are handled by the
// trickle-down recursion inside each loader, mirroring fedaco's eagerLoadRelations.
func (r *Query) runEagerLoads(parents []reflect.Value, entries []eagerLoadEntry) error {
	if len(parents) == 0 || len(entries) == 0 {
		return nil
	}
	parentModel := newSampleModel(parents[0])
	for _, entry := range entries {
		if strings.Contains(entry.relation, ".") {
			continue
		}
		nested := directNestedEntries(entries, entry.relation)
		if err := r.loadOneRelation(parents, parentModel, entry, nested); err != nil {
			return err
		}
	}
	return nil
}

func (r *Query) loadOneRelation(parents []reflect.Value, parentModel any, entry eagerLoadEntry, nested []eagerLoadEntry) error {
	desc, err := resolveRelation(r.instance, parentModel, entry.relation)
	if err != nil {
		return err
	}
	switch desc.kind {
	case relKindHasOne, relKindHasMany:
		return r.loadHasOneOrMany(parents, parentModel, desc, entry, nested, desc.kind == relKindHasMany)
	case relKindBelongsTo:
		return r.loadBelongsTo(parents, parentModel, desc, entry, nested)
	case relKindMany2Many:
		return r.loadMany2Many(parents, parentModel, desc, entry, nested)
	case relKindMorphOne, relKindMorphMany:
		return r.loadMorph(parents, parentModel, desc, entry, nested, desc.kind == relKindMorphMany)
	case relKindMorphTo:
		return r.loadMorphTo(parents, parentModel, desc, entry, nested)
	case relKindMorphToMany:
		return r.loadMorphToMany(parents, parentModel, desc, entry, nested)
	case relKindHasOneThrough, relKindHasManyThrough:
		return r.loadThrough(parents, parentModel, desc, entry, nested, desc.kind == relKindHasManyThrough)
	}
	return errors.OrmRelationUnsupported.Args(entry.relation, reflect.TypeOf(parentModel).String(), fmt.Sprintf("kind=%d", desc.kind))
}

// ---------------------------------------------------------------------------
// Per-kind loaders
// ---------------------------------------------------------------------------

func (r *Query) loadHasOneOrMany(parents []reflect.Value, parentModel any, desc *relationDescriptor, entry eagerLoadEntry, nested []eagerLoadEntry, isMany bool) error {
	if len(desc.references) == 0 {
		return errors.OrmRelationUnsupported.Args(entry.relation, "", "no references")
	}
	ref := desc.references[0]
	parentSchema, err := parseGormSchema(r.instance, parentModel)
	if err != nil {
		return err
	}
	parentField, ok := parentSchema.FieldsByDBName[ref.primaryColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, parentSchema.Name, "no parent field for "+ref.primaryColumn)
	}

	keys := extractKeys(r, parents, parentField)
	if len(keys) == 0 {
		return r.maybeRecurseEmpty(parents, entry.relation, isMany, nested)
	}

	rows, err := r.runChunkedRelatedQuery(keys, desc, entry, []string{ref.foreignColumn}, func(chunk []any) *gormio.DB {
		return r.freshSession().Table(desc.relatedTable).Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.relatedTable), quoteIdent(ref.foreignColumn)), chunk)
	})
	if err != nil {
		return err
	}

	relatedSchema, err := parseGormSchema(r.instance, desc.relatedModel)
	if err != nil {
		return err
	}
	fkField, ok := relatedSchema.FieldsByDBName[ref.foreignColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, relatedSchema.Name, "no related FK field for "+ref.foreignColumn)
	}

	dict := make(map[string][]reflect.Value, len(rows))
	for _, row := range rows {
		val, _ := fkField.ValueOf(r.ctx, row.Elem())
		dict[dictKey(val)] = append(dict[dictKey(val)], row)
	}

	if err := r.assignToParents(parents, parentField, entry.relation, dict, isMany); err != nil {
		return err
	}
	return r.recurseNested(rows, nested)
}

func (r *Query) loadBelongsTo(parents []reflect.Value, parentModel any, desc *relationDescriptor, entry eagerLoadEntry, nested []eagerLoadEntry) error {
	if len(desc.references) == 0 {
		return errors.OrmRelationUnsupported.Args(entry.relation, "", "no references")
	}
	ref := desc.references[0]
	parentSchema, err := parseGormSchema(r.instance, parentModel)
	if err != nil {
		return err
	}
	// For BelongsTo: ref.foreignTable=parent, ref.foreignColumn=FK on parent;
	// ref.primaryTable=related, ref.primaryColumn=PK on related.
	fkField, ok := parentSchema.FieldsByDBName[ref.foreignColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, parentSchema.Name, "no parent FK field for "+ref.foreignColumn)
	}

	keys := extractKeys(r, parents, fkField)
	if len(keys) == 0 {
		return r.maybeRecurseEmpty(parents, entry.relation, false, nested)
	}

	rows, err := r.runChunkedRelatedQuery(keys, desc, entry, []string{ref.primaryColumn}, func(chunk []any) *gormio.DB {
		return r.freshSession().Table(desc.relatedTable).Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.relatedTable), quoteIdent(ref.primaryColumn)), chunk)
	})
	if err != nil {
		return err
	}

	relatedSchema, err := parseGormSchema(r.instance, desc.relatedModel)
	if err != nil {
		return err
	}
	pkField, ok := relatedSchema.FieldsByDBName[ref.primaryColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, relatedSchema.Name, "no related PK field for "+ref.primaryColumn)
	}

	dict := make(map[string][]reflect.Value, len(rows))
	for _, row := range rows {
		val, _ := pkField.ValueOf(r.ctx, row.Elem())
		dict[dictKey(val)] = append(dict[dictKey(val)], row)
	}

	if err := r.assignToParents(parents, fkField, entry.relation, dict, false); err != nil {
		return err
	}
	return r.recurseNested(rows, nested)
}

func (r *Query) loadMorph(parents []reflect.Value, parentModel any, desc *relationDescriptor, entry eagerLoadEntry, nested []eagerLoadEntry, isMany bool) error {
	if len(desc.references) == 0 {
		return errors.OrmRelationUnsupported.Args(entry.relation, "", "no references")
	}
	ref := desc.references[0]
	parentSchema, err := parseGormSchema(r.instance, parentModel)
	if err != nil {
		return err
	}
	parentField, ok := parentSchema.FieldsByDBName[ref.primaryColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, parentSchema.Name, "no parent field for "+ref.primaryColumn)
	}

	keys := extractKeys(r, parents, parentField)
	if len(keys) == 0 {
		return r.maybeRecurseEmpty(parents, entry.relation, isMany, nested)
	}

	rows, err := r.runChunkedRelatedQuery(keys, desc, entry, []string{ref.foreignColumn, desc.morphTypeColumn}, func(chunk []any) *gormio.DB {
		return r.freshSession().
			Table(desc.relatedTable).
			Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.relatedTable), quoteIdent(ref.foreignColumn)), chunk).
			Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.relatedTable), quoteIdent(desc.morphTypeColumn)), desc.morphValue)
	})
	if err != nil {
		return err
	}

	relatedSchema, err := parseGormSchema(r.instance, desc.relatedModel)
	if err != nil {
		return err
	}
	fkField, ok := relatedSchema.FieldsByDBName[ref.foreignColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, relatedSchema.Name, "no related FK field for "+ref.foreignColumn)
	}

	dict := make(map[string][]reflect.Value, len(rows))
	for _, row := range rows {
		val, _ := fkField.ValueOf(r.ctx, row.Elem())
		dict[dictKey(val)] = append(dict[dictKey(val)], row)
	}

	if err := r.assignToParents(parents, parentField, entry.relation, dict, isMany); err != nil {
		return err
	}
	return r.recurseNested(rows, nested)
}

// loadMorphTo eager-loads the inverse polymorphic relation. Unlike the outbound MorphOne /
// MorphMany loaders, the related Go type is unknown at descriptor build time and discovered per
// row from the value of the *_type column. Parents are bucketed by morph type, and each bucket
// runs an IN query against its resolved table.
func (r *Query) loadMorphTo(parents []reflect.Value, parentModel any, desc *relationDescriptor, entry eagerLoadEntry, nested []eagerLoadEntry) error {
	parentSchema, err := parseGormSchema(r.instance, parentModel)
	if err != nil {
		return err
	}
	typeField, ok := parentSchema.FieldsByDBName[desc.morphTypeColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, parentSchema.Name, "no parent field for "+desc.morphTypeColumn)
	}
	idField, ok := parentSchema.FieldsByDBName[desc.morphIDColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, parentSchema.Name, "no parent field for "+desc.morphIDColumn)
	}

	// Bucket parents by morph type, deduplicating IDs per bucket.
	type bucket struct {
		keys     []any
		seenKeys map[string]struct{}
	}
	buckets := make(map[string]*bucket)
	parentMorphKey := make([]string, len(parents))  // morph type per parent
	parentBucketKey := make([]string, len(parents)) // dictKey of id per parent
	for i, parent := range parents {
		typeVal, typeZero := typeField.ValueOf(r.ctx, parent)
		idVal, idZero := idField.ValueOf(r.ctx, parent)
		if typeZero || idZero {
			continue
		}
		morphType, _ := typeVal.(string)
		if morphType == "" {
			morphType = fmt.Sprint(typeVal)
		}
		k := dictKey(idVal)
		b, exists := buckets[morphType]
		if !exists {
			b = &bucket{seenKeys: map[string]struct{}{}}
			buckets[morphType] = b
		}
		if _, dup := b.seenKeys[k]; !dup {
			b.seenKeys[k] = struct{}{}
			b.keys = append(b.keys, idVal)
		}
		parentMorphKey[i] = morphType
		parentBucketKey[i] = k
	}

	if len(buckets) == 0 {
		// No parents pointed at anything; clear the relation field on each.
		for _, parent := range parents {
			if err := setRelationField(parent, entry.relation, nil); err != nil {
				return err
			}
		}
		return nil
	}

	// For each bucket: resolve type, run IN query, build a per-bucket id->row dict.
	type resolvedBucket struct {
		dict map[string]reflect.Value // parent's dictKey(idVal) -> *RelatedModel
		rows []reflect.Value
	}
	resolved := make(map[string]resolvedBucket, len(buckets))
	allRows := make([]reflect.Value, 0)

	for morphType, b := range buckets {
		sample := morphmap.Find(morphType)
		if sample == nil {
			return errors.OrmMorphTypeUnknown.Args(morphType)
		}
		relatedTable, terr := tableNameFor(r.instance, sample)
		if terr != nil {
			return terr
		}
		// Make a per-bucket descriptor so runChunkedRelatedQuery's user-callback gets a
		// related-shaped query.
		bucketDesc := &relationDescriptor{
			kind:         relKindBelongsTo, // BelongsTo-shaped: WHERE related.<owner_key> IN ?
			parentTable:  desc.parentTable,
			relatedTable: relatedTable,
			relatedModel: sample,
			onQuery:      desc.onQuery, // propagate so the default scope applies per bucket
		}
		ownerKey := desc.morphOwnerKey
		if ownerKey == "" {
			ownerKey = "id"
		}

		rows, qerr := r.runChunkedRelatedQuery(b.keys, bucketDesc, entry, []string{ownerKey}, func(chunk []any) *gormio.DB {
			return r.freshSession().
				Table(relatedTable).
				Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(relatedTable), quoteIdent(ownerKey)), chunk)
		})
		if qerr != nil {
			return qerr
		}
		allRows = append(allRows, rows...)

		relatedSchema, perr := parseGormSchema(r.instance, sample)
		if perr != nil {
			return perr
		}
		ownerField, ok := relatedSchema.FieldsByDBName[ownerKey]
		if !ok {
			return errors.OrmRelationUnsupported.Args(entry.relation, relatedSchema.Name, "no owner key field for "+ownerKey)
		}
		dict := make(map[string]reflect.Value, len(rows))
		for _, row := range rows {
			val, _ := ownerField.ValueOf(r.ctx, row.Elem())
			dict[dictKey(val)] = row
		}
		resolved[morphType] = resolvedBucket{dict: dict, rows: rows}
	}

	// Assign each parent the row it pointed to.
	for i, parent := range parents {
		morphType := parentMorphKey[i]
		idKey := parentBucketKey[i]
		if morphType == "" || idKey == "" {
			if err := setRelationField(parent, entry.relation, nil); err != nil {
				return err
			}
			continue
		}
		bucketResult, ok := resolved[morphType]
		if !ok {
			if err := setRelationField(parent, entry.relation, nil); err != nil {
				return err
			}
			continue
		}
		row, ok := bucketResult.dict[idKey]
		if !ok {
			if err := setRelationField(parent, entry.relation, nil); err != nil {
				return err
			}
			continue
		}
		if err := setRelationField(parent, entry.relation, []reflect.Value{row}); err != nil {
			return err
		}
	}

	return r.recurseNested(allRows, nested)
}

// loadMany2Many eager-loads a regular many-to-many relation through a pivot table whose schema
// is described by GORM's parsed metadata.
func (r *Query) loadMany2Many(parents []reflect.Value, parentModel any, desc *relationDescriptor, entry eagerLoadEntry, nested []eagerLoadEntry) error {
	parentSchema, err := parseGormSchema(r.instance, parentModel)
	if err != nil {
		return err
	}
	parentField, ok := parentSchema.FieldsByDBName[desc.pivotParentRef.primaryColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, parentSchema.Name, "no parent field for "+desc.pivotParentRef.primaryColumn)
	}

	keys := extractKeys(r, parents, parentField)
	if len(keys) == 0 {
		return r.maybeRecurseEmpty(parents, entry.relation, true, nested)
	}

	pivotParentCol := desc.pivotParentRef.foreignColumn
	pivotRelatedCol := desc.pivotRelatedRef.foreignColumn

	// Check if the related model has a Pivot field — if yes, we'll SELECT extra pivot columns
	// using the desc.pivotUsing struct schema. struct-only: when desc.pivotUsing is nil, no pivot
	// hydration happens regardless of whether the related model has a Pivot field.
	relatedSchema, err := parseGormSchema(r.instance, desc.relatedModel)
	if err != nil {
		return err
	}
	pivotPlan, err := preparePivotHydration(r, desc)
	if err != nil {
		return err
	}

	// Build the pivot SELECT list: always include the two FK columns, plus the columns reported
	// by the Using-struct hydration plan when present.
	pivotSelectCols := []string{pivotParentCol, pivotRelatedCol}
	if pivotPlan != nil {
		pivotSelectCols = append(pivotSelectCols, pivotPlan.extraColumns...)
	}

	// Convert []string to []interface{} for GORM's Select signature.
	selectArgs := make([]interface{}, len(pivotSelectCols))
	for i, col := range pivotSelectCols {
		selectArgs[i] = col
	}

	pivotRows, err := r.chunkedFindMaps(keys, func(chunk []any) *gormio.DB {
		return r.freshSession().
			Table(desc.pivotTable).
			Select(selectArgs[0], selectArgs[1:]...).
			Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.pivotTable), quoteIdent(pivotParentCol)), chunk)
	})
	if err != nil {
		return err
	}
	if len(pivotRows) == 0 {
		return r.maybeRecurseEmpty(parents, entry.relation, true, nested)
	}

	relatedKeysSet := make(map[string]any, len(pivotRows))
	for _, p := range pivotRows {
		k := dictKey(p[pivotRelatedCol])
		if _, exists := relatedKeysSet[k]; !exists {
			relatedKeysSet[k] = p[pivotRelatedCol]
		}
	}
	relatedKeys := make([]any, 0, len(relatedKeysSet))
	for _, v := range relatedKeysSet {
		relatedKeys = append(relatedKeys, v)
	}

	rows, err := r.runChunkedRelatedQuery(relatedKeys, desc, entry, []string{desc.pivotRelatedRef.primaryColumn}, func(chunk []any) *gormio.DB {
		return r.freshSession().
			Table(desc.relatedTable).
			Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.relatedTable), quoteIdent(desc.pivotRelatedRef.primaryColumn)), chunk)
	})
	if err != nil {
		return err
	}

	relatedPKField, ok := relatedSchema.FieldsByDBName[desc.pivotRelatedRef.primaryColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, relatedSchema.Name, "no related PK field for "+desc.pivotRelatedRef.primaryColumn)
	}
	relatedByID := make(map[string]reflect.Value, len(rows))
	for _, row := range rows {
		val, _ := relatedPKField.ValueOf(r.ctx, row.Elem())
		relatedByID[dictKey(val)] = row
	}

	// Build pivot data map: key = relatedID, value = map of pivot column values to hydrate.
	var pivotDataByRelatedID map[string]map[string]any
	if pivotPlan != nil {
		pivotDataByRelatedID = make(map[string]map[string]any, len(pivotRows))
		for _, p := range pivotRows {
			relatedKey := dictKey(p[pivotRelatedCol])
			data := make(map[string]any, len(pivotPlan.extraColumns))
			for _, col := range pivotPlan.extraColumns {
				if val, ok := p[col]; ok {
					data[col] = val
				}
			}
			pivotDataByRelatedID[relatedKey] = data
		}
	}

	dict := make(map[string][]reflect.Value, len(parents))
	for _, p := range pivotRows {
		parentKey := dictKey(p[pivotParentCol])
		relatedKey := dictKey(p[pivotRelatedCol])
		if rel, ok := relatedByID[relatedKey]; ok {
			dict[parentKey] = append(dict[parentKey], rel)
		}
	}

	if err := r.assignToParents(parents, parentField, entry.relation, dict, true); err != nil {
		return err
	}

	// Hydrate Pivot field on each related row if we have pivot data.
	if pivotPlan != nil && len(pivotDataByRelatedID) > 0 {
		for _, row := range rows {
			val, _ := relatedPKField.ValueOf(r.ctx, row.Elem())
			relatedKey := dictKey(val)
			if data, ok := pivotDataByRelatedID[relatedKey]; ok {
				if err := writePivotField(r.ctx, row, data, pivotPlan); err != nil {
					return err
				}
			}
		}
	}

	return r.recurseNested(rows, nested)
}

// loadMorphToMany eager-loads polymorphic many-to-many. Mirrors loadMany2Many with one extra
// pivot WHERE that pins the morph_type column to desc.morphValue. Both MorphToMany (forward) and
// MorphedByMany (inverse) share this code path; the difference between them is captured in
// desc.morphValue at descriptor-build time.
func (r *Query) loadMorphToMany(parents []reflect.Value, parentModel any, desc *relationDescriptor, entry eagerLoadEntry, nested []eagerLoadEntry) error {
	parentSchema, err := parseGormSchema(r.instance, parentModel)
	if err != nil {
		return err
	}
	parentField, ok := parentSchema.FieldsByDBName[desc.pivotParentRef.primaryColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, parentSchema.Name, "no parent field for "+desc.pivotParentRef.primaryColumn)
	}

	keys := extractKeys(r, parents, parentField)
	if len(keys) == 0 {
		return r.maybeRecurseEmpty(parents, entry.relation, true, nested)
	}

	pivotParentCol := desc.pivotParentRef.foreignColumn
	pivotRelatedCol := desc.pivotRelatedRef.foreignColumn

	// Check if the related model has a Pivot field — if yes, we'll SELECT extra pivot columns
	// using the desc.pivotUsing struct schema. struct-only: when desc.pivotUsing is nil, no pivot
	// hydration happens regardless of whether the related model has a Pivot field.
	relatedSchema, err := parseGormSchema(r.instance, desc.relatedModel)
	if err != nil {
		return err
	}
	pivotPlan, err := preparePivotHydration(r, desc)
	if err != nil {
		return err
	}

	// Build the pivot SELECT list: always include the two FK columns, plus the columns reported
	// by the Using-struct hydration plan when present.
	pivotSelectCols := []string{pivotParentCol, pivotRelatedCol}
	if pivotPlan != nil {
		pivotSelectCols = append(pivotSelectCols, pivotPlan.extraColumns...)
	}

	// Convert []string to []interface{} for GORM's Select signature.
	selectArgs := make([]interface{}, len(pivotSelectCols))
	for i, col := range pivotSelectCols {
		selectArgs[i] = col
	}

	pivotRows, err := r.chunkedFindMaps(keys, func(chunk []any) *gormio.DB {
		return r.freshSession().
			Table(desc.pivotTable).
			Select(selectArgs[0], selectArgs[1:]...).
			Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.pivotTable), quoteIdent(pivotParentCol)), chunk).
			Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.morphTypeColumn)), desc.morphValue)
	})
	if err != nil {
		return err
	}
	if len(pivotRows) == 0 {
		return r.maybeRecurseEmpty(parents, entry.relation, true, nested)
	}

	relatedKeysSet := make(map[string]any, len(pivotRows))
	for _, p := range pivotRows {
		k := dictKey(p[pivotRelatedCol])
		if _, exists := relatedKeysSet[k]; !exists {
			relatedKeysSet[k] = p[pivotRelatedCol]
		}
	}
	relatedKeys := make([]any, 0, len(relatedKeysSet))
	for _, v := range relatedKeysSet {
		relatedKeys = append(relatedKeys, v)
	}

	rows, err := r.runChunkedRelatedQuery(relatedKeys, desc, entry, []string{desc.pivotRelatedRef.primaryColumn}, func(chunk []any) *gormio.DB {
		return r.freshSession().
			Table(desc.relatedTable).
			Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.relatedTable), quoteIdent(desc.pivotRelatedRef.primaryColumn)), chunk)
	})
	if err != nil {
		return err
	}

	relatedPKField, ok := relatedSchema.FieldsByDBName[desc.pivotRelatedRef.primaryColumn]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, relatedSchema.Name, "no related PK field for "+desc.pivotRelatedRef.primaryColumn)
	}
	relatedByID := make(map[string]reflect.Value, len(rows))
	for _, row := range rows {
		val, _ := relatedPKField.ValueOf(r.ctx, row.Elem())
		relatedByID[dictKey(val)] = row
	}

	// Build pivot data map: key = relatedID, value = map of pivot column values to hydrate.
	var pivotDataByRelatedID map[string]map[string]any
	if pivotPlan != nil {
		pivotDataByRelatedID = make(map[string]map[string]any, len(pivotRows))
		for _, p := range pivotRows {
			relatedKey := dictKey(p[pivotRelatedCol])
			data := make(map[string]any, len(pivotPlan.extraColumns))
			for _, col := range pivotPlan.extraColumns {
				if val, ok := p[col]; ok {
					data[col] = val
				}
			}
			pivotDataByRelatedID[relatedKey] = data
		}
	}

	dict := make(map[string][]reflect.Value, len(parents))
	for _, p := range pivotRows {
		parentKey := dictKey(p[pivotParentCol])
		relatedKey := dictKey(p[pivotRelatedCol])
		if rel, ok := relatedByID[relatedKey]; ok {
			dict[parentKey] = append(dict[parentKey], rel)
		}
	}

	if err := r.assignToParents(parents, parentField, entry.relation, dict, true); err != nil {
		return err
	}

	// Hydrate Pivot field on each related row if we have pivot data.
	if pivotPlan != nil && len(pivotDataByRelatedID) > 0 {
		for _, row := range rows {
			val, _ := relatedPKField.ValueOf(r.ctx, row.Elem())
			relatedKey := dictKey(val)
			if data, ok := pivotDataByRelatedID[relatedKey]; ok {
				if err := writePivotField(r.ctx, row, data, pivotPlan); err != nil {
					return err
				}
			}
		}
	}

	return r.recurseNested(rows, nested)
}

func (r *Query) loadThrough(parents []reflect.Value, parentModel any, desc *relationDescriptor, entry eagerLoadEntry, nested []eagerLoadEntry, isMany bool) error {
	parentSchema, err := parseGormSchema(r.instance, parentModel)
	if err != nil {
		return err
	}
	parentField, ok := parentSchema.FieldsByDBName[desc.localKey]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, parentSchema.Name, "no parent field for "+desc.localKey)
	}

	keys := extractKeys(r, parents, parentField)
	if len(keys) == 0 {
		return r.maybeRecurseEmpty(parents, entry.relation, isMany, nested)
	}

	throughRows, err := r.chunkedFindMaps(keys, func(chunk []any) *gormio.DB {
		return r.freshSession().
			Table(desc.throughTable).
			Select(desc.firstKey, desc.secondLocalKey).
			Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.throughTable), quoteIdent(desc.firstKey)), chunk)
	})
	if err != nil {
		return err
	}
	if len(throughRows) == 0 {
		return r.maybeRecurseEmpty(parents, entry.relation, isMany, nested)
	}

	secondKeysSet := make(map[string]any, len(throughRows))
	for _, t := range throughRows {
		k := dictKey(t[desc.secondLocalKey])
		if _, exists := secondKeysSet[k]; !exists {
			secondKeysSet[k] = t[desc.secondLocalKey]
		}
	}
	secondKeys := make([]any, 0, len(secondKeysSet))
	for _, v := range secondKeysSet {
		secondKeys = append(secondKeys, v)
	}

	rows, err := r.runChunkedRelatedQuery(secondKeys, desc, entry, []string{desc.secondKey}, func(chunk []any) *gormio.DB {
		return r.freshSession().
			Table(desc.relatedTable).
			Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.relatedTable), quoteIdent(desc.secondKey)), chunk)
	})
	if err != nil {
		return err
	}

	relatedSchema, err := parseGormSchema(r.instance, desc.relatedModel)
	if err != nil {
		return err
	}
	secondField, ok := relatedSchema.FieldsByDBName[desc.secondKey]
	if !ok {
		return errors.OrmRelationUnsupported.Args(entry.relation, relatedSchema.Name, "no related field for "+desc.secondKey)
	}
	relatedByThrough := make(map[string][]reflect.Value, len(rows))
	for _, row := range rows {
		val, _ := secondField.ValueOf(r.ctx, row.Elem())
		k := dictKey(val)
		relatedByThrough[k] = append(relatedByThrough[k], row)
	}

	dict := make(map[string][]reflect.Value, len(parents))
	for _, t := range throughRows {
		parentKey := dictKey(t[desc.firstKey])
		secondKey := dictKey(t[desc.secondLocalKey])
		if rels, ok := relatedByThrough[secondKey]; ok {
			dict[parentKey] = append(dict[parentKey], rels...)
		}
	}

	if err := r.assignToParents(parents, parentField, entry.relation, dict, isMany); err != nil {
		return err
	}
	return r.recurseNested(rows, nested)
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// runRelatedQuery applies the user's callback and column pruning to the inner builder, executes
// it, and returns the result rows as []reflect.Value where each value is a *RelatedModel.
//
// requiredCols are columns the loader needs back (FK columns, PK columns) to build dictionaries;
// they are appended to the user's prune list when not already present so the caller does not have
// to remember to include them.
func (r *Query) runRelatedQuery(inner *gormio.DB, desc *relationDescriptor, entry eagerLoadEntry, requiredCols []string) ([]reflect.Value, error) {
	// Apply the per-relation default scope first so callers can layer extra constraints on top
	// via With("Books", func(q) { ... }).
	if desc.onQuery != nil {
		wrapper := r.wrap(inner)
		wrapped := desc.onQuery(wrapper)
		if w, ok := wrapped.(*Query); ok {
			inner = w.buildConditions().instance
		}
	}
	var oneOfMany *oneOfManyConfig
	if entry.callback != nil {
		wrapper := r.wrap(inner)
		wrapped := entry.callback(wrapper)
		if w, ok := wrapped.(*Query); ok {
			oneOfMany = w.conditions.oneOfMany
			inner = w.buildConditions().instance
		}
	}
	if oneOfMany != nil {
		inner = r.applyOneOfManyJoin(inner, desc, oneOfMany)
	}
	if len(entry.columns) > 0 {
		cols := append([]string(nil), entry.columns...)
		for _, req := range requiredCols {
			if !slices.ContainsFunc(cols, func(c string) bool {
				if c == req {
					return true
				}
				_, suffix, ok := strings.Cut(c, ".")
				return ok && suffix == req
			}) {
				cols = append(cols, req)
			}
		}
		inner = inner.Select(cols)
	}

	relatedType := reflect.TypeOf(desc.relatedModel)
	if relatedType.Kind() == reflect.Pointer {
		relatedType = relatedType.Elem()
	}
	sliceType := reflect.SliceOf(reflect.PointerTo(relatedType))
	slicePtr := reflect.New(sliceType)
	if err := inner.Find(slicePtr.Interface()).Error; err != nil {
		return nil, err
	}
	slice := slicePtr.Elem()
	out := make([]reflect.Value, 0, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		out = append(out, slice.Index(i))
	}
	return out, nil
}

// runChunkedRelatedQuery runs runRelatedQuery once per chunk of keys and concatenates rows. Each
// chunk gets a freshly built inner query from buildInner so the user's callback / column pruning
// is applied per-chunk.
//
// Note: when entry.callback installs a LIMIT, that LIMIT applies *per chunk*, not globally —
// same semantics as Eloquent's chunkById iteration and unavoidable for any chunked IN approach.
func (r *Query) runChunkedRelatedQuery(keys []any, desc *relationDescriptor, entry eagerLoadEntry, requiredCols []string, buildInner func(chunk []any) *gormio.DB) ([]reflect.Value, error) {
	chunks := chunkKeys(keys, r.chunkSize())
	var all []reflect.Value
	for _, chunk := range chunks {
		rows, err := r.runRelatedQuery(buildInner(chunk), desc, entry, requiredCols)
		if err != nil {
			return nil, err
		}
		all = append(all, rows...)
	}
	return all, nil
}

// chunkedFindMaps runs the pivot / through intermediate query in chunks of keys and accumulates
// results into a single []map[string]any. Used by loadMany2Many and loadThrough for the lookup
// queries that don't go through runRelatedQuery.
func (r *Query) chunkedFindMaps(keys []any, buildQuery func(chunk []any) *gormio.DB) ([]map[string]any, error) {
	chunks := chunkKeys(keys, r.chunkSize())
	var all []map[string]any
	for _, chunk := range chunks {
		var rows []map[string]any
		if err := buildQuery(chunk).Find(&rows).Error; err != nil {
			return nil, err
		}
		all = append(all, rows...)
	}
	return all, nil
}

// assignToParents writes the dictionary entries onto each parent's relation field using
// setRelationField. When isMany is false and a parent has multiple matches, only the first one is
// assigned (HasOne / BelongsTo / MorphOne / HasOneThrough cases).
func (r *Query) assignToParents(parents []reflect.Value, parentField *gormschema.Field, relation string, dict map[string][]reflect.Value, isMany bool) error {
	for _, parent := range parents {
		val, zero := parentField.ValueOf(r.ctx, parent)
		if zero {
			if !isMany {
				continue
			}
			if err := setRelationField(parent, relation, nil); err != nil {
				return err
			}
			continue
		}
		match := dict[dictKey(val)]
		if !isMany && len(match) > 1 {
			match = match[:1]
		}
		if err := setRelationField(parent, relation, match); err != nil {
			return err
		}
	}
	return nil
}

func (r *Query) recurseNested(rows []reflect.Value, nested []eagerLoadEntry) error {
	if len(rows) == 0 || len(nested) == 0 {
		return nil
	}
	nestedParents := make([]reflect.Value, 0, len(rows))
	for _, row := range rows {
		nestedParents = append(nestedParents, row.Elem())
	}
	return r.runEagerLoads(nestedParents, nested)
}

// maybeRecurseEmpty is the no-op fast path: when there are no parent keys to load against, leave
// each parent's relation field at its zero value (or empty slice for many) and skip nested.
func (r *Query) maybeRecurseEmpty(parents []reflect.Value, relation string, isMany bool, _ []eagerLoadEntry) error {
	if !isMany {
		return nil
	}
	for _, parent := range parents {
		if err := setRelationField(parent, relation, nil); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Reflect / extraction helpers
// ---------------------------------------------------------------------------

// collectEagerParents extracts the addressable struct values from dest. dest may be *Struct,
// *[]Struct, or *[]*Struct; each form yields a flat slice of struct values whose fields can be
// mutated.
func collectEagerParents(dest any) ([]reflect.Value, error) {
	if dest == nil {
		return nil, nil
	}
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return nil, nil
	}
	elem := rv.Elem()
	switch elem.Kind() {
	case reflect.Struct:
		return []reflect.Value{elem}, nil
	case reflect.Slice:
		out := make([]reflect.Value, 0, elem.Len())
		for i := 0; i < elem.Len(); i++ {
			item := elem.Index(i)
			if item.Kind() == reflect.Pointer {
				if item.IsNil() {
					continue
				}
				item = item.Elem()
			}
			if item.Kind() != reflect.Struct {
				continue
			}
			out = append(out, item)
		}
		return out, nil
	}
	return nil, nil
}

// newSampleModel returns a fresh pointer-to-struct of the same type as parent. resolveRelation
// expects an addressable model instance (it parses the schema and inspects its tags), and we
// don't want to hand it one of our actual parent rows (which may carry mutated fields).
func newSampleModel(parent reflect.Value) any {
	t := parent.Type()
	return reflect.New(t).Interface()
}

func parseGormSchema(db *gormio.DB, model any) (*gormschema.Schema, error) {
	stmt := &gormio.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return nil, err
	}
	return stmt.Schema, nil
}

// extractKeys pulls the unique non-zero values of field from the parent slice.
func extractKeys(r *Query, parents []reflect.Value, field *gormschema.Field) []any {
	seen := make(map[string]struct{}, len(parents))
	out := make([]any, 0, len(parents))
	for _, parent := range parents {
		val, zero := field.ValueOf(r.ctx, parent)
		if zero {
			continue
		}
		k := dictKey(val)
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, val)
	}
	return out
}

// dictKey reduces any value to a canonical string for use as a map key, paving over the type
// mismatch between Go field types (uint, int64, string) and database-layer scan types
// (often int64 or []byte). Mirrors fedaco's _getDictionaryKey.
func dictKey(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(x)
	case string:
		return x
	}
	return fmt.Sprint(v)
}

// chunkSize returns the eager-load IN-clause chunk size, falling back to the default when the
// config value is unset or invalid. A non-positive value disables chunking.
func (r *Query) chunkSize() int {
	if r.config == nil {
		return defaultEagerLoadChunkSize
	}
	v := r.config.GetInt("database.eager_load_chunk_size", defaultEagerLoadChunkSize)
	if v == 0 {
		return defaultEagerLoadChunkSize
	}
	return v
}

// chunkKeys splits keys into batches of at most size. Returns the input unchanged in a single
// batch when size <= 0 or len(keys) <= size, which lets callers stay on the cheap single-query
// path for typical workloads.
func chunkKeys(keys []any, size int) [][]any {
	if size <= 0 || len(keys) <= size {
		return [][]any{keys}
	}
	return lo.Chunk(keys, size)
}

// setRelationField writes loaded rows back to parent's relation field. Supports *Model,
// []*Model, []Model, and `any` (interface) field shapes — the last is used by MorphTo, where the
// concrete loaded type is determined per row from the morph map.
func setRelationField(parent reflect.Value, fieldName string, rows []reflect.Value) error {
	field := parent.FieldByName(fieldName)
	if !field.IsValid() {
		return errors.OrmEagerLoadCannotAssign.Args(fieldName, parent.Type().String())
	}
	if !field.CanSet() {
		return errors.OrmEagerLoadCannotAssign.Args(fieldName, parent.Type().String())
	}

	switch field.Kind() {
	case reflect.Interface:
		if len(rows) == 0 {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		row := rows[0]
		if row.Type().Implements(field.Type()) {
			field.Set(row)
			return nil
		}
		// `any` (empty interface) — anything implements it.
		if field.Type().NumMethod() == 0 {
			field.Set(row)
			return nil
		}
		return errors.OrmEagerLoadCannotAssign.Args(fieldName, parent.Type().String())

	case reflect.Pointer:
		if len(rows) == 0 {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		row := rows[0]
		if row.Type() == field.Type() {
			field.Set(row)
			return nil
		}
		if row.Kind() == reflect.Pointer && row.Type().Elem() == field.Type().Elem() {
			field.Set(row)
			return nil
		}
		return errors.OrmEagerLoadCannotAssign.Args(fieldName, parent.Type().String())

	case reflect.Slice:
		elemType := field.Type().Elem()
		out := reflect.MakeSlice(field.Type(), 0, len(rows))
		for _, row := range rows {
			switch elemType.Kind() {
			case reflect.Pointer:
				if row.Type() == elemType {
					out = reflect.Append(out, row)
					continue
				}
				if row.Kind() == reflect.Pointer && row.Type().Elem() == elemType.Elem() {
					out = reflect.Append(out, row)
					continue
				}
				return errors.OrmEagerLoadCannotAssign.Args(fieldName, parent.Type().String())
			case reflect.Struct:
				if row.Kind() == reflect.Pointer && row.Type().Elem() == elemType {
					out = reflect.Append(out, row.Elem())
					continue
				}
				if row.Type() == elemType {
					out = reflect.Append(out, row)
					continue
				}
				return errors.OrmEagerLoadCannotAssign.Args(fieldName, parent.Type().String())
			default:
				return errors.OrmEagerLoadCannotAssign.Args(fieldName, parent.Type().String())
			}
		}
		field.Set(out)
		return nil
	}
	return errors.OrmEagerLoadCannotAssign.Args(fieldName, parent.Type().String())
}

// pivotHydrationPlan precomputes everything writePivotField needs to copy a row of pivot column
// values into the configured pivot field on each eager-loaded related model. It is built once
// per loadMany2Many / loadMorphToMany invocation by preparePivotHydration. nil means "no Pivot
// hydration" (related model has no field by desc.pivotField).
type pivotHydrationPlan struct {
	// fieldName is the struct field on the related model that we hydrate (typically "Pivot").
	fieldName string
	// extraColumns is the pivot SELECT list contributed by the pivot struct (every db-tagged
	// field's column name), not including the two FK columns which loadMany2Many always selects.
	extraColumns []string
	// fieldByColumn maps each db column name to the *gormschema.Field that owns it on the pivot
	// struct. Used by writePivotField to set struct fields from the SELECT row.
	fieldByColumn map[string]*gormschema.Field
}

// preparePivotHydration inspects the related model for a field named desc.pivotField. When found,
// returns a plan that drives the pivot SELECT list and field-by-field hydration. Returns nil (no
// error) when the related model has no field by that name — pivot data still flows through the
// SELECT but nothing is surfaced. Returns an error when the field exists but isn't a struct
// (catches typos like `Pivot string`).
func preparePivotHydration(r *Query, desc *relationDescriptor) (*pivotHydrationPlan, error) {
	relatedType := reflect.TypeOf(desc.relatedModel)
	if relatedType.Kind() == reflect.Pointer {
		relatedType = relatedType.Elem()
	}
	if relatedType.Kind() != reflect.Struct {
		return nil, nil
	}
	pivotStructField, ok := relatedType.FieldByName(desc.pivotField)
	if !ok {
		// No field by this name — silently skip hydration; the relation still works for joining.
		return nil, nil
	}
	if pivotStructField.Type.Kind() != reflect.Struct {
		return nil, errors.OrmRelationPivotFieldNotStruct.Args(
			relatedType.String(), desc.pivotField, pivotStructField.Type.Kind().String(),
		)
	}
	// Parse the field type's GORM schema by instantiating a zero value of it.
	usingSchema, err := parseGormSchema(r.instance, reflect.New(pivotStructField.Type).Interface())
	if err != nil {
		return nil, err
	}
	cols := make([]string, 0, len(usingSchema.Fields))
	byCol := make(map[string]*gormschema.Field, len(usingSchema.Fields))
	for _, f := range usingSchema.Fields {
		if f.DBName == "" {
			continue
		}
		cols = append(cols, f.DBName)
		byCol[f.DBName] = f
	}
	return &pivotHydrationPlan{
		fieldName:     desc.pivotField,
		extraColumns:  cols,
		fieldByColumn: byCol,
	}, nil
}

// writePivotField copies the column values in data into rv's pivot struct field (named
// plan.fieldName), using plan.fieldByColumn to resolve column names to *gormschema.Fields on the
// pivot struct. Caller guarantees plan is non-nil.
func writePivotField(ctx context.Context, rv reflect.Value, data map[string]any, plan *pivotHydrationPlan) error {
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	pivotField := rv.FieldByName(plan.fieldName)
	if !pivotField.IsValid() || !pivotField.CanSet() {
		return nil
	}
	for col, val := range data {
		f, ok := plan.fieldByColumn[col]
		if !ok {
			continue
		}
		if err := f.Set(ctx, pivotField, val); err != nil {
			return err
		}
	}
	return nil
}
