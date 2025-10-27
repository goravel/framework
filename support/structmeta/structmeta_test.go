package structmeta

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type simpleStruct struct {
	Name string `json:"name" db:"user_name"`
	Age  int    `validate:"min=1"`
}

type methodStruct struct {
	Data string
}

func (m methodStruct) ValueMethod(a int) string              { return "value" }
func (m *methodStruct) PointerMethod(b bool) (string, error) { return "ptr", nil }

type embedLevel1 struct {
	Level1 string `tag:"L1"`
}

type embedLevel2 struct {
	embedLevel1
	Level2 int `tag:"L2"`
}

type embedLevel3 struct {
	embedLevel2
	Level3 bool `tag:"L3"`
}

type embedPtrLevel1 struct {
	PtrLevel1 string `tag:"PL1"`
}
type embedWithPointer struct {
	*embedPtrLevel1
	OtherField int `tag:"other"`
}

// For max depth test (const maxEmbeddedParseDepth = 2)
type embedDepth0 struct {
	Field0 string
}

type embedDepth1 struct {
	embedDepth0 // depth 1
}

type embedDepth2 struct {
	embedDepth1 // depth 2
}

type embedDepth3 struct {
	embedDepth2 // depth > 2 (fields of embedDepth0 should NOT be parsed)
}

func TestWalkStruct_InvalidInputs(t *testing.T) {
	testCases := []struct {
		name  string
		input any
	}{
		{"Nil", nil},
		{"Int", 42},
		{"String", "not a struct"},
		{"Pointer to Int", new(int)},
		{"Slice", []int{1, 2}},
		{"Map", map[string]int{"a": 1}},
		{"Func", func() {}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			meta := WalkStruct(tc.input)
			assert.Equal(t, "", meta.Name, "Name should be empty")
			assert.Empty(t, meta.Fields, "Fields should be empty")
			assert.Empty(t, meta.Methods, "Methods should be empty")
			assert.False(t, meta.ReflectValue.IsValid(), "ReflectValue should be invalid")
		})
	}
}

func TestWalkStruct_BasicFields(t *testing.T) {
	t.Run("Struct Value", func(t *testing.T) {
		meta := WalkStruct(simpleStruct{Name: "Test", Age: 30})
		assert.Equal(t, "structmeta.simpleStruct", meta.Name)
		assert.Len(t, meta.Fields, 2)
		assert.Equal(t, reflect.Struct, meta.ReflectValue.Kind())

		// Field 0: Name
		f0 := meta.Fields[0]
		assert.Equal(t, "Name", f0.Name)
		assert.Equal(t, "string", f0.Type)
		assert.Equal(t, reflect.String, f0.Kind)
		assert.False(t, f0.Anonymous)
		assert.Equal(t, []int{0}, f0.IndexPath)
		assert.Equal(t, "name", f0.Tag.Get("json"))
		assert.Equal(t, "user_name", f0.Tag.Get("db"))

		// Field 1: Age
		f1 := meta.Fields[1]
		assert.Equal(t, "Age", f1.Name)
		assert.Equal(t, "int", f1.Type)
		assert.Equal(t, reflect.Int, f1.Kind)
		assert.False(t, f1.Anonymous)
		assert.Equal(t, []int{1}, f1.IndexPath)
		assert.Equal(t, "min=1", f1.Tag.Get("validate"))
	})

	t.Run("Struct Pointer", func(t *testing.T) {
		meta := WalkStruct(&simpleStruct{Name: "Test", Age: 30})
		assert.Equal(t, "structmeta.simpleStruct", meta.Name)
		assert.Len(t, meta.Fields, 2)
		assert.Equal(t, reflect.Ptr, meta.ReflectValue.Kind())

		// Fields should be identical to the value test
		assert.Equal(t, "Name", meta.Fields[0].Name)
		assert.Equal(t, []int{0}, meta.Fields[0].IndexPath)
		assert.Equal(t, "Age", meta.Fields[1].Name)
		assert.Equal(t, []int{1}, meta.Fields[1].IndexPath)
	})
}

func TestWalkStruct_EmbeddedFields(t *testing.T) {
	meta := WalkStruct(embedLevel3{})
	assert.Equal(t, "structmeta.embedLevel3", meta.Name)

	expectedFields := []FieldMetadata{
		{Name: "embedLevel2", Type: "structmeta.embedLevel2", Kind: reflect.Struct, Anonymous: true, IndexPath: []int{0}},
		{Name: "embedLevel1", Type: "structmeta.embedLevel1", Kind: reflect.Struct, Anonymous: true, IndexPath: []int{0, 0}},
		{Name: "Level1", Type: "string", Kind: reflect.String, Anonymous: false, IndexPath: []int{0, 0, 0}},
		{Name: "Level2", Type: "int", Kind: reflect.Int, Anonymous: false, IndexPath: []int{0, 1}},
		{Name: "Level3", Type: "bool", Kind: reflect.Bool, Anonymous: false, IndexPath: []int{1}},
	}

	assert.Len(t, meta.Fields, len(expectedFields))

	for i, expected := range expectedFields {
		actual := meta.Fields[i]
		assert.Equal(t, expected.Name, actual.Name, "Field %d Name", i)
		assert.Equal(t, expected.Type, actual.Type, "Field %d Type", i)
		assert.Equal(t, expected.Kind, actual.Kind, "Field %d Kind", i)
		assert.Equal(t, expected.Anonymous, actual.Anonymous, "Field %d Anonymous", i)
		assert.Equal(t, expected.IndexPath, actual.IndexPath, "Field %d IndexPath", i)

		if expected.Name == "Level1" {
			assert.Equal(t, "L1", actual.Tag.Get("tag"))
		}
		if expected.Name == "Level2" {
			assert.Equal(t, "L2", actual.Tag.Get("tag"))
		}
		if expected.Name == "Level3" {
			assert.Equal(t, "L3", actual.Tag.Get("tag"))
		}
	}
}

