package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testContextKey any

// utilsContextKey mirrors the typed-string pattern used by framework
// packages (`type contextKey string`) so we exercise the stringification path.
type utilsContextKey string

func TestGetContextValues(t *testing.T) {
	ctx := context.Background()
	values := make(map[any]any)
	getContextValues(ctx, values)
	assert.Equal(t, make(map[any]any), values)

	ctx = context.WithValue(ctx, testContextKey("a"), "b")
	ctx = context.WithValue(ctx, testContextKey(1), 2)
	ctx = context.WithValue(ctx, testContextKey("c"), map[string]any{"d": "e"})

	type T struct {
		A string
	}
	ctx = context.WithValue(ctx, testContextKey("d"), T{A: "a"})

	values = make(map[any]any)
	getContextValues(ctx, values)
	assert.Equal(t, map[any]any{
		"a": "b",
		1:   2,
		"c": map[string]any{"d": "e"},
		"d": T{A: "a"},
	}, values)
}

func TestGetContextValues_TypedNilPointer(t *testing.T) {
	var typedNil *struct{ Context context.Context }
	values := make(map[any]any)
	assert.NotPanics(t, func() {
		getContextValues(typedNil, values)
	})
	assert.Empty(t, values)
}

func TestFilterContextValues(t *testing.T) {
	tests := []struct {
		name   string
		values map[any]any
		user   []string
		expect map[string]any
	}{
		{
			name:   "empty input returns nil",
			values: map[any]any{},
			expect: nil,
		},
		{
			name: "framework keys (plain string) dropped by default",
			values: map[any]any{
				"GoravelAuthJwt":           "secret",
				"goravel_http_client_name": "client-a",
				"locale":                   "en",
				"fallback_locale":          "en-US",
				"request_id":               "req-1",
			},
			expect: map[string]any{"request_id": "req-1"},
		},
		{
			name: "framework keys (typed string) dropped by default",
			values: map[any]any{
				utilsContextKey("GoravelAuthJwt"): "secret",
				utilsContextKey("locale"):         "en",
				utilsContextKey("request_id"):     "req-1",
			},
			expect: map[string]any{"request_id": "req-1"},
		},
		{
			name: "user-supplied keys extend defaults",
			values: map[any]any{
				"GoravelAuthJwt": "secret",
				"trace_id":       "t-1",
				"request_id":     "req-1",
			},
			user:   []string{"trace_id"},
			expect: map[string]any{"request_id": "req-1"},
		},
		{
			name:   "non-string keys kept under their %v form",
			values: map[any]any{42: "answer"},
			expect: map[string]any{"42": "answer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterContext(tt.values, newExcludeSet(tt.user))
			assert.Equal(t, tt.expect, got)
		})
	}
}
