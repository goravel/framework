// Package morphmap holds the process-wide registry of polymorphic aliases shared by the orm
// and gorm packages. It mirrors fedaco's static Relation._morphMap (libs/fedaco/src/fedaco/
// relations/relation.ts:31) while remaining safe for concurrent reads from many query goroutines.
//
// Most app code interacts with the registry via the wrappers in package orm
// (orm.MorphMap, orm.MorphedModel, orm.MorphAlias). The lower-level gorm wrapper imports this
// package directly to resolve morph values during query construction without introducing a
// circular dependency on package orm.
package morphmap

import (
	"reflect"
	"sync"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

var (
	mu          sync.RWMutex
	aliasToType = map[string]reflect.Type{}
	typeToAlias = map[reflect.Type]string{}
)

// Register stores a set of alias-to-sample-model bindings. Subsequent calls merge with previously
// registered entries; later writes win on conflict. Pass false in the optional merge argument to
// replace the registry instead of merging.
func Register(entries map[string]any, merge ...bool) {
	doMerge := true
	if len(merge) > 0 {
		doMerge = merge[0]
	}
	mu.Lock()
	defer mu.Unlock()
	if !doMerge {
		aliasToType = map[string]reflect.Type{}
		typeToAlias = map[reflect.Type]string{}
	}
	for alias, sample := range entries {
		typ := indirectType(reflect.TypeOf(sample))
		if typ == nil {
			continue
		}
		// Drop any previous bindings on either side so re-registration leaves a single canonical
		// mapping in both directions.
		if oldType, ok := aliasToType[alias]; ok {
			delete(typeToAlias, oldType)
		}
		if oldAlias, ok := typeToAlias[typ]; ok {
			delete(aliasToType, oldAlias)
		}
		aliasToType[alias] = typ
		typeToAlias[typ] = alias
	}
}

// Find returns a fresh pointer to a new instance of the model registered under alias, or nil if
// no model is registered.
func Find(alias string) any {
	mu.RLock()
	typ, ok := aliasToType[alias]
	mu.RUnlock()
	if !ok {
		return nil
	}
	return reflect.New(typ).Interface()
}

// AliasOf returns the alias registered for the given model's underlying type.
func AliasOf(model any) (string, bool) {
	typ := indirectType(reflect.TypeOf(model))
	if typ == nil {
		return "", false
	}
	mu.RLock()
	defer mu.RUnlock()
	alias, ok := typeToAlias[typ]
	return alias, ok
}

// All returns a snapshot copy of the alias-to-type registry.
func All() map[string]reflect.Type {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]reflect.Type, len(aliasToType))
	for alias, typ := range aliasToType {
		out[alias] = typ
	}
	return out
}

// Reset clears all entries. Intended for tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	aliasToType = map[string]reflect.Type{}
	typeToAlias = map[reflect.Type]string{}
}

// MorphValue resolves the morph alias for a model. Resolution order:
//  1. model.MorphClass() if the model (or its pointer) implements ModelWithMorphClass
//  2. global morph map (registered via Register)
//
// Returns "" and false if neither resolves. The caller is then expected to fall back to GORM's
// `polymorphicValue:` tag or the parent's table name.
func MorphValue(model any) (string, bool) {
	if alias, ok := tryMorphClass(model); ok && alias != "" {
		return alias, true
	}
	if alias, ok := AliasOf(model); ok {
		return alias, true
	}
	return "", false
}

// tryMorphClass invokes MorphClass() on model whether it has a value-receiver method or a
// pointer-receiver method, and whether the caller passed a value or a pointer.
func tryMorphClass(model any) (string, bool) {
	if m, ok := model.(contractsorm.ModelWithMorphClass); ok {
		return m.MorphClass(), true
	}
	rv := reflect.ValueOf(model)
	switch rv.Kind() {
	case reflect.Pointer:
		if rv.IsNil() {
			return "", false
		}
		if m, ok := rv.Elem().Interface().(contractsorm.ModelWithMorphClass); ok {
			return m.MorphClass(), true
		}
	case reflect.Struct:
		ptr := reflect.New(rv.Type())
		ptr.Elem().Set(rv)
		if m, ok := ptr.Interface().(contractsorm.ModelWithMorphClass); ok {
			return m.MorphClass(), true
		}
	}
	return "", false
}

func indirectType(t reflect.Type) reflect.Type {
	if t == nil {
		return nil
	}
	if t.Kind() == reflect.Pointer {
		return t.Elem()
	}
	return t
}
