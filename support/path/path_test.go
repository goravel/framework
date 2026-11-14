package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanRelative(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
		ok       bool
	}{
		{name: "empty is the root", path: "", expected: ".", ok: true},
		{name: "dot is the root", path: ".", expected: ".", ok: true},
		{name: "plain path", path: "public", expected: "public", ok: true},
		{name: "nested path", path: "resources/assets", expected: "resources/assets", ok: true},
		{name: "dot segment", path: "a/./b", expected: "a/b", ok: true},
		{name: "parent segment inside the root", path: "a/../b", expected: "b", ok: true},
		{name: "leaves a subdirectory, stays inside", path: "resources/js/../assets/x.jpg", expected: "resources/assets/x.jpg", ok: true},
		{name: "trailing slash", path: "public/", expected: "public", ok: true},
		{name: "parent", path: "..", ok: false},
		{name: "starts with parent", path: "../shop", ok: false},
		{name: "escapes after cleaning", path: "a/../../x", ok: false},
		{name: "deep escape", path: "modules/blog/resources/js/../../../../../.env", ok: false},
		{name: "absolute", path: "/etc/passwd", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned, ok := CleanRelative(tt.path)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.expected, cleaned)
		})
	}
}