func TestWalkStruct_EmbeddedFields_Pointer(t *testing.T) {
	meta := WalkStruct(embedWithPointer{})

	expectedFields := []FieldMetadata{
		{Name: "embedPtrLevel1", Type: "*structmeta.embedPtrLevel1", Kind: reflect.Ptr, Anonymous: true, IndexPath: []int{0}},
		{Name: "PtrLevel1", Type: "string", Kind: reflect.String, Anonymous: false, IndexPath: []int{0, 0}},
		{Name: "OtherField", Type: "int", Kind: reflect.Int, Anonymous: false, IndexPath: []int{1}},
	}

	assert.Len(t, meta.Fields, len(expectedFields))

	for i, expected := range expectedFields {
		actual := meta.Fields[i]
		assert.Equal(t, expected.Name, actual.Name, "Field %d Name", i)
		assert.Equal(t, expected.Type, actual.Type, "Field %d Type", i)
		assert.Equal(t, expected.Kind, actual.Kind, "Field %d Kind", i)
		assert.Equal(t, expected.Anonymous, actual.Anonymous, "Field %d Anonymous", i)
		assert.Equal(t, expected.IndexPath, actual.IndexPath, "Field %d IndexPath", i)
	}
}

func TestWalkStruct_EmbeddedFields_MaxDepth(t *testing.T) {
	meta := WalkStruct(embedDepth3{})

	// It should parse:
	// 1. embedDepth2 (depth 0, anon)
	// 2.   -> embedDepth1 (depth 1, anon)
	// 3.     -> embedDepth0 (depth 2, anon)
	// 4.       -> STOPS. 'Field0' is at depth 3 and should NOT be included.

	expectedNames := []string{"embedDepth2", "embedDepth1", "embedDepth0"}

	assert.Len(t, meta.Fields, len(expectedNames))

	names := make([]string, len(meta.Fields))
	for i, f := range meta.Fields {
		names[i] = f.Name
	}
	assert.Equal(t, expectedNames, names)
}

func TestWalkStruct_MethodBinding_AllScenarios(t *testing.T) {
	type expectedMethod struct {
		Receiver   string
		Params     []string
		Returns    []string
		IsCallable bool // The most important check: m.ReflectValue.IsValid()
	}

	expectedMethodSet := map[string]expectedMethod{
		"ValueMethod": {
			Receiver:   "structmeta.methodStruct",
			Params:     []string{"int"},
			Returns:    []string{"string"},
			IsCallable: true,
		},
		"PointerMethod": {
			Receiver:   "*structmeta.methodStruct",
			Params:     []string{"bool"},
			Returns:    []string{"string", "error"},
			IsCallable: true,
		},
	}

	testCases := []struct {
		name     string
		input    any
		expected map[string]expectedMethod
	}{
		{
			name:  "Pointer Instance",
			input: &methodStruct{Data: "ptr"},
			expected: map[string]expectedMethod{
				"ValueMethod":   expectedMethodSet["ValueMethod"],   // Value method is callable on pointer
				"PointerMethod": expectedMethodSet["PointerMethod"], // Pointer method is callable on pointer
			},
		},
		{
			name:  "Value Instance",
			input: methodStruct{Data: "val"},
			expected: map[string]expectedMethod{
				"ValueMethod":   expectedMethodSet["ValueMethod"],   // Value method is callable on value
				"PointerMethod": expectedMethodSet["PointerMethod"], // **Pointer method is callable on value (via tmp copy)**
			},
		},
		{
			name:  "Value in Interface",
			input: any(methodStruct{Data: "val-iface"}),
			expected: map[string]expectedMethod{
				"ValueMethod":   expectedMethodSet["ValueMethod"],
				"PointerMethod": expectedMethodSet["PointerMethod"], // **Pointer method is callable (via tmp copy)**
			},
		},
		{
			name:  "Pointer in Interface",
			input: any(&methodStruct{Data: "ptr-iface"}),
			expected: map[string]expectedMethod{
				"ValueMethod":   expectedMethodSet["ValueMethod"],
				"PointerMethod": expectedMethodSet["PointerMethod"],
			},
		},
		{
			name:  "Nil Pointer Instance",
			input: (*methodStruct)(nil),
			expected: map[string]expectedMethod{
				"ValueMethod":   {Receiver: "structmeta.methodStruct", Params: []string{"int"}, Returns: []string{"string"}, IsCallable: false},
				"PointerMethod": {Receiver: "*structmeta.methodStruct", Params: []string{"bool"}, Returns: []string{"string", "error"}, IsCallable: false},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			meta := WalkStruct(tc.input)

			assert.Equal(t, "structmeta.methodStruct", meta.Name)
			assert.Len(t, meta.Methods, len(tc.expected))

			foundMethods := make(map[string]MethodMetadata)
			for _, m := range meta.Methods {
				foundMethods[m.Name] = m
			}

			for name, expected := range tc.expected {
				actual, ok := foundMethods[name]
				assert.True(t, ok, "Method '%s' was not found", name)
				if !ok {
					continue
				}

				assert.Equal(t, expected.Receiver, actual.Receiver, "Method '%s' Receiver", name)
				assert.Equal(t, expected.Params, actual.Parameters, "Method '%s' Parameters", name)
				assert.Equal(t, expected.Returns, actual.Returns, "Method '%s' Returns", name)
				assert.Equal(t, expected.IsCallable, actual.ReflectValue.IsValid(), "Method '%s' IsCallable", name)
			}
		})
	}
}
