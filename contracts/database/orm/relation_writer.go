package orm

import (
	"github.com/goravel/framework/contracts/database/db"
)

// RelationWriter is the write-side builder for a single (parent, relation) pair, returned by
// Query.Relation / Orm.Relation. All write operations on a relation flow through this interface;
// flat methods like Orm.Save(parent, relation, child) intentionally do not exist.
//
// RelationWriter has NO Where / OrderBy / chain methods — fedaco-style "where(...).updateOrCreate(...)"
// composition is not supported because Goravel's relation system is metadata-driven (declared via
// ModelWithRelations.Relations()), not method-driven. Search criteria for the find-then-write
// methods (FirstOrNew/FirstOrCreate/UpdateOrCreate/FindOrNew) are passed via the attrs map, which
// is combined with the relation's foreign-key scope.
//
// For chained reads on a relation, use Query.Related(parent, name) instead — it returns a regular
// Query already scoped by the relation's foreign key, supporting the full Where/OrderBy/Get chain.
type RelationWriter interface {
	// Save inserts or updates child as a member of the relation. Sets child's foreign key (and
	// morph_type for MorphOne/MorphMany) from parent's local key, then persists child.
	// Supported kinds: HasOne, HasMany, MorphOne, MorphMany, Many2Many, MorphToMany.
	Save(child any) error
	// SaveMany is the slice form of Save. children must be a slice or pointer-to-slice.
	SaveMany(children any) error
	// SaveWithPivot is Save with caller-supplied pivot column values for BelongsToMany kinds.
	// On HasOneOrMany kinds attrs is ignored (no pivot row).
	SaveWithPivot(child any, attrs map[string]any) error
	// SaveManyWithPivot is the slice form of SaveWithPivot. attrsPerChild is keyed by the related
	// PK of each child; an entry may be nil to attach without extra columns.
	SaveManyWithPivot(children any, attrsPerChild map[any]map[string]any) error

	// Create persists a new related row. For HasOneOrMany kinds the framework pre-sets FK (and
	// morph type) on dest from parent, then inserts. For BelongsToMany kinds inserts dest first,
	// then writes a pivot row.
	Create(dest any) error
	// CreateMany is the slice form of Create.
	CreateMany(dests any) error

	// FindOrNew finds the related row with primary key id. If absent, fills dest with a new
	// instance of the related model and pre-sets FK (and morph type) — does NOT persist.
	FindOrNew(id any, dest any) error
	// FirstOrNew finds the first related row matching attrs. If absent, fills dest with a new
	// instance carrying attrs+values and pre-set FK — does NOT persist.
	FirstOrNew(attrs, values map[string]any, dest any) error
	// FirstOrCreate is FirstOrNew that persists when no matching row exists. For BelongsToMany
	// kinds also writes a pivot row.
	FirstOrCreate(attrs, values map[string]any, dest any) error
	// UpdateOrCreate finds the first related row matching attrs (or creates one), then overlays
	// values onto it and persists. For BelongsToMany kinds also writes a pivot row when freshly
	// created. Always saves dest.
	UpdateOrCreate(attrs, values map[string]any, dest any) error

	// Associate sets parent's foreign key (and morph_type for MorphTo) to point at owner, then
	// persists parent. Supported kinds: BelongsTo, MorphTo. owner must be a non-nil pointer to a
	// struct.
	Associate(owner any) error
	// Dissociate clears parent's foreign key (and morph_type for MorphTo) and persists parent.
	// Supported kinds: BelongsTo, MorphTo.
	Dissociate() error

	// Attach inserts pivot rows linking parent to each id in ids. For polymorphic pivots the
	// morph_type column is filled from the parent's morph value. Skips ids that already have a
	// pivot row. Supported kinds: Many2Many, MorphToMany, MorphedByMany.
	Attach(ids []any) error
	// AttachWithPivot is Attach with per-row pivot column values. The map key is the related id;
	// the map value is the column-name-to-value map applied to that pivot row.
	AttachWithPivot(idsWithAttrs map[any]map[string]any) error
	// Detach removes pivot rows linking parent to the given ids. With nil ids, removes all pivot
	// rows for parent (and morph type, for polymorphic). Returns the number of rows removed.
	Detach(ids ...any) (int64, error)

	// Sync replaces parent's pivot rows so they exactly match ids: detaches missing entries,
	// attaches new ones, leaves existing untouched.
	Sync(ids []any) (*db.SyncResult, error)
	// SyncWithPivot is Sync with per-ID pivot column values. The map key is the related id; the
	// map value is the column-name-to-value map applied to that pivot row. For existing pivot
	// rows with non-empty attrs, updates the pivot columns (reported in SyncResult.Updated).
	SyncWithPivot(idsWithAttrs map[any]map[string]any) (*db.SyncResult, error)
	// SyncWithPivotValues is a convenience wrapper that applies the same pivot column values to
	// all ids.
	SyncWithPivotValues(ids []any, pivotValues map[string]any) (*db.SyncResult, error)
	// SyncWithoutDetaching is Sync minus the detach step — adds missing entries only.
	SyncWithoutDetaching(ids []any) (*db.SyncResult, error)
	// SyncWithoutDetachingWithPivot is SyncWithPivot minus the detach step.
	SyncWithoutDetachingWithPivot(idsWithAttrs map[any]map[string]any) (*db.SyncResult, error)
	// Toggle attaches missing entries and detaches existing ones.
	Toggle(ids []any) (*db.SyncResult, error)
	// ToggleWithPivot is Toggle with per-ID pivot column values for newly attached rows.
	ToggleWithPivot(idsWithAttrs map[any]map[string]any) (*db.SyncResult, error)

	// UpdateExistingPivot updates pivot columns for an already-attached id.
	UpdateExistingPivot(id any, attrs map[string]any) (int64, error)
}
