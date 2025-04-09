package enum

import "github.com/spf13/cast"

type Enum[K comparable, V any] struct {
	key   K
	value V
}

func NewEnum[K comparable, V any](key K, value V) Enum[K, V] {
	return Enum[K, V]{
		key:   key,
		value: value,
	}
}

func (r Enum[K, V]) Key() K {
	return r.key
}

func (r Enum[K, V]) Value() V {
	return r.value
}

func (r Enum[K, V]) String() string {
	return cast.ToString(r.value)
}

func ParseEnumByKey[K comparable, V any](id K, enumMap map[K]Enum[K, V]) Enum[K, V] {
	if val, exists := enumMap[id]; exists {
		return val
	}

	return Enum[K, V]{}
}

func ParseEnumByValue[K comparable, V comparable](value V, enumMap map[V]Enum[K, V]) Enum[K, V] {
	if val, exists := enumMap[value]; exists {
		return val
	}

	return Enum[K, V]{}
}

func GenerateEnumMaps[K comparable, V comparable](list []Enum[K, V]) (map[K]Enum[K, V], map[V]Enum[K, V]) {
	keyMap := make(map[K]Enum[K, V])
	valueMap := make(map[V]Enum[K, V])
	for _, entry := range list {
		keyMap[entry.Key()] = entry
		valueMap[entry.Value()] = entry
	}

	return keyMap, valueMap
}
