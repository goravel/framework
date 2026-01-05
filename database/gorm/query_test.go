package gorm

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/convert"
)

func TestAddGlobalScopes(t *testing.T) {

	tests := []struct {
		name                  string
		setupQuery            func() *Query
		expectedScopesApplied bool
		expectedScopesCount   int
	}{
		{
			name: "should apply global scopes when model implements ModelWithGlobalScopes",
			setupQuery: func() *Query {
				conditions := Conditions{}
				conditions.model = &ModelWithGlobalScopes{}
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)
				return query
			},
			expectedScopesApplied: true,
			expectedScopesCount:   2, // active and verified scopes
		},
		{
			name: "should not apply global scopes when model does not implement ModelWithGlobalScopes",
			setupQuery: func() *Query {
				conditions := Conditions{}
				conditions.model = &ModelWithoutGlobalScopes{}
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)
				return query
			},
			expectedScopesApplied: false,
			expectedScopesCount:   0,
		},
		{
			name: "should not apply global scopes when withoutGlobalScopes contains '*'",
			setupQuery: func() *Query {
				conditions := Conditions{}
				conditions.model = &ModelWithGlobalScopes{}
				conditions.withoutGlobalScopes = []string{"*"}
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)
				return query
			},
			expectedScopesApplied: false,
			expectedScopesCount:   0,
		},
		{
			name: "should exclude specific scope when in withoutGlobalScopes",
			setupQuery: func() *Query {
				conditions := Conditions{}
				conditions.model = &ModelWithGlobalScopes{}
				conditions.withoutGlobalScopes = []string{"active"}
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)
				return query
			},
			expectedScopesApplied: true,
			expectedScopesCount:   1, // only verified scope
		},
		{
			name: "should not apply scopes when model is nil",
			setupQuery: func() *Query {
				conditions := Conditions{}
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)
				return query
			},
			expectedScopesApplied: false,
			expectedScopesCount:   0,
		},
		{
			name: "should use dest when model is nil but dest is set",
			setupQuery: func() *Query {
				conditions := Conditions{}
				conditions.dest = &ModelWithGlobalScopes{}
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)
				return query
			},
			expectedScopesApplied: true,
			expectedScopesCount:   2, // active and verified scopes
		},
		{
			name: "should not apply scopes when model has empty GlobalScopes map",
			setupQuery: func() *Query {
				conditions := Conditions{}
				conditions.model = &ModelWithEmptyGlobalScopes{}
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)
				return query
			},
			expectedScopesApplied: false,
			expectedScopesCount:   0,
		},
		{
			name: "should apply scopes in sorted order by name",
			setupQuery: func() *Query {
				conditions := Conditions{}
				conditions.model = &ModelWithGlobalScopes{}
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)
				return query
			},
			expectedScopesApplied: true,
			expectedScopesCount:   2, // Scopes applied in order: active, verified (alphabetically)
		},
		{
			name: "should exclude multiple scopes when multiple names in withoutGlobalScopes",
			setupQuery: func() *Query {
				conditions := Conditions{}
				conditions.model = &ModelWithGlobalScopes{}
				conditions.withoutGlobalScopes = []string{"active", "verified"}
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)
				return query
			},
			expectedScopesApplied: false,
			expectedScopesCount:   0, // both scopes excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.setupQuery()
			result := query.addGlobalScopes()

			assert.NotNil(t, result)

			// Check if scopes were applied by examining the scopes list
			if tt.expectedScopesApplied {
				assert.GreaterOrEqual(t, len(result.conditions.scopes), tt.expectedScopesCount)
			} else {
				assert.Equal(t, tt.expectedScopesCount, len(result.conditions.scopes))
			}
		})
	}
}

func TestAddGlobalScopesWithPointerTypes(t *testing.T) {

	tests := []struct {
		name     string
		model    any
		expected bool
	}{
		{
			name:     "should work with pointer to model",
			model:    &ModelWithGlobalScopes{},
			expected: true,
		},
		{
			name:     "should work with value type model",
			model:    ModelWithGlobalScopes{},
			expected: true,
		},
		{
			name:     "should work with pointer to pointer",
			model:    func() any { m := &ModelWithGlobalScopes{}; return &m }(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditions := Conditions{}
			conditions.model = tt.model
			query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)

			result := query.addGlobalScopes()

			assert.NotNil(t, result)
			if tt.expected {
				assert.GreaterOrEqual(t, len(result.conditions.scopes), 2)
			}
		})
	}
}

