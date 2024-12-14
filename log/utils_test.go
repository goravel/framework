package log

import (
	"context"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestGetContextValues(t *testing.T) {
	ctx := context.Background()
	values := make(map[any]any)
	getContextValues(ctx, values)
	assert.Equal(t, make(map[any]any), values)

	ctx = context.WithValue(ctx, "a", "b")
	ctx = context.WithValue(ctx, 1, 2)
	ctx = context.WithValue(ctx, "c", map[string]any{"d": "e"})

	type T struct {
		A string
	}
	ctx = context.WithValue(ctx, "d", T{A: "a"})

	values = make(map[any]any)
	getContextValues(ctx, values)
	assert.Equal(t, map[any]any{
		"a": "b",
		1:   2,
		"c": map[string]any{"d": "e"},
		"d": T{A: "a"},
	}, values)
}
