package orm

// RelationCallback is the signature of a closure used to scope a relationship existence query.
// Mirrors the (q) => void closure shape from the QueriesRelationships mixin. The returned Query
// is the inner subquery used for has / whereHas / withCount / etc.
type RelationCallback func(query Query) Query

// MorphRelationCallback is the per-type variant of RelationCallback used by the *HasMorph family,
// matching the `function ($query, $type)` callback for whereHasMorph. The second argument is the
// morph type currently being scoped (the related model's morph class - the table name in GORM's
// polymorphic convention).
type MorphRelationCallback func(query Query, morphType string) Query

// PivotQuery is a small builder surface for scoping the pivot-table SQL emitted by Sync / Attach
// duplicate-skip / Detach / UpdateExistingPivot. It is intentionally narrower than the full Query
// interface — only the WHERE-style methods that make sense on a single-table pivot read/write are
// exposed. Mirrors fedaco's `wherePivot` / `wherePivotIn` / `wherePivotNull` accumulators on
// BelongsToMany.
//
// NOTE: PivotQuery filters only SELECT / UPDATE / DELETE on the pivot table — equality conditions
// added here are NOT auto-injected into INSERT rows on Attach. Callers that need the conditions
// to appear on inserted rows should pass them through Attach attrs.
type PivotQuery interface {
	// Where adds a `column op value` clause to the pivot query. operator defaults to "=" when
	// only two args are passed (column, value).
	Where(column string, args ...any) PivotQuery
	// WhereIn adds a `column IN (...)` clause to the pivot query.
	WhereIn(column string, values []any) PivotQuery
	// WhereNotIn adds a `column NOT IN (...)` clause to the pivot query.
	WhereNotIn(column string, values []any) PivotQuery
	// WhereNull adds a `column IS NULL` clause to the pivot query.
	WhereNull(column string) PivotQuery
	// WhereNotNull adds a `column IS NOT NULL` clause to the pivot query.
	WhereNotNull(column string) PivotQuery
}

// PivotCallback scopes pivot-table reads (and the corresponding diff-driven writes) for a
// BelongsToMany relation. Declared via Many2Many / MorphToMany / MorphedByMany's OnPivotQuery
// field; applied automatically to Sync / Detach / UpdateExistingPivot and Attach's duplicate-
// detection SELECT.
type PivotCallback func(query PivotQuery) PivotQuery

