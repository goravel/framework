package orm

import (
	"github.com/goravel/framework/database/orm/morphmap"
)

// MorphMap registers polymorphic aliases. Each entry maps an alias (the value stored in a
// `*_type` column) to a sample model instance from which the registry derives the underlying Go
// type.
//
// Subsequent calls merge with previously registered entries; later writes win on conflict. Pass
// false in the optional merge argument to replace the registry instead of merging.
//
//	orm.MorphMap(map[string]any{
//	    "post":  &Post{},
//	    "video": &Video{},
//	})
//
// Models that implement orm.ModelWithMorphClass take precedence over the registry.
func MorphMap(entries map[string]any, merge ...bool) {
	morphmap.Register(entries, merge...)
}

// MorphedModel returns a fresh pointer to a new instance of the model registered under alias, or
// nil if no model is registered. Used by the MorphTo loader to allocate the right Go type for a
// row whose `*_type` column contains alias.
func MorphedModel(alias string) any {
	return morphmap.Find(alias)
}

// MorphAlias returns the alias registered for the given model's underlying type. Useful at insert
// time when writing the value of a `*_type` column from a Go value.
func MorphAlias(model any) (string, bool) {
	return morphmap.AliasOf(model)
}
