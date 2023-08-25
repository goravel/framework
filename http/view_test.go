package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestView(t *testing.T) {
	view := NewView()
	view.Share("a", "b")
	assert.Equal(t, "b", view.Shared("a"))
	assert.Equal(t, "c", view.Shared("b", "c"))
	assert.Equal(t, map[string]any{"a": "b"}, view.GetShared())
}
