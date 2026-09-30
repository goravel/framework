package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormtests "gorm.io/gorm/utils/tests"

	contractsdb "github.com/goravel/framework/contracts/database/db"
)

// collectQueryEvents registers a listener that records dispatched events and
// returns them. The registry is package-level, so every test clears it.
func collectQueryEvents(t *testing.T) func() []*contractsdb.QueryExecuted {
	t.Helper()
	t.Cleanup(ClearQueryListeners)

	var events []*contractsdb.QueryExecuted
	ListenQuery(func(event *contractsdb.QueryExecuted) error {
		events = append(events, event)

		return nil
	})

	return func() []*contractsdb.QueryExecuted {
		return events
	}
}

func TestListenQueryDispatchesToMultipleListeners(t *testing.T) {
	t.Cleanup(ClearQueryListeners)

	var order []string
	ListenQuery(func(event *contractsdb.QueryExecuted) error {
		order = append(order, "first:"+event.Sql)

		return nil
	})
	ListenQuery(func(event *contractsdb.QueryExecuted) error {
		order = append(order, "second:"+event.Sql)

		return nil
	})

	DispatchQueryEvent(&contractsdb.QueryExecuted{Sql: "SELECT 1"})

	assert.Equal(t, []string{"first:SELECT 1", "second:SELECT 1"}, order)
}

func TestListenQueryNilListenerIgnored(t *testing.T) {
	t.Cleanup(ClearQueryListeners)

	assert.False(t, HasQueryListeners())

	ListenQuery(nil)
	assert.False(t, HasQueryListeners())

	// Dispatching with only a nil listener registered must not panic.
	assert.NotPanics(t, func() {
		DispatchQueryEvent(&contractsdb.QueryExecuted{Sql: "SELECT 1"})
	})
}

func TestDispatchQueryEventListenerErrorDoesNotStopOthers(t *testing.T) {
	t.Cleanup(ClearQueryListeners)

	var secondCalled bool
	ListenQuery(func(event *contractsdb.QueryExecuted) error {
		return errors.New("listener failed")
	})
	ListenQuery(func(event *contractsdb.QueryExecuted) error {
		secondCalled = true

		return nil
	})

	// Listener errors are swallowed by the registry: the query flow must never
	// break and later listeners must still run.
	assert.NotPanics(t, func() {
		DispatchQueryEvent(&contractsdb.QueryExecuted{Sql: "SELECT 1"})
	})
	assert.True(t, secondCalled)
}

func TestDispatchQueryEventPanickingListenerIsRecovered(t *testing.T) {
	t.Cleanup(ClearQueryListeners)

	var secondCalled bool
	ListenQuery(func(event *contractsdb.QueryExecuted) error {
		panic("listener panic")
	})
	ListenQuery(func(event *contractsdb.QueryExecuted) error {
		secondCalled = true

		return nil
	})

	assert.NotPanics(t, func() {
		DispatchQueryEvent(&contractsdb.QueryExecuted{Sql: "SELECT 1"})
	})
	// The panic is swallowed per listener, so the remaining ones still run.
	assert.True(t, secondCalled)
}

func TestHasQueryListenersAndClear(t *testing.T) {
	t.Cleanup(ClearQueryListeners)
	ClearQueryListeners()

	assert.False(t, HasQueryListeners())

	ListenQuery(func(event *contractsdb.QueryExecuted) error { return nil })
	assert.True(t, HasQueryListeners())

	ListenQuery(func(event *contractsdb.QueryExecuted) error { return nil })
	ClearQueryListeners()
	assert.False(t, HasQueryListeners())
}

func TestQueryBindingsFromContext(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		ctx := WithQueryBindings(context.Background(), "SELECT * FROM users WHERE id = ?", []any{1, "two"})

		sql, args, ok := QueryBindingsFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "SELECT * FROM users WHERE id = ?", sql)
		assert.Equal(t, []any{1, "two"}, args)
	})

	t.Run("without bindings", func(t *testing.T) {
		sql, args, ok := QueryBindingsFromContext(context.Background())
		assert.False(t, ok)
		assert.Empty(t, sql)
		assert.Nil(t, args)
	})

	t.Run("wrong type under the key", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), queryBindingsKey{}, "not bindings")

		sql, args, ok := QueryBindingsFromContext(ctx)
		assert.False(t, ok)
		assert.Empty(t, sql)
		assert.Nil(t, args)
	})
}

func TestQueryEventPluginName(t *testing.T) {
	assert.Equal(t, QueryEventPluginName, NewQueryEventPlugin().Name())
}

func TestQueryEventPluginInitialize(t *testing.T) {
	instance, err := gorm.Open(gormtests.DummyDialector{}, &gorm.Config{DryRun: true})
	require.NoError(t, err)

	require.NoError(t, NewQueryEventPlugin().Initialize(instance))

	// Re-initializing replaces the same-named callbacks without error, and the
	// plugin must still be registered under its name.
	require.NoError(t, NewQueryEventPlugin().Initialize(instance))
	assert.NotNil(t, instance.Callback().Query().Get(QueryEventPluginName+":after"))
}

func newDummyGormDB(t *testing.T) *gorm.DB {
	t.Helper()

	instance, err := gorm.Open(gormtests.DummyDialector{}, &gorm.Config{DryRun: true})
	require.NoError(t, err)

	return instance
}

func TestQueryEventPluginAfter(t *testing.T) {
	plugin := NewQueryEventPlugin()

	t.Run("skipped without listeners", func(t *testing.T) {
		t.Cleanup(ClearQueryListeners)
		ClearQueryListeners()

		db := newDummyGormDB(t)
		tx := db.Session(&gorm.Session{NewDB: true})
		tx.Statement.SQL.WriteString("SELECT * FROM users WHERE id = ?")
		tx.Statement.Vars = append(tx.Statement.Vars, 1)

		plugin.after(tx)

		_, _, ok := QueryBindingsFromContext(tx.Statement.Context)
		assert.False(t, ok)
	})

	t.Run("carries sql and a copy of vars", func(t *testing.T) {
		collectQueryEvents(t)

		db := newDummyGormDB(t)
		tx := db.Session(&gorm.Session{NewDB: true})
		tx.Statement.SQL.WriteString("SELECT * FROM users WHERE id = ?")
		tx.Statement.Vars = append(tx.Statement.Vars, 1)

		plugin.after(tx)

		sql, args, ok := QueryBindingsFromContext(tx.Statement.Context)
		require.True(t, ok)
		assert.Equal(t, "SELECT * FROM users WHERE id = ?", sql)
		assert.Equal(t, []any{1}, args)

		// The captured bindings must be a copy: later mutation of the
		// statement vars must not change what was carried on the context.
		tx.Statement.Vars[0] = 99
		_, args, ok = QueryBindingsFromContext(tx.Statement.Context)
		require.True(t, ok)
		assert.Equal(t, []any{1}, args)
	})

	t.Run("skipped with empty sql", func(t *testing.T) {
		collectQueryEvents(t)

		db := newDummyGormDB(t)
		tx := db.Session(&gorm.Session{NewDB: true})

		plugin.after(tx)

		_, _, ok := QueryBindingsFromContext(tx.Statement.Context)
		assert.False(t, ok)
	})

	t.Run("nil statement does not panic", func(t *testing.T) {
		collectQueryEvents(t)

		assert.NotPanics(t, func() {
			plugin.after(&gorm.DB{})
		})
	})
}
