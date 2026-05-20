package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRelation_Kind verifies the Kind() method on every relation type returns the expected
// constant. The resolver dispatches on the concrete type rather than this value, but Kind() is
// used by error messages and diagnostics.
func TestRelation_Kind(t *testing.T) {
	tests := []struct {
		name     string
		relation Relation
		expected RelationKind
	}{
		{"HasOne", HasOne{}, KindHasOne},
		{"HasMany", HasMany{}, KindHasMany},
		{"BelongsTo", BelongsTo{}, KindBelongsTo},
		{"Many2Many", Many2Many{}, KindMany2Many},
		{"MorphOne", MorphOne{}, KindMorphOne},
		{"MorphMany", MorphMany{}, KindMorphMany},
		{"MorphTo", MorphTo{}, KindMorphTo},
		{"MorphToMany", MorphToMany{}, KindMorphToMany},
		{"MorphedByMany", MorphedByMany{}, KindMorphedByMany},
		{"HasOneThrough", HasOneThrough{}, KindHasOneThrough},
		{"HasManyThrough", HasManyThrough{}, KindHasManyThrough},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.relation.Kind())
		})
	}
}

// TestRelation_KindConstants verifies the named constants have the expected string values.
// These strings appear in error messages, so they're part of the public contract.
func TestRelation_KindConstants(t *testing.T) {
	assert.Equal(t, RelationKind("hasOne"), KindHasOne)
	assert.Equal(t, RelationKind("hasMany"), KindHasMany)
	assert.Equal(t, RelationKind("belongsTo"), KindBelongsTo)
	assert.Equal(t, RelationKind("many2Many"), KindMany2Many)
	assert.Equal(t, RelationKind("morphOne"), KindMorphOne)
	assert.Equal(t, RelationKind("morphMany"), KindMorphMany)
	assert.Equal(t, RelationKind("morphTo"), KindMorphTo)
	assert.Equal(t, RelationKind("morphToMany"), KindMorphToMany)
	assert.Equal(t, RelationKind("morphedByMany"), KindMorphedByMany)
	assert.Equal(t, RelationKind("hasOneThrough"), KindHasOneThrough)
	assert.Equal(t, RelationKind("hasManyThrough"), KindHasManyThrough)
}