// QueryWithRelations is the Go port of the QueriesRelationships mixin from Laravel and the 1:1
// TypeScript port in fedaco at libs/fedaco/src/fedaco/mixins/queries-relationships.ts.
//
// Where the upstream framework has first-class Relation objects with getRelationExistenceQuery
// methods, GORM models its relationships through struct-tag metadata. The bridge here surfaces
// relationship-existence and aggregate-subselect queries on top of that metadata. Callers can
// write, for example:
//
//	users := []User{}
//	query.Query().Has("Books", ">=", 3).WithCount("Roles").Get(&users)
//
// Existence-style methods (Has / OrHas / DoesntHave / WhereHas / ...) accept a variadic args
// slice that may carry, in any order:
//   - a RelationCallback or func(Query) Query to scope the inner subquery
//   - a string operator (e.g. ">=", "<", ">", "=") - defaults to ">="
//   - an int count - defaults to 1
//
// Morph-style methods take an additional types []any of model instances; the morph value used in
// the type column is derived from each model's GORM-resolved table name (e.g. *User -> "users").
//
// QueryWithRelations is embedded into Query, so all of these methods are also reachable directly
// off Query without a type assertion.
type QueryWithRelations interface {
	// Has adds a relationship count / exists condition to the query.
	// Defaults to operator ">=" and count 1.
	Has(relation string, args ...any) Query
	// OrHas adds a relationship count / exists condition to the query with an "or" conjunction.
	OrHas(relation string, args ...any) Query
	// DoesntHave adds a relationship absence condition - equivalent to Has(rel, "<", 1).
	DoesntHave(relation string, args ...any) Query
	// OrDoesntHave adds a relationship absence condition with an "or" conjunction.
	OrDoesntHave(relation string, args ...any) Query
	// WhereHas adds a relationship count / exists condition to the query with where clauses.
	// Identical semantics to Has but conventionally used with a callback first arg.
	WhereHas(relation string, args ...any) Query
	// OrWhereHas adds a relationship count / exists condition to the query with where clauses
	// and an "or" conjunction.
	OrWhereHas(relation string, args ...any) Query
	// WhereDoesntHave adds a relationship absence condition to the query with where clauses.
	WhereDoesntHave(relation string, args ...any) Query
	// OrWhereDoesntHave adds a relationship absence condition to the query with where clauses
	// and an "or" conjunction.
	OrWhereDoesntHave(relation string, args ...any) Query

	// HasMorph adds a polymorphic relationship count / exists condition to the query.
	// types is a slice of model instances (e.g. []any{&Post{}, &Video{}}); the morph value
	// used in the type column is derived from each model's table name.
	//
	// Note: auto-discovery of distinct morph values via `types = ['*']` is not supported.
	// An explicit list of model instances is required.
	HasMorph(relation string, types []any, args ...any) Query
	// OrHasMorph adds a polymorphic relationship count / exists condition with an "or" conjunction.
	OrHasMorph(relation string, types []any, args ...any) Query
	// DoesntHaveMorph adds a polymorphic relationship absence condition.
	DoesntHaveMorph(relation string, types []any, args ...any) Query
	// OrDoesntHaveMorph adds a polymorphic relationship absence condition with an "or" conjunction.
	OrDoesntHaveMorph(relation string, types []any, args ...any) Query
	// WhereHasMorph adds a polymorphic relationship count / exists condition to the query with
	// where clauses. Callbacks may be MorphRelationCallback for per-type scoping.
	WhereHasMorph(relation string, types []any, args ...any) Query
	// OrWhereHasMorph adds a polymorphic relationship count / exists condition with where clauses
	// and an "or" conjunction.
	OrWhereHasMorph(relation string, types []any, args ...any) Query
	// WhereDoesntHaveMorph adds a polymorphic relationship absence condition with where clauses.
	WhereDoesntHaveMorph(relation string, types []any, args ...any) Query
	// OrWhereDoesntHaveMorph adds a polymorphic relationship absence condition with where clauses
	// and an "or" conjunction.
	OrWhereDoesntHaveMorph(relation string, types []any, args ...any) Query

	// WithAggregate adds a sub-select query to include an aggregate value for a relationship.
	// fn must be one of: count, max, min, sum, avg, exists.
	WithAggregate(relation, column, fn string, args ...any) Query
	// WithCount adds sub-select queries to count the relations. Each entry may be either a
	// string ("Books") or a RelationCount struct for scoped/aliased counts.
	WithCount(relations ...any) Query
	// WithMax adds sub-select queries to include the max of the relation's column.
	WithMax(relation, column string, args ...any) Query
	// WithMin adds sub-select queries to include the min of the relation's column.
	WithMin(relation, column string, args ...any) Query
	// WithSum adds sub-select queries to include the sum of the relation's column.
	WithSum(relation, column string, args ...any) Query
	// WithAvg adds sub-select queries to include the average of the relation's column.
	WithAvg(relation, column string, args ...any) Query
	// WithExists adds sub-select queries to include the existence of related models. The result
	// is emitted as `CASE WHEN EXISTS (...) THEN 1 ELSE 0 END` for cross-dialect portability
	// (SQL Server has no boolean literal), but the dest field may be either `bool` or an integer
	// type - Go's database/sql layer converts 0/1 ints to bool automatically.
	WithExists(relations ...string) Query
}

// RelationCount is an entry accepted by WithCount that pairs a relation name with an optional
// scope callback and result alias. Equivalent to the array-keyed `withCount(['posts as p_count' =>
// fn ...])` idiom in Laravel, expressed as a Go struct:
//
//	q.WithCount(orm.RelationCount{Name: "Books", Alias: "book_total", Callback: func(q) q.Where(...)})
type RelationCount struct {
	// Name is the relation method/field name on the parent model (e.g. "Books").
	Name string
	// Alias overrides the default `<relation>_count` column alias when non-empty.
	Alias string
	// Callback scopes the inner count subquery, mirroring the upstream array-keyed callback shape.
	Callback RelationCallback
}

