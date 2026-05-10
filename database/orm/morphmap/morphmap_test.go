package morphmap

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type morphmapPost struct {
	ID uint
}

type morphmapVideo struct {
	ID uint
}

type morphmapUser struct{}

func (morphmapUser) MorphClass() string { return "user" }

type morphmapTag struct{}

func (*morphmapTag) MorphClass() string { return "tag" }

type morphmapEmpty struct{}

func (morphmapEmpty) MorphClass() string { return "" }

func TestRegister_AndLookup(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		alias     string
		wantType  reflect.Type
		wantNil   bool
		modelLook any
		wantAlias string
		wantOk    bool
	}{
		{
			name: "registered alias yields fresh pointer to model type",
			setup: func() {
				Reset()
				Register(map[string]any{"post": &morphmapPost{}})
			},
			alias:     "post",
			wantType:  reflect.TypeOf(&morphmapPost{}),
			modelLook: &morphmapPost{},
			wantAlias: "post",
			wantOk:    true,
		},
		{
			name: "value-type sample is normalised to elem type",
			setup: func() {
				Reset()
				Register(map[string]any{"video": morphmapVideo{}})
			},
			alias:     "video",
			wantType:  reflect.TypeOf(&morphmapVideo{}),
			modelLook: &morphmapVideo{},
			wantAlias: "video",
			wantOk:    true,
		},
		{
			name:    "unregistered alias returns nil",
			setup:   Reset,
			alias:   "missing",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got := Find(tt.alias)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			assert.Equal(t, tt.wantType, reflect.TypeOf(got))
			alias, ok := AliasOf(tt.modelLook)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantAlias, alias)
		})
	}
}

func TestRegister_MergeAndReplace(t *testing.T) {
	Reset()
	Register(map[string]any{"post": &morphmapPost{}})
	Register(map[string]any{"video": &morphmapVideo{}}) // merge
	assert.NotNil(t, Find("post"))
	assert.NotNil(t, Find("video"))

	Register(map[string]any{"user": &morphmapUser{}}, false) // replace
	assert.Nil(t, Find("post"))
	assert.Nil(t, Find("video"))
	assert.NotNil(t, Find("user"))
}

func TestRegister_LaterWriteWins(t *testing.T) {
	Reset()
	Register(map[string]any{"post": &morphmapPost{}})
	Register(map[string]any{"post": &morphmapVideo{}}) // overwrite under same alias
	got := Find("post")
	assert.Equal(t, reflect.TypeOf(&morphmapVideo{}), reflect.TypeOf(got))

	// AliasOf should also rebind: Post no longer maps to "post"; Video does.
	_, ok := AliasOf(&morphmapPost{})
	assert.False(t, ok)

	alias, ok := AliasOf(&morphmapVideo{})
	assert.True(t, ok)
	assert.Equal(t, "post", alias)
}

func TestRegister_RebindOnConflictingType(t *testing.T) {
	// Same type registered under two different aliases — only the later alias survives.
	Reset()
	Register(map[string]any{"first": &morphmapPost{}})
	Register(map[string]any{"second": &morphmapPost{}})

	assert.Nil(t, Find("first"))
	got := Find("second")
	assert.Equal(t, reflect.TypeOf(&morphmapPost{}), reflect.TypeOf(got))

	alias, ok := AliasOf(&morphmapPost{})
	assert.True(t, ok)
	assert.Equal(t, "second", alias)
}

func TestMorphValue_PriorityOrder(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		input     any
		wantValue string
		wantOk    bool
	}{
		{
			name: "MorphClass method takes precedence over registry",
			setup: func() {
				Reset()
				Register(map[string]any{"registered_user": &morphmapUser{}})
			},
			input:     &morphmapUser{},
			wantValue: "user", // from MorphClass(), not "registered_user"
			wantOk:    true,
		},
		{
			name: "MorphClass works on pointer-receiver method via value caller",
			setup: func() {
				Reset()
			},
			input:     morphmapTag{}, // value, but MorphClass has pointer receiver
			wantValue: "tag",
			wantOk:    true,
		},
		{
			name: "MorphClass works on pointer caller too",
			setup: func() {
				Reset()
			},
			input:     &morphmapTag{},
			wantValue: "tag",
			wantOk:    true,
		},
		{
			name: "empty MorphClass falls through to registry",
			setup: func() {
				Reset()
				Register(map[string]any{"fallback": &morphmapEmpty{}})
			},
			input:     &morphmapEmpty{},
			wantValue: "fallback",
			wantOk:    true,
		},
		{
			name: "registry hit when no MorphClass",
			setup: func() {
				Reset()
				Register(map[string]any{"post": &morphmapPost{}})
			},
			input:     &morphmapPost{},
			wantValue: "post",
			wantOk:    true,
		},
		{
			name:   "no MorphClass and no registry entry yields not-found",
			setup:  Reset,
			input:  &morphmapPost{},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			value, ok := MorphValue(tt.input)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestAll_Snapshot(t *testing.T) {
	Reset()
	Register(map[string]any{
		"post":  &morphmapPost{},
		"video": &morphmapVideo{},
	})
	snap := All()
	assert.Equal(t, 2, len(snap))
	assert.Equal(t, reflect.TypeOf(morphmapPost{}), snap["post"])
	assert.Equal(t, reflect.TypeOf(morphmapVideo{}), snap["video"])

	// Mutating the snapshot must not affect the registry.
	delete(snap, "post")
	assert.NotNil(t, Find("post"))
}
