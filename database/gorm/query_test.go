package gorm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gormio "gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormtests "gorm.io/gorm/utils/tests"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
	mocksdriver "github.com/goravel/framework/mocks/database/driver"
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
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)
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
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)
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
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)
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
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)
				return query
			},
			expectedScopesApplied: true,
			expectedScopesCount:   1, // only verified scope
		},
		{
			name: "should not apply scopes when model is nil",
			setupQuery: func() *Query {
				conditions := Conditions{}
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)
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
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)
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
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)
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
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)
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
				query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)
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
			query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)

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
			query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, &conditions, nil)

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

func TestCursor(t *testing.T) {
	// The consumer receives one row and cancels the context. When it is canceled, the producer is blocked
	// sending the third row, and the second one is in the buffer: the producer replaces it with the error.
	cancelAfterFirstRow := func(t *testing.T) (chan contractsdb.Row, *cursorDriver, func()) {
		ctx, cancel := context.WithCancel(context.Background())
		query, fake, closeDB := newCursorQuery(t, ctx, 3)

		ch := query.Cursor()
		first := <-ch
		assert.NoError(t, first.Err())

		synctest.Wait()
		cancel()
		synctest.Wait()

		return ch, fake, closeDB
	}

	t.Run("stop reading after canceling the context", func(t *testing.T) {
		// synctest fails if the producer goroutine is still blocked when the test ends.
		synctest.Test(t, func(t *testing.T) {
			_, fake, closeDB := cancelAfterFirstRow(t)
			defer closeDB()

			assert.True(t, fake.closed.Load())
		})
	})

	t.Run("keep reading after canceling the context", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ch, fake, closeDB := cancelAfterFirstRow(t)
			defer closeDB()

			var errs []error
			for row := range ch {
				errs = append(errs, row.Err())
			}

			assert.Equal(t, []error{context.Canceled}, errs)
			assert.True(t, fake.closed.Load())
		})
	})
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

