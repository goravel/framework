package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeepCopyMap(t *testing.T) {
	tests := []struct {
		name   string
		source map[string]any
	}{
		{name: "empty", source: map[string]any{}},
		{name: "scalars", source: map[string]any{"string": "a", "int": 1, "bool": true, "nil": nil}},
		{name: "nested map", source: map[string]any{"a": map[string]any{"b": map[string]any{"c": "d"}}}},
		{name: "slice of any", source: map[string]any{"a": []any{map[string]any{"b": 1}, nil}}},
		{name: "typed slices", source: map[string]any{"strings": []string{"a"}, "ints": []int{1}, "nil_strings": []string(nil)}},
		{name: "typed maps", source: map[string]any{"strings": map[string]string{"a": "b"}, "nil_map": map[string]any(nil)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.source, deepCopyMap(tt.source))
		})
	}
}

func TestDeepCopyMap_DoesNotShareValues(t *testing.T) {
	source := map[string]any{
		"nested":  map[string]any{"value": "origin"},
		"list":    []any{map[string]any{"value": "origin"}},
		"typed":   []string{"origin"},
		"headers": map[string]string{"X-Foo": "origin"},
	}
	copied := deepCopyMap(source)

	copied["nested"].(map[string]any)["value"] = "mutated"
	copied["list"].([]any)[0].(map[string]any)["value"] = "mutated"
	copied["typed"].([]string)[0] = "mutated"
	copied["headers"].(map[string]string)["X-Foo"] = "mutated"

	assert.Equal(t, "origin", source["nested"].(map[string]any)["value"])
	assert.Equal(t, "origin", source["list"].([]any)[0].(map[string]any)["value"])
	assert.Equal(t, "origin", source["typed"].([]string)[0])
	assert.Equal(t, "origin", source["headers"].(map[string]string)["X-Foo"])
}
