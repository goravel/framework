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

type structSentinelKey struct{}
type otherStructSentinel struct{}
type collidingNameKey string
type collidingNameKey2 string
type collidingNameKey3 string

func TestFilterContextValues(t *testing.T) {
	tests := []struct {
		name   string
		values map[any]any
		user   []any
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
				"request_id":               "req-1",
			},
			expect: map[string]any{"request_id": "req-1"},
		},
		{
			name: "framework keys (typed string) dropped by default",
			values: map[any]any{
				utilsContextKey("GoravelAuthJwt"): "secret",
				utilsContextKey("request_id"):    "req-1",
			},
			expect: map[string]any{"request_id": "req-1"},
		},
		{
			name: "user string entries extend defaults",
			values: map[any]any{
				"GoravelAuthJwt": "secret",
				"trace_id":       "t-1",
				"request_id":     "req-1",
			},
			user:   []any{"trace_id"},
			expect: map[string]any{"request_id": "req-1"},
		},
		{
			name: "struct-sentinel keys use short type name when unique",
			values: map[any]any{
				structSentinelKey{}: "user-42",
				"request_id":        "req-1",
			},
			expect: map[string]any{
				"logger.structSentinelKey": "user-42",
				"request_id":               "req-1",
			},
		},
		{
			name: "distinct struct sentinels keep short names",
			values: map[any]any{
				structSentinelKey{}:   "from-logger",
				otherStructSentinel{}: "from-other",
			},
			expect: map[string]any{
				"logger.structSentinelKey":   "from-logger",
				"logger.otherStructSentinel": "from-other",
			},
		},
		{
			name: "user can exclude a struct-sentinel by passing the value",
			values: map[any]any{
				structSentinelKey{}: "user-42",
				"request_id":        "req-1",
			},
			user:   []any{structSentinelKey{}},
			expect: map[string]any{"request_id": "req-1"},
		},
		{
			name: "default excludes still apply when keys with the same label collide",
			values: map[any]any{
				"GoravelAuthJwt":                  "from-string",
				utilsContextKey("GoravelAuthJwt"): "from-typed",
				"request_id":                      "req-1",
			},
			expect: map[string]any{"request_id": "req-1"},
		},
		{
			name: "uncomparable user entries are ignored, not panicking",
			values: map[any]any{
				"request_id": "req-1",
			},
			user:   []any{[]string{"oops"}},
			expect: map[string]any{"request_id": "req-1"},
		},
		{
			name: "colliding labels escalate typed keys to type name",
			values: map[any]any{
				"session_id":                   "from-string",
				collidingNameKey("session_id"): "from-typed",
			},
			expect: map[string]any{
				"session_id":              "from-string",
				"logger.collidingNameKey": "from-typed",
			},
		},
		{
			name: "three-way label collision escalates every key",
			values: map[any]any{
				collidingNameKey("session_id"):  "from-typed-1",
				collidingNameKey2("session_id"): "from-typed-2",
				collidingNameKey3("session_id"): "from-typed-3",
			},
			expect: map[string]any{
				"logger.collidingNameKey":  "from-typed-1",
				"logger.collidingNameKey2": "from-typed-2",
				"logger.collidingNameKey3": "from-typed-3",
			},
		},
		{
			name: "user can exclude by qualified name without collision",
			values: map[any]any{
				structSentinelKey{}: "user-42",
			},
			user:   []any{"github.com/goravel/framework/log/logger.structSentinelKey"},
			expect: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterContext(tt.values, newExcludeSet(tt.user))
			assert.Equal(t, tt.expect, got)
		})
	}
}
