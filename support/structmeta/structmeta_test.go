package structmeta

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type simpleStruct struct {
	Name string `json:"name"`
	Age  int    `validate:"min=1"`
}

type methodStruct struct{}

func (m methodStruct) ValueMethod() string       { return "value" }
func (m *methodStruct) PointerMethod(int) string { return "ptr" }

type embedLevel1 struct {
	Level1 string `tag:"L1"`
}

type embedLevel2 struct {
	embedLevel1
	Level2 int `tag:"L2"`
}

type embedLevel3 struct {
	embedLevel2
	Level3 bool
}

func TestWalkStruct_BasicFields(t *testing.T) {
	meta := WalkStruct(simpleStruct{})

	assert.Equal(t, "structmeta.simpleStruct", meta.Name)
	assert.Len(t, meta.Fields, 2)

	field := meta.Fields[0]
	assert.Equal(t, "Name", field.Name)
	assert.Equal(t, "string", field.Type)
	assert.Equal(t, reflect.String, field.Kind)
	assert.Equal(t, "name", field.Tag.Get("json"))

	field = meta.Fields[1]
	assert.Equal(t, "Age", field.Name)
	assert.Equal(t, "int", field.Type)
	assert.Equal(t, reflect.Int, field.Kind)
	assert.Equal(t, "min=1", field.Tag.Get("validate"))
}

func TestWalkStruct_MethodDetection(t *testing.T) {
	meta := WalkStruct(&methodStruct{})

	methods := map[string]struct {
		Receiver   string
		Params     []string
		Returns    []string
		IsCallable bool
	}{
		"ValueMethod": {
			Receiver:   "structmeta.methodStruct",
			Params:     nil,
			Returns:    []string{"string"},
			IsCallable: true,
		},
		"PointerMethod": {
			Receiver:   "*structmeta.methodStruct",
			Params:     []string{"int"},
			Returns:    []string{"string"},
			IsCallable: true,
		},
	}

	for _, m := range meta.Methods {
		expect, ok := methods[m.Name]
		if !ok {
			t.Errorf("unexpected method: %s", m.Name)
			continue
		}
		assert.Equal(t, expect.Receiver, m.Receiver)
		assert.Equal(t, expect.Params, m.Parameters)
		assert.Equal(t, expect.Returns, m.Returns)
		assert.Equal(t, expect.IsCallable, m.ReflectValue.IsValid(), "%s callable", m.Name)
		delete(methods, m.Name)
	}

	if len(methods) > 0 {
		t.Errorf("missing expected methods: %+v", methods)
	}
}

func TestWalkStruct_EmbeddedFields(t *testing.T) {
	meta := WalkStruct(embedLevel3{})

	names := make([]string, len(meta.Fields))
	for i, f := range meta.Fields {
		names[i] = f.Name
	}

	expected := []string{"embedLevel2", "embedLevel1", "Level1", "Level2", "Level3"}
	assert.Equal(t, expected, names)
}

func TestWalkStruct_NilPointer(t *testing.T) {
	var s *simpleStruct
	meta := WalkStruct(s)
	assert.Equal(t, "structmeta.simpleStruct", meta.Name)
	assert.Len(t, meta.Fields, 2)
}

func TestWalkStruct_NonStructInput(t *testing.T) {
	assert.Empty(t, WalkStruct(nil).Fields)
	assert.Empty(t, WalkStruct(42).Fields)
	assert.Empty(t, WalkStruct("not a struct").Fields)
}

func TestNewTagMetadata(t *testing.T) {
	tag := reflect.StructTag(`json:"id,omitempty" validate:"min=1,max=10"`)
	meta := NewTagMetadata(tag)

	assert.True(t, meta.HasKey("json"))
	assert.True(t, meta.HasKey("validate"))
	assert.False(t, meta.HasKey("nonexistent"))

	assert.Equal(t, "id,omitempty", meta.Get("json"))
	assert.Equal(t, []string{"id", "omitempty"}, meta.GetParts("json"))

	assert.Equal(t, "min=1,max=10", meta.Get("validate"))
	assert.Equal(t, []string{"min=1", "max=10"}, meta.GetParts("validate"))
}