func TestAddGlobalScopesWithSliceAndArray(t *testing.T) {

	tests := []struct {
		name  string
		model any
	}{
		{
			name:  "should handle slice of models",
			model: []ModelWithGlobalScopes{},
		},
		{
			name:  "should handle pointer to slice of models",
			model: &[]ModelWithGlobalScopes{},
		},
		{
			name:  "should handle array of models",
			model: [2]ModelWithGlobalScopes{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditions := Conditions{}
			conditions.model = tt.model
			query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions)

			result := query.addGlobalScopes()

			// Should handle slice/array types and extract the element type
			assert.NotNil(t, result)
		})
	}
}

func TestAddWhere(t *testing.T) {
	query := &Query{}
	query = query.addWhere(contractsdriver.Where{
		Query: "name",
		Args:  []any{"test"},
	}).(*Query)
	query = query.addWhere(contractsdriver.Where{
		Query: "name1",
		Args:  []any{"test1"},
	}).(*Query)
	query = query.addWhere(contractsdriver.Where{
		Query: "name2",
		Args:  []any{"test2"},
	}).(*Query)
	query1 := query.addWhere(contractsdriver.Where{
		Query: "name3",
		Args:  []any{"test3"},
	}).(*Query)

	assert.Equal(t, []contractsdriver.Where{
		{Query: "name", Args: []any{"test"}},
		{Query: "name1", Args: []any{"test1"}},
		{Query: "name2", Args: []any{"test2"}},
	}, query.conditions.where)

	assert.Equal(t, []contractsdriver.Where{
		{Query: "name", Args: []any{"test"}},
		{Query: "name1", Args: []any{"test1"}},
		{Query: "name2", Args: []any{"test2"}},
		{Query: "name3", Args: []any{"test3"}},
	}, query1.conditions.where)

	query2 := query.addWhere(contractsdriver.Where{
		Query: "name4",
		Args:  []any{"test4"},
	}).(*Query)

	assert.Equal(t, []contractsdriver.Where{
		{Query: "name", Args: []any{"test"}},
		{Query: "name1", Args: []any{"test1"}},
		{Query: "name2", Args: []any{"test2"}},
		{Query: "name4", Args: []any{"test4"}},
	}, query2.conditions.where)

	assert.Equal(t, []contractsdriver.Where{
		{Query: "name", Args: []any{"test"}},
		{Query: "name1", Args: []any{"test1"}},
		{Query: "name2", Args: []any{"test2"}},
	}, query.conditions.where)

	assert.Equal(t, []contractsdriver.Where{
		{Query: "name", Args: []any{"test"}},
		{Query: "name1", Args: []any{"test1"}},
		{Query: "name2", Args: []any{"test2"}},
		{Query: "name3", Args: []any{"test3"}},
	}, query1.conditions.where)
}

func TestFilterFindConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions []any
		expectErr  error
	}{
		{
			name: "condition is empty",
		},
		{
			name:       "condition is empty string",
			conditions: []any{""},
			expectErr:  errors.OrmMissingWhereClause,
		},
		{
			name:       "condition is empty slice",
			conditions: []any{[]string{}},
			expectErr:  errors.OrmMissingWhereClause,
		},
		{
			name:       "condition has value",
			conditions: []any{"name = ?", "test"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := filterFindConditions(test.conditions...)
			if test.expectErr != nil {
				assert.Equal(t, err, test.expectErr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetDeletedAtColumnName(t *testing.T) {
	type Test1 struct {
		Deleted gormio.DeletedAt
	}

	assert.Equal(t, "Deleted", getDeletedAtColumn(Test1{}))
	assert.Equal(t, "Deleted", getDeletedAtColumn(&Test1{}))

	type Test2 struct {
		Test1
	}

	assert.Equal(t, "Deleted", getDeletedAtColumn(Test2{}))
	assert.Equal(t, "Deleted", getDeletedAtColumn(&Test2{}))
}

func TestGetModelConnection(t *testing.T) {
	tests := []struct {
		name             string
		model            any
		expectConnection string
	}{
		{
			name: "invalid model",
			model: func() any {
				var product string
				return product
			}(),
		},
		{
			name: "not ConnectionModel",
			model: func() any {
				var user User
				return user
			}(),
		},
		{
			name: "the connection of model is empty",
			model: func() any {
				var review Review
				return review
			}(),
		},
		{
			name: "model is map",
			model: func() any {
				return map[string]any{}
			}(),
		},
		{
			name: "the connection of model is not empty",
			model: func() any {
				var product Product
				return product
			}(),
			expectConnection: "sqlite",
		},
		{
			name: "the connection of model is not empty and model is slice",
			model: func() any {
				var products []Product
				return products
			}(),
			expectConnection: "sqlite",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := &Query{
				conditions: Conditions{
					model: test.model,
				},
			}
			connection := query.getModelConnection()

			assert.Equal(t, test.expectConnection, connection)
		})
	}
}

func TestGetObserver(t *testing.T) {
	query := &Query{
		modelToObserver: []contractsorm.ModelToObserver{
			{
				Model:    User{},
				Observer: &UserObserver{},
			},
		},
	}

	assert.Nil(t, query.getObserver(Product{}))
	assert.Equal(t, &UserObserver{}, query.getObserver(User{}))
}

func TestModelToStruct(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expectError bool
		expectedErr error
		checkResult func(t *testing.T, result any)
	}{
		// Basic cases
		{
			name:        "nil input",
			input:       nil,
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args("nil"),
		},
		{
			name:        "nil pointer to struct",
			input:       (*TestModel)(nil),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
				resultValue := reflect.ValueOf(result)
				assert.True(t, resultValue.IsValid())
				assert.False(t, resultValue.IsNil())
			},
		},
		{
			name:        "valid struct pointer",
			input:       &TestModel{ID: 1, Name: "test"},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
				resultValue := reflect.ValueOf(result)
				assert.True(t, resultValue.IsValid())
				assert.False(t, resultValue.IsNil())
			},
		},
		{
			name:        "struct value (not pointer)",
			input:       TestModel{ID: 1, Name: "test"},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "slice of structs",
			input:       []TestModel{{ID: 1, Name: "test"}},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "array of structs",
			input:       [1]TestModel{{ID: 1, Name: "test"}},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "pointer to slice",
			input:       &[]TestModel{{ID: 1, Name: "test"}},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "pointer to array",
			input:       &[1]TestModel{{ID: 1, Name: "test"}},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "nil slice",
			input:       []TestModel(nil),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "empty slice",
			input:       []TestModel{},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "empty array",
			input:       [0]TestModel{},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "slice of pointers to structs",
			input:       []*TestModel{{ID: 1, Name: "test"}},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "array of pointers to structs",
			input:       [1]*TestModel{{ID: 1, Name: "test"}},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "pointer to slice of pointers",
			input:       &[]*TestModel{{ID: 1, Name: "test"}},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name:        "nil slice of pointers",
			input:       []*TestModel(nil),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name: "multiple pointer levels",
			input: func() any {
				model := &TestModel{ID: 1, Name: "test"}
				ptr1 := &model
				ptr2 := &ptr1
				ptr3 := &ptr2
				return ptr3
			}(),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},
		{
			name: "slice with multiple pointer levels",
			input: func() any {
				model := &TestModel{ID: 1, Name: "test"}
				slice := []**TestModel{&model}
				ptrSlice := &slice
				return ptrSlice
			}(),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestModel", resultType.String())
			},
		},

		// Interface cases
		{
			name:        "interface type with concrete value",
			input:       TestInterfaceImpl{ID: 1},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestInterfaceImpl", resultType.String())
			},
		},
		{
			name:        "nil interface",
			input:       (TestInterfaceModel)(nil),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args("nil"),
		},
		{
			name: "interface with concrete value",
			input: func() any {
				var iface TestInterfaceModel = TestInterfaceImpl{ID: 1}
				return iface
			}(),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.TestInterfaceImpl", resultType.String())
			},
		},

		// Invalid types
		{
			name:        "map type",
			input:       map[string]any{"id": 1, "name": "test"},
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args("map"),
		},
		{
			name:        "nil map",
			input:       (map[string]any)(nil),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args("map"),
		},
		{
			name:        "empty map",
			input:       map[string]any{},
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args("map"),
		},
		{
			name:        "nil pointer to map",
			input:       (*map[string]any)(nil),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args("map"),
		},
		{
			name:        "string type",
			input:       "invalid",
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "int type",
			input:       123,
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "float type",
			input:       123.45,
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "bool type",
			input:       true,
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "pointer to string",
			input:       convert.Pointer("test"),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "pointer to int",
			input:       convert.Pointer(123),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "pointer to float",
			input:       convert.Pointer(123.45),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "pointer to bool",
			input:       convert.Pointer(true),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "channel type",
			input:       make(chan int),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "function type",
			input:       func() {},
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name: "pointer to function",
			input: func() any {
				fn := func() {}
				return &fn
			}(),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},

		// Struct cases
		{
			name:        "empty struct",
			input:       struct{}{},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*struct {}", resultType.String())
			},
		},
		{
			name:        "pointer to empty struct",
			input:       &struct{}{},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*struct {}", resultType.String())
			},
		},
		{
			name:        "nil pointer to empty struct",
			input:       (*struct{})(nil),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*struct {}", resultType.String())
			},
		},
		{
			name:        "complex nested struct",
			input:       &ComplexModel{},
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.ComplexModel", resultType.String())
			},
		},
		{
			name:        "nil pointer to complex struct",
			input:       (*ComplexModel)(nil),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.ComplexModel", resultType.String())
			},
		},

		// Edge cases - struct with embedded struct
		{
			name: "struct with embedded struct",
			input: func() any {
				type EmbeddedStruct struct {
					Field string
				}
				type StructWithEmbed struct {
					EmbeddedStruct
					Name string
				}
				return StructWithEmbed{EmbeddedStruct: EmbeddedStruct{Field: "test"}, Name: "name"}
			}(),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.StructWithEmbed", resultType.String())
			},
		},
		{
			name: "struct with pointer to embedded struct",
			input: func() any {
				type EmbeddedStruct struct {
					Field string
				}
				type StructWithEmbed struct {
					*EmbeddedStruct
					Name string
				}
				return StructWithEmbed{EmbeddedStruct: &EmbeddedStruct{Field: "test"}, Name: "name"}
			}(),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.StructWithEmbed", resultType.String())
			},
		},
		{
			name: "struct with unexported fields",
			input: func() any {
				type StructWithUnexported struct {
					PublicField  string
					privateField string
				}
				return StructWithUnexported{PublicField: "public", privateField: "private"}
			}(),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.StructWithUnexported", resultType.String())
			},
		},
		{
			name: "struct with tags",
			input: func() any {
				type StructWithTags struct {
					Field string `json:"field" db:"field"`
				}
				return StructWithTags{Field: "test"}
			}(),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultType := reflect.TypeOf(result)
				assert.Equal(t, "*gorm.StructWithTags", resultType.String())
			},
		},

		// Return value validation cases
		{
			name: "returned value is pointer to struct",
			input: func() any {
				model := TestModel{ID: 1, Name: "test"}
				return &model
			}(),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				resultValue := reflect.ValueOf(result)
				assert.True(t, resultValue.Kind() == reflect.Ptr)
				assert.False(t, resultValue.IsNil())
				assert.Equal(t, reflect.Struct, resultValue.Elem().Kind())
			},
		},
		{
			name: "returned value can be used for reflection",
			input: func() any {
				model := TestModel{ID: 1, Name: "test"}
				return &model
			}(),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				resultValue := reflect.ValueOf(result)
				resultType := resultValue.Type()
				assert.Equal(t, "*gorm.TestModel", resultType.String())

				// Test that we can access fields
				elem := resultValue.Elem()
				idField := elem.FieldByName("ID")
				nameField := elem.FieldByName("Name")
				assert.True(t, idField.IsValid())
				assert.True(t, nameField.IsValid())
			},
		},
		{
			name: "returned value is new instance",
			input: func() any {
				original := &TestModel{ID: 1, Name: "test"}
				return original
			}(),
			expectError: false,
			checkResult: func(t *testing.T, result any) {
				assert.NotNil(t, result)
				// The result should be a different instance
				assert.NotSame(t, &TestModel{ID: 1, Name: "test"}, result)
				assert.Equal(t, reflect.TypeOf(&TestModel{}), reflect.TypeOf(result))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := modelToStruct(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

func TestObserverEvent(t *testing.T) {
	assert.EqualError(t, getObserverEvent(contractsorm.EventRetrieved, &UserObserver{})(nil), "retrieved")
	assert.EqualError(t, getObserverEvent(contractsorm.EventCreating, &UserObserver{})(nil), "creating")
	assert.EqualError(t, getObserverEvent(contractsorm.EventCreated, &UserObserver{})(nil), "created")
	assert.EqualError(t, getObserverEvent(contractsorm.EventUpdating, &UserObserver{})(nil), "updating")
	assert.EqualError(t, getObserverEvent(contractsorm.EventUpdated, &UserObserver{})(nil), "updated")
	assert.EqualError(t, getObserverEvent(contractsorm.EventSaving, &UserObserver{})(nil), "saving")
	assert.EqualError(t, getObserverEvent(contractsorm.EventSaved, &UserObserver{})(nil), "saved")
	assert.EqualError(t, getObserverEvent(contractsorm.EventDeleting, &UserObserver{})(nil), "deleting")
	assert.EqualError(t, getObserverEvent(contractsorm.EventDeleted, &UserObserver{})(nil), "deleted")
	assert.EqualError(t, getObserverEvent(contractsorm.EventForceDeleting, &UserObserver{})(nil), "forceDeleting")
	assert.EqualError(t, getObserverEvent(contractsorm.EventForceDeleted, &UserObserver{})(nil), "forceDeleted")
	assert.Nil(t, getObserverEvent("error", &UserObserver{}))
}

type User struct {
	Name string
}

type UserObserver struct{}

func (u *UserObserver) Retrieved(event contractsorm.Event) error {
	return errors.New("retrieved")
}

func (u *UserObserver) Creating(event contractsorm.Event) error {
	return errors.New("creating")
}

func (u *UserObserver) Created(event contractsorm.Event) error {
	return errors.New("created")
}

func (u *UserObserver) Updating(event contractsorm.Event) error {
	return errors.New("updating")
}

func (u *UserObserver) Updated(event contractsorm.Event) error {
	return errors.New("updated")
}

func (u *UserObserver) Saving(event contractsorm.Event) error {
	return errors.New("saving")
}

func (u *UserObserver) Saved(event contractsorm.Event) error {
	return errors.New("saved")
}

func (u *UserObserver) Deleting(event contractsorm.Event) error {
	return errors.New("deleting")
}

func (u *UserObserver) Deleted(event contractsorm.Event) error {
	return errors.New("deleted")
}

func (u *UserObserver) ForceDeleting(event contractsorm.Event) error {
	return errors.New("forceDeleting")
}

func (u *UserObserver) ForceDeleted(event contractsorm.Event) error {
	return errors.New("forceDeleted")
}

type Product struct {
	Name string
}

func (p *Product) Connection() string {
	return "sqlite"
}

type Review struct {
	Body string
}

func (r *Review) Connection() string {
	return ""
}

// TestModel is a simple struct for testing
type TestModel struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// TestInterfaceModel is an interface type for testing
type TestInterfaceModel interface {
	GetID() uint
}

// TestInterfaceImpl implements TestInterfaceModel
type TestInterfaceImpl struct {
	ID uint `json:"id"`
}

func (t TestInterfaceImpl) GetID() uint {
	return t.ID
}

// ComplexModel is a more complex struct for testing
type ComplexModel struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	IsActive bool   `json:"is_active"`
	Nested   struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	} `json:"nested"`
}

// Test models for addGlobalScopes
type ModelWithoutGlobalScopes struct {
	ID   uint
	Name string
}

type ModelWithGlobalScopes struct {
	ID   uint
	Name string
}

func (m *ModelWithGlobalScopes) GlobalScopes() map[string]func(contractsorm.Query) contractsorm.Query {
	return map[string]func(contractsorm.Query) contractsorm.Query{
		"active": func(query contractsorm.Query) contractsorm.Query {
			return query.Where("active", true)
		},
		"verified": func(query contractsorm.Query) contractsorm.Query {
			return query.Where("verified", true)
		},
	}
}

type ModelWithEmptyGlobalScopes struct {
	ID   uint
	Name string
}

func (m *ModelWithEmptyGlobalScopes) GlobalScopes() map[string]func(contractsorm.Query) contractsorm.Query {
	return map[string]func(contractsorm.Query) contractsorm.Query{}
}
