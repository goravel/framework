package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

func TestParseEagerLoad(t *testing.T) {
	cb := contractsorm.RelationCallback(func(q contractsorm.Query) contractsorm.Query { return q })

	type expected struct {
		relation    string
		columns     []string
		callbackSet bool
	}

	cases := []struct {
		name string
		args []any
		want []expected
	}{
		{
			name: "single string",
			args: []any{"Books"},
			want: []expected{{relation: "Books"}},
		},
		{
			name: "string + callback",
			args: []any{"Books", cb},
			want: []expected{{relation: "Books", callbackSet: true}},
		},
		{
			name: "multiple strings",
			args: []any{"Books", "Roles", "Address"},
			want: []expected{
				{relation: "Books"},
				{relation: "Roles"},
				{relation: "Address"},
			},
		},
		{
			name: "column pruning",
			args: []any{"Books:id,name"},
			want: []expected{{relation: "Books", columns: []string{"id", "name"}}},
		},
		{
			name: "column pruning with whitespace",
			args: []any{"Books: id , name "},
			want: []expected{{relation: "Books", columns: []string{"id", "name"}}},
		},
		{
			name: "map with callback",
			args: []any{map[string]contractsorm.RelationCallback{"Books": cb}},
			want: []expected{{relation: "Books", callbackSet: true}},
		},
		{
			name: "map with nil callback",
			args: []any{map[string]contractsorm.RelationCallback{"Roles": nil}},
			want: []expected{{relation: "Roles"}},
		},
		{
			name: "func map literal",
			args: []any{map[string]func(contractsorm.Query) contractsorm.Query{"Roles": func(q contractsorm.Query) contractsorm.Query { return q }}},
			want: []expected{{relation: "Roles", callbackSet: true}},
		},
		{
			name: "[]string",
			args: []any{[]string{"Books", "Roles"}},
			want: []expected{
				{relation: "Books"},
				{relation: "Roles"},
			},
		},
		{
			name: "[]any mix",
			args: []any{[]any{"Books", map[string]contractsorm.RelationCallback{"Roles": cb}}},
			want: []expected{
				{relation: "Books"},
				{relation: "Roles", callbackSet: true},
			},
		},
		{
			name: "nested dot",
			args: []any{"Books.Author"},
			want: []expected{
				{relation: "Books"},
				{relation: "Books.Author"},
			},
		},
		{
			name: "nested dot with callback on leaf",
			args: []any{"Books.Author", cb},
			want: []expected{
				{relation: "Books"},
				{relation: "Books.Author", callbackSet: true},
			},
		},
		{
			name: "duplicate later wins",
			args: []any{"Books", map[string]contractsorm.RelationCallback{"Books": cb}},
			want: []expected{{relation: "Books", callbackSet: true}},
		},
		{
			name: "synthetic prefix does not clobber leaf already set",
			args: []any{"Books", "Books.Author"},
			want: []expected{
				{relation: "Books"},
				{relation: "Books.Author"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseEagerLoad(tc.args)
			assert.NoError(t, err)
			assert.Len(t, got, len(tc.want), "entry count")
			for i := 0; i < len(tc.want) && i < len(got); i++ {
				assert.Equal(t, tc.want[i].relation, got[i].relation, "relation[%d]", i)
				assert.Equal(t, tc.want[i].columns, got[i].columns, "columns[%d]", i)
				assert.Equal(t, tc.want[i].callbackSet, got[i].callback != nil, "callback set[%d]", i)
			}
		})
	}
}

func TestParseEagerLoadErrors(t *testing.T) {
	t.Run("unsupported type", func(t *testing.T) {
		_, err := parseEagerLoad([]any{123})
		assert.True(t, errors.Is(err, errors.OrmEagerLoadInvalidArgument))
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := parseEagerLoad([]any{""})
		assert.True(t, errors.Is(err, errors.OrmEagerLoadEmptyRelation))
	})

	t.Run("dot path with empty segment", func(t *testing.T) {
		_, err := parseEagerLoad([]any{"Books..Author"})
		assert.True(t, errors.Is(err, errors.OrmEagerLoadEmptyRelation))
	})

	t.Run("nil values are skipped", func(t *testing.T) {
		got, err := parseEagerLoad([]any{nil, "Books", nil})
		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "Books", got[0].relation)
	})
}

func TestSplitRelationSelect(t *testing.T) {
	cases := []struct {
		raw     string
		name    string
		columns []string
	}{
		{"Books", "Books", nil},
		{"Books:id", "Books", []string{"id"}},
		{"Books:id,name", "Books", []string{"id", "name"}},
		{"Books: id , name ", "Books", []string{"id", "name"}},
		{"Books:", "Books", nil},
		{"Books:,,", "Books", nil},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			name, cols := splitRelationSelect(tc.raw)
			assert.Equal(t, tc.name, name)
			assert.Equal(t, tc.columns, cols)
		})
	}
}

func TestDirectNestedEntries(t *testing.T) {
	list, err := parseEagerLoad([]any{"Books", "Books.Author", "Books.Reviews", "Roles"})
	assert.NoError(t, err)

	got := directNestedEntries(list, "Books")
	assert.Len(t, got, 2)
	assert.Equal(t, "Author", got[0].relation)
	assert.Equal(t, "Reviews", got[1].relation)
}

func TestChunkKeys(t *testing.T) {
	cases := []struct {
		name string
		keys []any
		size int
		want [][]any
	}{
		{
			name: "size 0 disables chunking",
			keys: []any{1, 2, 3},
			size: 0,
			want: [][]any{{1, 2, 3}},
		},
		{
			name: "negative size disables chunking",
			keys: []any{1, 2, 3, 4, 5},
			size: -1,
			want: [][]any{{1, 2, 3, 4, 5}},
		},
		{
			name: "len <= size returns single chunk",
			keys: []any{1, 2, 3},
			size: 5,
			want: [][]any{{1, 2, 3}},
		},
		{
			name: "exact multiple",
			keys: []any{1, 2, 3, 4},
			size: 2,
			want: [][]any{{1, 2}, {3, 4}},
		},
		{
			name: "non-exact: last chunk shorter",
			keys: []any{1, 2, 3, 4, 5},
			size: 2,
			want: [][]any{{1, 2}, {3, 4}, {5}},
		},
		{
			name: "size 1 yields one item per chunk",
			keys: []any{1, 2, 3},
			size: 1,
			want: [][]any{{1}, {2}, {3}},
		},
		{
			name: "empty input",
			keys: []any{},
			size: 10,
			want: [][]any{{}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := chunkKeys(tc.keys, tc.size)
			assert.Equal(t, tc.want, got)
		})
	}
}
