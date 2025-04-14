package enum

import (
	"encoding/json"
	"errors"

	"github.com/spf13/cast"
)

var ErrEnumNotFound = errors.New("enum not found")

type Enum[K comparable, V any] interface {
	Key() K
	Value() V
	String() string
	MarshalJSON() ([]byte, error)
}

type Impl[K comparable, V any] struct {
	key   K
	value V
}

func New[K comparable, V any](key K, value V) Enum[K, V] {
	return &Impl[K, V]{
		key:   key,
		value: value,
	}
}

func (r *Impl[K, V]) Key() K {
	return r.key
}

func (r *Impl[K, V]) Value() V {
	return r.value
}

func (r *Impl[K, V]) String() string {
	return cast.ToString(r.value)
}

func (r *Impl[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.value)
}

func ParseEnumByKey[K comparable, V any, E Enum[K, V]](id K, enumMap map[K]E) (E, error) {
	if val, exists := enumMap[id]; exists {
		return val, nil
	}

	var zero E
	return zero, ErrEnumNotFound
}

func ParseEnumByValue[K comparable, V comparable, E Enum[K, V]](value V, enumMap map[V]E) (E, error) {
	if val, exists := enumMap[value]; exists {
		return val, nil
	}

	var zero E
	return zero, ErrEnumNotFound
}

func GenerateEnumMaps[K comparable, V comparable, E Enum[K, V]](list []E) (map[K]E, map[V]E) {
	keyMap := make(map[K]E)
	valueMap := make(map[V]E)

	for _, entry := range list {
		keyMap[entry.Key()] = entry
		valueMap[entry.Value()] = entry
	}

	return keyMap, valueMap
}