func TestIncrementAndDecrement(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	var (
		sql  string
		vars []any
	)
	assert.NoError(t, db.Callback().Update().After("gorm:update").Register("test:capture", func(tx *gormio.DB) {
		sql = tx.Statement.SQL.String()
		vars = tx.Statement.Vars
	}))

	newQuery := func(t *testing.T) contractsorm.Query {
		mockGrammar := mocksdriver.NewGrammar(t)
		mockGrammar.EXPECT().CompileJsonColumnsUpdate(mock.Anything).RunAndReturn(func(values map[string]any) (map[string]any, error) {
			return values, nil
		}).Maybe()

		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, mockGrammar, nil, nil, nil, nil)
	}

	tests := []struct {
		name        string
		run         func(query contractsorm.Query) error
		expectSql   string
		expectVars  []any
		expectError error
	}{
		{
			name: "Increment with default amount",
			run: func(query contractsorm.Query) error {
				return query.Table("users").Where("id", 1).Increment("votes")
			},
			expectSql:  "UPDATE `users` SET `votes`=`votes` + ? WHERE `id` = ?",
			expectVars: []any{uint64(1), 1},
		},
		{
			name: "Increment with amount",
			run: func(query contractsorm.Query) error {
				return query.Table("users").Where("id", 1).Increment("votes", 5)
			},
			expectSql:  "UPDATE `users` SET `votes`=`votes` + ? WHERE `id` = ?",
			expectVars: []any{uint64(5), 1},
		},
		{
			name: "Decrement with default amount",
			run: func(query contractsorm.Query) error {
				return query.Table("users").Where("id", 1).Decrement("votes")
			},
			expectSql:  "UPDATE `users` SET `votes`=`votes` - ? WHERE `id` = ?",
			expectVars: []any{uint64(1), 1},
		},
		{
			name: "Decrement with amount",
			run: func(query contractsorm.Query) error {
				return query.Table("users").Where("id", 1).Decrement("votes", 3)
			},
			expectSql:  "UPDATE `users` SET `votes`=`votes` - ? WHERE `id` = ?",
			expectVars: []any{uint64(3), 1},
		},
		{
			name: "Increment with qualified column",
			run: func(query contractsorm.Query) error {
				return query.Table("users").Where("id", 1).Increment("users.votes")
			},
			expectSql:  "UPDATE `users` SET `users`.`votes`=`users`.`votes` + ? WHERE `id` = ?",
			expectVars: []any{uint64(1), 1},
		},
		{
			name: "Increment rejects JSON path columns",
			run: func(query contractsorm.Query) error {
				return query.Table("users").Where("id", 1).Increment("data->views")
			},
			expectError: errors.OrmQueryIncrementJsonColumn.Args("data->views"),
		},
		{
			name: "Decrement rejects JSON path columns",
			run: func(query contractsorm.Query) error {
				return query.Table("users").Where("id", 1).Decrement("data->stats->views", 2)
			},
			expectError: errors.OrmQueryIncrementJsonColumn.Args("data->stats->views"),
		},
		{
			name: "Increment applies global scopes and soft deletes",
			run: func(query contractsorm.Query) error {
				return query.Model(&IncrementModel{}).Where("id", 1).Increment("votes")
			},
			expectSql:  "UPDATE `increment_models` SET `votes`=`votes` + ? WHERE `id` = ? AND `active` = ? AND `increment_models`.`deleted_at` IS NULL",
			expectVars: []any{uint64(1), 1, true},
		},
		{
			name: "Increment does not apply global scopes when they are disabled",
			run: func(query contractsorm.Query) error {
				return query.Model(&IncrementModel{}).WithoutGlobalScopes().WithTrashed().Where("id", 1).Increment("votes")
			},
			expectSql:  "UPDATE `increment_models` SET `votes`=`votes` + ? WHERE `id` = ?",
			expectVars: []any{uint64(1), 1},
		},
		{
			name: "Increment fires the updating event on a model with primary key",
			run: func(query contractsorm.Query) error {
				return query.Model(&IncrementEventModel{ID: 1}).Increment("votes")
			},
			expectError: errors.New("updating"),
		},
		{
			name: "Increment does not fire events on a model without primary key",
			run: func(query contractsorm.Query) error {
				return query.Model(&IncrementEventModel{}).Where("id", 1).Increment("votes")
			},
			expectSql:  "UPDATE `increment_event_models` SET `votes`=`votes` + ? WHERE `id` = ?",
			expectVars: []any{uint64(1), 1},
		},
		{
			name: "Decrement does not fire events when they are disabled",
			run: func(query contractsorm.Query) error {
				return query.Model(&IncrementEventModel{ID: 1}).WithoutEvents().Decrement("votes")
			},
			expectSql:  "UPDATE `increment_event_models` SET `votes`=`votes` - ? WHERE `id` = ?",
			expectVars: []any{uint64(1), uint(1)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sql, vars = "", nil

			err := test.run(newQuery(t))
			if test.expectError != nil {
				assert.EqualError(t, err, test.expectError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expectSql, sql)
			assert.Equal(t, test.expectVars, vars)
		})
	}
}

func TestIncrementAndDecrementTouchUpdatedAt(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	var sql string
	assert.NoError(t, db.Callback().Update().After("gorm:update").Register("test:capture", func(tx *gormio.DB) {
		sql = tx.Statement.SQL.String()
	}))

	mockGrammar := mocksdriver.NewGrammar(t)
	mockGrammar.EXPECT().CompileJsonColumnsUpdate(mock.Anything).RunAndReturn(func(values map[string]any) (map[string]any, error) {
		return values, nil
	}).Once()

	model := &IncrementTimestampModel{ID: 1, Votes: 1}
	query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, mockGrammar, nil, nil, nil, nil)

	assert.NoError(t, query.Model(model).Increment("votes"))
	assert.Equal(t, "UPDATE `increment_timestamp_models` SET `votes`=`votes` + ?,`updated_at`=? WHERE `id` = ?", sql)
	assert.False(t, model.UpdatedAt.IsZero())
	// No row is affected in DryRun mode, so the column is not synced in memory.
	assert.Equal(t, 1, model.Votes)
}