// RelationKind names a relationship flavour for diagnostic / error-message use only. The
// per-kind structs below (HasOne, HasMany, ...) are the actual user-facing declaration types;
// the RelationKind constants exist purely so error messages can refer to a kind by name.
type RelationKind string

const (
	KindHasOne         RelationKind = "hasOne"
	KindHasMany        RelationKind = "hasMany"
	KindBelongsTo      RelationKind = "belongsTo"
	KindMany2Many      RelationKind = "many2Many"
	KindMorphOne       RelationKind = "morphOne"
	KindMorphMany      RelationKind = "morphMany"
	KindMorphTo        RelationKind = "morphTo"
	KindMorphToMany    RelationKind = "morphToMany"
	KindMorphedByMany  RelationKind = "morphedByMany"
	KindHasOneThrough  RelationKind = "hasOneThrough"
	KindHasManyThrough RelationKind = "hasManyThrough"
)

// Relation is the sealed interface implemented by every per-kind relation declaration struct
// (HasOne, HasMany, BelongsTo, Many2Many, MorphOne, MorphMany, MorphTo, MorphToMany,
// MorphedByMany, HasOneThrough, HasManyThrough). The relation() method is unexported so external
// packages cannot define new kinds — the resolver type-switches on the closed set defined here.
//
// Models declare their relationships in a single map:
//
//	func (User) Relations() map[string]orm.Relation {
//	    return map[string]orm.Relation{
//	        "Books":   orm.HasMany{Related: &Book{}},
//	        "Roles":   orm.Many2Many{Related: &Role{}, Table: "user_roles"},
//	        "Houses":  orm.MorphMany{Related: &House{}, Name: "houseable"},
//	        "Posts":   orm.HasManyThrough{Related: &Post{}, Through: &Account{}},
//	    }
//	}
//
// All relation fields on the model struct must be tagged `gorm:"-"` so GORM doesn't try to
// auto-resolve them from struct tags.
type Relation interface {
	// Kind returns the relation flavour for diagnostics (error messages, logging). The resolver
	// itself dispatches by Go type, not by the Kind value.
	Kind() RelationKind
}

// HasOne declares a one-to-one relation where the related row holds a foreign key referencing
// this model.
//
// Defaults: ForeignKey = singular(parentTable) + "_id"; LocalKey = "id".
type HasOne struct {
	// Related is a sample instance of the related model (e.g. &Profile{}).
	Related any
	// ForeignKey is the column on the related table referencing the parent. Optional.
	ForeignKey string
	// LocalKey is the column on the parent referenced by ForeignKey. Optional, defaults to "id".
	LocalKey string
	// OnQuery is a default scope applied to every query built for this relation (eager loads,
	// existence checks, aggregates, Related). Applied before any caller-supplied callback.
	OnQuery RelationCallback
}

func (HasOne) Kind() RelationKind { return KindHasOne }

// HasMany declares a one-to-many relation — the multi-result variant of HasOne.
//
// Defaults: ForeignKey = singular(parentTable) + "_id"; LocalKey = "id".
type HasMany struct {
	Related    any
	ForeignKey string
	LocalKey   string
	OnQuery    RelationCallback
}

func (HasMany) Kind() RelationKind { return KindHasMany }

// BelongsTo declares the inverse of HasOne / HasMany — this model holds a foreign key
// referencing the related row.
//
// Defaults: ForeignKey = singular(relatedTable) + "_id"; OwnerKey = "id".
type BelongsTo struct {
	Related any
	// ForeignKey is the column on the parent table referencing the related row. Optional.
	ForeignKey string
	// OwnerKey is the column on the related table referenced by ForeignKey. Optional, "id".
	OwnerKey string
	OnQuery  RelationCallback
}

func (BelongsTo) Kind() RelationKind { return KindBelongsTo }