func TestIncrementAndDecrementSyncTheModel(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)
	// DryRun affects no rows: pretend one row was updated so the in-memory sync runs.
	assert.NoError(t, db.Callback().Update().After("gorm:update").Register("test:rows", func(tx *gormio.DB) {
		tx.RowsAffected = 1
	}))

	newQuery := func() contractsorm.Query {
		mockGrammar := mocksdriver.NewGrammar(t)
		mockGrammar.EXPECT().CompileJsonColumnsUpdate(mock.Anything).RunAndReturn(func(values map[string]any) (map[string]any, error) {
			return values, nil
		}).Maybe()

		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, mockGrammar, nil, nil, nil, nil)
	}

	ptr := 10
	model := &IncrementSyncModel{ID: 1, Int: 5, Uint: 5, Float: 1.5, Ptr: &ptr, Name: "goravel", NullInt: sql.NullInt64{Int64: 2, Valid: true}}

	assert.NoError(t, newQuery().Model(model).Increment("int", 2))
	assert.NoError(t, newQuery().Model(model).Decrement("uint"))
	assert.NoError(t, newQuery().Model(model).Increment("float", 3))
	assert.NoError(t, newQuery().Model(model).Decrement("ptr", 4))
	// Nil pointers, non numeric and unknown fields are left alone.
	assert.NoError(t, newQuery().Model(model).Increment("nil_ptr"))
	assert.NoError(t, newQuery().Model(model).Increment("name"))
	assert.NoError(t, newQuery().Model(model).Increment("unknown"))

	assert.Equal(t, 7, model.Int)
	assert.Equal(t, uint(4), model.Uint)
	assert.Equal(t, 4.5, model.Float)
	assert.Equal(t, 6, *model.Ptr)
	assert.Nil(t, model.NilPtr)
	assert.Equal(t, "goravel", model.Name)

	// Decrementing the other numeric kinds.
	assert.NoError(t, newQuery().Model(model).Decrement("int"))
	assert.NoError(t, newQuery().Model(model).Increment("uint", 6))
	assert.NoError(t, newQuery().Model(model).Decrement("float", 2))
	assert.Equal(t, 6, model.Int)
	assert.Equal(t, uint(10), model.Uint)
	assert.Equal(t, 2.5, model.Float)

	// Unsigned values that would underflow are left untouched instead of wrapping around.
	assert.NoError(t, newQuery().Model(model).Decrement("uint", 11))
	assert.Equal(t, uint(10), model.Uint)

	// sql.Null* fields are set through GORM, NULL values are left alone.
	assert.NoError(t, newQuery().Model(model).Increment("null_int", 3))
	assert.Equal(t, sql.NullInt64{Int64: 5, Valid: true}, model.NullInt)
	assert.NoError(t, newQuery().Model(model).Increment("null_int_invalid"))
	assert.Equal(t, sql.NullInt64{}, model.NullIntInvalid)

	// A model without a primary key value has no condition to update by: GORM
	// refuses the update and the struct isn't touched.
	unsaved := &IncrementSyncModel{Int: 1}
	assert.ErrorIs(t, newQuery().Model(unsaved).Increment("int"), gormio.ErrMissingWhereClause)
	assert.Equal(t, 1, unsaved.Int)
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
			input:       new("test"),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "pointer to int",
			input:       new(123),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "pointer to float",
			input:       new(123.45),
			expectError: true,
			expectedErr: errors.OrmQueryInvalidModel.Args(""),
		},
		{
			name:        "pointer to bool",
			input:       new(true),
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
				assert.True(t, resultValue.Kind() == reflect.Pointer)
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
				assert.Equal(t, reflect.TypeFor[*TestModel](), reflect.TypeOf(result))
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

func TestQuoteColumns(t *testing.T) {
	group := clause.Column{Name: "group"}
	order := clause.Column{Name: "order"}

	tests := []struct {
		name          string
		query         func(contractsorm.Query) contractsorm.Query
		expectedWhere []contractsdriver.Where
		expectedOrder []any
	}{
		{
			name: "WhereIn",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereIn("group", []any{"a"})
			},
			expectedWhere: []contractsdriver.Where{{Query: "? IN ?", Args: []any{group, []any{"a"}}}},
		},
		{
			name: "OrWhereIn",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.OrWhereIn("group", []any{"a"})
			},
			expectedWhere: []contractsdriver.Where{{Query: "? IN ?", Args: []any{group, []any{"a"}}, Or: true}},
		},
		{
			name: "WhereNotIn",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereNotIn("group", []any{"a"})
			},
			expectedWhere: []contractsdriver.Where{{Query: "? NOT IN ?", Args: []any{group, []any{"a"}}}},
		},
		{
			name: "OrWhereNotIn",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.OrWhereNotIn("group", []any{"a"})
			},
			expectedWhere: []contractsdriver.Where{{Query: "? NOT IN ?", Args: []any{group, []any{"a"}}, Or: true}},
		},
		{
			name: "WhereBetween",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereBetween("order", 1, 2)
			},
			expectedWhere: []contractsdriver.Where{{Query: "? BETWEEN ? AND ?", Args: []any{order, 1, 2}}},
		},
		{
			name: "WhereNotBetween",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereNotBetween("order", 1, 2)
			},
			expectedWhere: []contractsdriver.Where{{Query: "? NOT BETWEEN ? AND ?", Args: []any{order, 1, 2}}},
		},
		{
			name: "OrWhereBetween",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.OrWhereBetween("order", 1, 2)
			},
			expectedWhere: []contractsdriver.Where{{Query: "? BETWEEN ? AND ?", Args: []any{order, 1, 2}, Or: true}},
		},
		{
			name: "OrWhereNotBetween",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.OrWhereNotBetween("order", 1, 2)
			},
			expectedWhere: []contractsdriver.Where{{Query: "? NOT BETWEEN ? AND ?", Args: []any{order, 1, 2}, Or: true}},
		},
		{
			name: "WhereNull",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereNull("group")
			},
			expectedWhere: []contractsdriver.Where{{Query: "? IS NULL", Args: []any{group}}},
		},
		{
			name: "OrWhereNull",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.OrWhereNull("group")
			},
			expectedWhere: []contractsdriver.Where{{Query: "? IS NULL", Args: []any{group}, Or: true}},
		},
		{
			name: "WhereNotNull",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereNotNull("group")
			},
			expectedWhere: []contractsdriver.Where{{Query: "? IS NOT NULL", Args: []any{group}}},
		},
		{
			name: "WhereAll",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereAll([]string{"group", "order"}, "a")
			},
			expectedWhere: []contractsdriver.Where{{Query: "? = ? AND ? = ?", Args: []any{group, "a", order, "a"}}},
		},
		{
			name: "WhereAny",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereAny([]string{"group", "order"}, "a")
			},
			expectedWhere: []contractsdriver.Where{{Query: "(? = ? OR ? = ?)", Args: []any{group, "a", order, "a"}}},
		},
		{
			name: "WhereNone",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereNone([]string{"group", "order"}, "LIKE", "a")
			},
			expectedWhere: []contractsdriver.Where{{Query: "NOT (? LIKE ?) AND NOT (? LIKE ?)", Args: []any{group, "a", order, "a"}}},
		},
		{
			name: "WhereNone with equal operator",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereNone([]string{"group"}, "a")
			},
			expectedWhere: []contractsdriver.Where{{Query: "? <> ?", Args: []any{group, "a"}}},
		},
		{
			name: "OrderBy",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.OrderBy("order")
			},
			expectedOrder: []any{clause.OrderByColumn{Column: order}},
		},
		{
			name: "OrderBy with lower case direction",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.OrderBy("order", "desc")
			},
			expectedOrder: []any{clause.OrderByColumn{Column: order, Desc: true}},
		},
		{
			name: "OrderBy with custom direction is not quoted",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.OrderBy("order", "DESC NULLS LAST")
			},
			expectedOrder: []any{"order DESC NULLS LAST"},
		},
		{
			name: "OrderByDesc",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.OrderByDesc("order")
			},
			expectedOrder: []any{clause.OrderByColumn{Column: order, Desc: true}},
		},
		{
			name: "table qualified column",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereNull("users.group")
			},
			expectedWhere: []contractsdriver.Where{{Query: "? IS NULL", Args: []any{clause.Column{Name: "users.group"}}}},
		},
		{
			name: "expression is not quoted",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereNull("LOWER(name)")
			},
			expectedWhere: []contractsdriver.Where{{Query: "LOWER(name) IS NULL"}},
		},
		{
			name: "json selector is not quoted",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereNull("data->name")
			},
			expectedWhere: []contractsdriver.Where{{Query: "data->name IS NULL"}},
		},
		{
			name: "quoted column is not quoted again",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.WhereNull("`group`")
			},
			expectedWhere: []contractsdriver.Where{{Query: "`group` IS NULL"}},
		},
		{
			name: "expression in OrderBy is not quoted",
			query: func(query contractsorm.Query) contractsorm.Query {
				return query.OrderByDesc("LOWER(name)")
			},
			expectedOrder: []any{"LOWER(name) DESC"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// No gorm instance: the column is quoted when the query is built, by the connection that runs it,
			// so these methods must not need one (e.g. when the database is not configured).
			query := test.query(NewQuery(context.Background(), nil, contractsdatabase.Config{}, nil, nil, nil, nil, nil, nil)).(*Query)

			assert.Equal(t, test.expectedWhere, query.conditions.where)
			assert.Equal(t, test.expectedOrder, query.conditions.order)
		})
	}
}

func TestWhen(t *testing.T) {
	query := &Query{}
	trueQuery := &Query{}
	falseQuery := &Query{}

	callback := func(q contractsorm.Query) contractsorm.Query {
		assert.Same(t, query, q)
		return trueQuery
	}
	falseCallback := func(q contractsorm.Query) contractsorm.Query {
		assert.Same(t, query, q)
		return falseQuery
	}

	assert.Same(t, trueQuery, query.When(true, callback))
	assert.Same(t, trueQuery, query.When(true, callback, falseCallback))
	assert.Same(t, query, query.When(false, callback))
	assert.Same(t, falseQuery, query.When(false, callback, falseCallback))
}

func TestLatest(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func() contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).Table("users")
	}

	tests := []struct {
		name      string
		query     func() contractsorm.Query
		expectSql string
	}{
		{
			name:      "default column",
			query:     func() contractsorm.Query { return newQuery().Latest() },
			expectSql: "SELECT * FROM `users` ORDER BY `created_at` DESC",
		},
		{
			name:      "empty column falls back to default",
			query:     func() contractsorm.Query { return newQuery().Latest("") },
			expectSql: "SELECT * FROM `users` ORDER BY `created_at` DESC",
		},
		{
			name:      "custom column",
			query:     func() contractsorm.Query { return newQuery().Latest("id") },
			expectSql: "SELECT * FROM `users` ORDER BY `id` DESC",
		},
		{
			name:      "combined with other orders",
			query:     func() contractsorm.Query { return newQuery().Latest("updated_at").OrderBy("id") },
			expectSql: "SELECT * FROM `users` ORDER BY `updated_at` DESC,`id`",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := test.query().(*Query).buildConditions()
			statement := query.instance.Find(&[]map[string]any{}).Statement
			assert.NoError(t, statement.Error)
			assert.Equal(t, test.expectSql, statement.SQL.String())
		})
	}
}