// Many2Many declares a many-to-many relation through a pivot table.
//
// Defaults:
//
//	Table            = alphabetical singular pair (e.g. "post_tag")
//	ForeignPivotKey  = singular(parentTable) + "_id"
//	RelatedPivotKey  = singular(relatedTable) + "_id"
//	ParentKey        = "id"
//	RelatedKey       = "id"
type Many2Many struct {
	Related any
	// Table is the pivot table name. Optional.
	Table string
	// ForeignPivotKey is the pivot column referencing the parent. Optional.
	ForeignPivotKey string
	// RelatedPivotKey is the pivot column referencing the related. Optional.
	RelatedPivotKey string
	// ParentKey is the column on the parent referenced by ForeignPivotKey. Optional, "id".
	ParentKey string
	// RelatedKey is the column on the related referenced by RelatedPivotKey. Optional, "id".
	RelatedKey string
	// PivotField is the name of the struct field on the related model that the eager loader will
	// hydrate with pivot column values (e.g. "Pivot", "UserPivot"). Optional — defaults to
	// "Pivot". The field's Go type drives both the pivot SELECT list (every db-tagged column on
	// the struct) and the hydration target. When the related model has no field by this name,
	// no Pivot hydration happens — the relation still works for joining, just doesn't surface
	// pivot columns. Use a non-default name when one related model serves multiple m2m relations
	// with different pivot schemas (e.g. Role with both UserPivot and GroupPivot fields).
	PivotField string
	// PivotTimestamps enables auto-stamping of the pivot table's created_at / updated_at columns
	// on Attach / Sync / Save (and updated_at on UpdateExistingPivot), using default column names.
	// Most users don't need to set this flag explicitly — see "Detection priority" below.
	//
	// Detection priority for pivot timestamps (highest first):
	//
	//   1. Pivot struct field with `gorm:"autoCreateTime"` / `gorm:"autoUpdateTime"` tag. Works
	//      for any field name; column name is taken from the struct's GORM schema.
	//   2. Pivot struct field named CreatedAt / UpdatedAt of type time.Time (Go/GORM convention).
	//   3. PivotTimestamps: true. Fallback for when no Pivot struct is declared (or its struct
	//      has no timestamp fields) but the underlying table still has created_at / updated_at
	//      columns you want auto-filled. Uses default column names.
	//
	// Customize column names via `gorm:"column:..."` on the Pivot struct field. There is
	// intentionally no relation-level override — the Pivot struct is the single source of truth
	// for column metadata.
	PivotTimestamps bool
	OnQuery         RelationCallback
	// OnPivotQuery scopes pivot-table SELECT / UPDATE / DELETE for Sync / Detach /
	// UpdateExistingPivot operations on this relation. Equality conditions added here are NOT
	// auto-injected into Attach INSERT rows — pass them via Attach attrs if needed.
	OnPivotQuery PivotCallback
	// Touches, when true, causes Sync / Attach / Detach / Toggle / UpdateExistingPivot on this
	// relation to bump the parent's `updated_at` after the pivot write completes (and only when
	// the operation actually changed pivot rows). Mirrors fedaco's `touchIfTouching`. Silently
	// no-ops when the parent's schema doesn't carry an updated_at field.
	Touches bool
}

func (Many2Many) Kind() RelationKind { return KindMany2Many }

// MorphOne declares a one-to-one polymorphic relation — the related row holds <Name>_id and
// <Name>_type referencing one of several possible parent kinds.
//
// Defaults: TypeColumn = Name + "_type"; IDColumn = Name + "_id"; LocalKey = "id".
type MorphOne struct {
	Related any
	// Name is the polymorphic name (e.g. "imageable", "taggable"). Required.
	Name string
	// TypeColumn is the polymorphic type column on the related table. Optional.
	TypeColumn string
	// IDColumn is the polymorphic id column on the related table. Optional.
	IDColumn string
	// LocalKey is the column on the parent referenced by IDColumn. Optional, "id".
	LocalKey string
	OnQuery  RelationCallback
}

func (MorphOne) Kind() RelationKind { return KindMorphOne }

// MorphMany is the multi-result variant of MorphOne.
type MorphMany struct {
	Related    any
	Name       string
	TypeColumn string
	IDColumn   string
	LocalKey   string
	OnQuery    RelationCallback
}

func (MorphMany) Kind() RelationKind { return KindMorphMany }