func TestValueWithNonPointerDest(t *testing.T) {
	var name string
	err := (&Query{}).Value("name", name)

	assert.EqualError(t, err, errors.DatabaseUnsupportedType.Args(reflect.String, "pointer").Error())
}

func TestValueAndDoesntExist(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func(db *gormio.DB) contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).Table("users")
	}

	// Value scans, which GORM refuses in dry run mode: the query is built and
	// the error is returned without touching dest.
	name := "unchanged"
	assert.ErrorIs(t, newQuery(db).Where("id", 1).Value("name", &name), gormio.ErrDryRunModeUnsupported)
	assert.Equal(t, "unchanged", name)

	doesntExist, err := newQuery(db).Where("id", 1).DoesntExist()
	assert.NoError(t, err)
	assert.True(t, doesntExist)

	// Errors of the underlying query are returned.
	failing := db.Session(&gormio.Session{})
	_ = failing.AddError(assert.AnError)

	assert.ErrorIs(t, newQuery(failing).Value("name", &name), assert.AnError)

	doesntExist, err = newQuery(failing).DoesntExist()
	assert.ErrorIs(t, err, assert.AnError)
	assert.False(t, doesntExist)
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

// cursorDriver is a database/sql driver whose queries return `rows` rows with an "id" column, and that
// records when the rows are closed.
type cursorDriver struct {
	rows   int
	closed atomic.Bool
}