// MorphTo declares the inverse polymorphic side: this model holds <Name>_id + <Name>_type and
// resolves to one of several parent kinds via the morph map registry. There is no Related — the
// concrete type is determined per row from the type column.
//
// Defaults: TypeColumn = Name + "_type"; IDColumn = Name + "_id"; OwnerKey = "id".
type MorphTo struct {
	// Name is the polymorphic name. Required.
	Name string
	// TypeColumn is the polymorphic type column on this table. Optional.
	TypeColumn string
	// IDColumn is the polymorphic id column on this table. Optional.
	IDColumn string
	// OwnerKey is the column on each related table referenced by IDColumn. Optional, "id".
	OwnerKey string
	OnQuery  RelationCallback
}

func (MorphTo) Kind() RelationKind { return KindMorphTo }

// MorphToMany declares a polymorphic many-to-many — through a pivot that carries
// <Name>_id + <Name>_type plus a related FK.
//
// Defaults:
//
//	Table            = pluralize(Name)  (e.g. "taggables")
//	TypeColumn       = Name + "_type"
//	ForeignPivotKey  = Name + "_id"
//	RelatedPivotKey  = singular(relatedTable) + "_id"
//	ParentKey        = "id"
//	RelatedKey       = "id"
type MorphToMany struct {
	Related         any
	Name            string
	Table           string
	TypeColumn      string
	ForeignPivotKey string
	RelatedPivotKey string
	ParentKey       string
	RelatedKey      string
	// PivotField — see Many2Many.PivotField.
	PivotField string
	// PivotTimestamps — see Many2Many.PivotTimestamps.
	PivotTimestamps bool
	OnQuery         RelationCallback
	// OnPivotQuery — see Many2Many.OnPivotQuery.
	OnPivotQuery PivotCallback
	// Touches — see Many2Many.Touches.
	Touches bool
}

func (MorphToMany) Kind() RelationKind { return KindMorphToMany }

// MorphedByMany is the inverse side of MorphToMany — the morph value pins on the related rather
// than the parent. Field semantics and defaults match MorphToMany.
type MorphedByMany struct {
	Related         any
	Name            string
	Table           string
	TypeColumn      string
	ForeignPivotKey string
	RelatedPivotKey string
	ParentKey       string
	RelatedKey      string
	// PivotField — see Many2Many.PivotField.
	PivotField string
	// PivotTimestamps — see Many2Many.PivotTimestamps.
	PivotTimestamps bool
	OnQuery         RelationCallback
	// OnPivotQuery — see Many2Many.OnPivotQuery.
	OnPivotQuery PivotCallback
	// Touches — see Many2Many.Touches.
	Touches bool
}

func (MorphedByMany) Kind() RelationKind { return KindMorphedByMany }

// HasOneThrough declares a relation reached through an intermediate ("through") table.
//
// Defaults:
//
//	FirstKey       = singular(parentTable) + "_id"
//	SecondKey      = singular(throughTable) + "_id"
//	LocalKey       = "id"
//	SecondLocalKey = "id"
type HasOneThrough struct {
	Related any
	// Through is the intermediate model.
	Through any
	// FirstKey is the FK on the through table pointing at parent. Optional.
	FirstKey string
	// SecondKey is the FK on the related table pointing at through. Optional.
	SecondKey string
	// LocalKey is the PK on the parent referenced by FirstKey. Optional, "id".
	LocalKey string
	// SecondLocalKey is the PK on the through table referenced by SecondKey. Optional, "id".
	SecondLocalKey string
	OnQuery        RelationCallback
}

func (HasOneThrough) Kind() RelationKind { return KindHasOneThrough }

// HasManyThrough is the multi-result variant of HasOneThrough.
type HasManyThrough struct {
	Related        any
	Through        any
	FirstKey       string
	SecondKey      string
	LocalKey       string
	SecondLocalKey string
	OnQuery        RelationCallback
}

func (HasManyThrough) Kind() RelationKind { return KindHasManyThrough }

// ModelWithRelations is implemented by every model that declares relationships. The single map
// returned by Relations() is the only place relations are declared. GORM relation struct tags
// (`foreignKey:`, `references:`, `many2many:`, `polymorphic:`) are forbidden — fields that hold
// related rows must be tagged `gorm:"-"`.
type ModelWithRelations interface {
	Relations() map[string]Relation
}