func (r *cursorDriver) Open(string) (driver.Conn, error) { return &cursorConn{driver: r}, nil }

type cursorConn struct{ driver *cursorDriver }

func (r *cursorConn) Prepare(string) (driver.Stmt, error) { return &cursorStmt{driver: r.driver}, nil }
func (r *cursorConn) Close() error                        { return nil }
func (r *cursorConn) Begin() (driver.Tx, error)           { return nil, fmt.Errorf("not supported") }

type cursorStmt struct{ driver *cursorDriver }

func (r *cursorStmt) Close() error  { return nil }
func (r *cursorStmt) NumInput() int { return -1 }
func (r *cursorStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, fmt.Errorf("not supported")
}
func (r *cursorStmt) Query([]driver.Value) (driver.Rows, error) {
	return &cursorRows{driver: r.driver}, nil
}

type cursorRows struct {
	driver *cursorDriver
	next   int
}

func (r *cursorRows) Columns() []string { return []string{"id"} }
func (r *cursorRows) Close() error {
	r.driver.closed.Store(true)
	return nil
}
func (r *cursorRows) Next(dest []driver.Value) error {
	if r.next >= r.driver.rows {
		return io.EOF
	}
	r.next++
	dest[0] = int64(r.next)
	return nil
}

var cursorDriverID atomic.Int64

// newCursorQuery returns a query on a new cursorDriver, and a function that closes its database.
func newCursorQuery(t *testing.T, ctx context.Context, rows int) (*Query, *cursorDriver, func()) {
	fake := &cursorDriver{rows: rows}
	name := fmt.Sprintf("goravel-orm-cursor-%d", cursorDriverID.Add(1))
	sql.Register(name, fake)

	sqlDB, err := sql.Open(name, "")
	assert.NoError(t, err)

	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{ConnPool: sqlDB})
	assert.NoError(t, err)

	query := NewQuery(ctx, nil, contractsdatabase.Config{}, db.WithContext(ctx), nil, nil, nil, nil, nil)

	return query.Table("users").(*Query), fake, func() { _ = sqlDB.Close() }
}

type IncrementSyncModel struct {
	ID     uint `gorm:"primaryKey"`
	Int    int
	Uint   uint
	Float  float64
	Ptr    *int
	NilPtr *int
	Name   string

	NullInt        sql.NullInt64
	NullIntInvalid sql.NullInt64
}

type IncrementModel struct {
	ID        uint `gorm:"primaryKey"`
	Votes     int
	DeletedAt gormio.DeletedAt
}

func (m *IncrementModel) GlobalScopes() map[string]func(contractsorm.Query) contractsorm.Query {
	return map[string]func(contractsorm.Query) contractsorm.Query{
		"active": func(query contractsorm.Query) contractsorm.Query {
			return query.Where("active", true)
		},
	}
}

type IncrementEventModel struct {
	ID    uint `gorm:"primaryKey"`
	Votes int
}

func (m *IncrementEventModel) DispatchesEvents() map[contractsorm.EventType]func(contractsorm.Event) error {
	return map[contractsorm.EventType]func(contractsorm.Event) error{
		contractsorm.EventUpdating: func(event contractsorm.Event) error {
			return errors.New("updating")
		},
	}
}

type IncrementTimestampModel struct {
	ID        uint `gorm:"primaryKey"`
	Votes     int
	UpdatedAt time.Time
}
