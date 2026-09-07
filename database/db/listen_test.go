package db

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormtests "gorm.io/gorm/utils/tests"

	contractsdb "github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/database/logger"
	"github.com/goravel/framework/database/utils"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	"github.com/goravel/framework/support/carbon"
)

// collectEvents registers a listener that records events for the given
// connection only, and returns the recorded events. The registry is global and
// package-level, so tests filter by a unique connection name to stay isolated.
func collectEvents(t *testing.T, connection string) func() []*contractsdb.QueryExecuted {
	t.Helper()
	t.Cleanup(utils.ClearQueryListeners)

	var (
		mu     sync.Mutex
		events []*contractsdb.QueryExecuted
	)

	Listen(func(event *contractsdb.QueryExecuted) error {
		if event.Connection != connection {
			return nil
		}

		mu.Lock()
		defer mu.Unlock()
		events = append(events, event)

		return nil
	})

	return func() []*contractsdb.QueryExecuted {
		mu.Lock()
		defer mu.Unlock()

		return events
	}
}

func TestListenLoggerTrace(t *testing.T) {
	events := collectEvents(t, "listen-trace")

	logger := &Logger{connection: "listen-trace", level: logger.Silent}
	begin := carbon.Now().SubDuration((50 * time.Millisecond).String())
	logger.Trace(context.Background(), begin, "SELECT * FROM users WHERE id = 1", 1, nil)

	got := events()
	require.Len(t, got, 1)
	assert.Equal(t, "listen-trace", got[0].Connection)
	assert.Equal(t, "SELECT * FROM users WHERE id = 1", got[0].Sql)
	assert.Equal(t, "SELECT * FROM users WHERE id = 1", got[0].RawSql)
	assert.Nil(t, got[0].Bindings)
	assert.Nil(t, got[0].Error)
	assert.Greater(t, got[0].Time, time.Duration(0))
}

func TestListenWithBindingsFromContext(t *testing.T) {
	events := collectEvents(t, "listen-bindings")

	ctx := utils.WithQueryBindings(context.Background(), "SELECT * FROM users WHERE id = ?", []any{1})
	logger := &Logger{connection: "listen-bindings"}
	logger.Trace(ctx, carbon.Now(), "SELECT * FROM users WHERE id = 1", 1, nil)

	got := events()
	require.Len(t, got, 1)
	assert.Equal(t, "SELECT * FROM users WHERE id = ?", got[0].Sql)
	assert.Equal(t, []any{1}, got[0].Bindings)
	assert.Equal(t, "SELECT * FROM users WHERE id = 1", got[0].RawSql)
}

func TestListenErrorPropagated(t *testing.T) {
	events := collectEvents(t, "listen-error")

	queryErr := errors.New("boom")
	logger := &Logger{connection: "listen-error"}
	logger.Trace(context.Background(), carbon.Now(), "INSERT INTO users", -1, queryErr)

	got := events()
	require.Len(t, got, 1)
	assert.ErrorIs(t, got[0].Error, queryErr)
}

func TestListenNilListenerIgnored(t *testing.T) {
	t.Cleanup(utils.ClearQueryListeners)

	Listen(nil)
	// A nil listener must not be registered; dispatching must not panic.
	assert.NotPanics(t, func() {
		logger := &Logger{connection: "listen-nil"}
		logger.Trace(context.Background(), carbon.Now(), "SELECT 1", 1, nil)
	})
}

func TestListenPanickingListenerDoesNotBreakQuery(t *testing.T) {
	t.Cleanup(utils.ClearQueryListeners)

	Listen(func(event *contractsdb.QueryExecuted) error {
		if event.Connection == "listen-panic" {
			panic("listener panic")
		}

		return nil
	})

	logger := &Logger{connection: "listen-panic"}
	assert.NotPanics(t, func() {
		logger.Trace(context.Background(), carbon.Now(), "SELECT 1", 1, nil)
	})
}

func TestListenGormEndToEnd(t *testing.T) {
	events := collectEvents(t, "listen-gorm")

	instance, err := gorm.Open(gormtests.DummyDialector{}, &gorm.Config{
		DryRun: true,
		Logger: (&Logger{connection: "listen-gorm"}).ToGorm(),
	})
	require.NoError(t, err)
	require.NoError(t, instance.Use(utils.NewQueryEventPlugin()))

	var dest []map[string]any
	require.NoError(t, instance.WithContext(context.Background()).Table("users").Where("id = ?", 1).Find(&dest).Error)

	got := events()
	require.Len(t, got, 1)
	assert.Equal(t, "listen-gorm", got[0].Connection)
	assert.Contains(t, got[0].Sql, "id = ?")
	assert.Equal(t, []any{1}, got[0].Bindings)
	assert.Contains(t, got[0].RawSql, "id = 1")
	assert.Nil(t, got[0].Error)
	// Dry-run queries finish within one clock tick on Windows, so only assert
	// the duration is non-negative here; TestListenLoggerTrace covers Time.
	assert.GreaterOrEqual(t, got[0].Time, time.Duration(0))
}

func TestDBListenRegistersGlobally(t *testing.T) {
	t.Cleanup(utils.ClearQueryListeners)
	events := collectEvents(t, "listen-db")

	r := &DB{}
	r.Listen(func(event *contractsdb.QueryExecuted) error { return nil })

	logger := &Logger{connection: "listen-db"}
	logger.Trace(context.Background(), carbon.Now(), "SELECT 1", 1, nil)

	assert.Len(t, events(), 1)
}

func TestLoggerWithConnection(t *testing.T) {
	events := collectEvents(t, "listen-connection")

	r := &Logger{}
	assert.Same(t, r, r.WithConnection("listen-connection"))
	assert.Equal(t, "listen-connection", r.connection)

	// The connection set through WithConnection is reported in events.
	r.Trace(context.Background(), carbon.Now(), "SELECT 1", 1, nil)

	got := events()
	require.Len(t, got, 1)
	assert.Equal(t, "listen-connection", got[0].Connection)
}

func TestTxSelectWithListenersCarriesBindings(t *testing.T) {
	events := collectEvents(t, "listen-tx-select")

	ctx := context.Background()
	parameterizedSQL := "SELECT * FROM users WHERE name = ?"
	explainedSQL := `SELECT * FROM users WHERE name = "John"`

	mockBuilder := mocksdb.NewTxBuilder(t)
	mockBuilder.EXPECT().Explain(parameterizedSQL, "John").Return(explainedSQL).Once()
	mockBuilder.EXPECT().SelectContext(ctx, mock.Anything, parameterizedSQL, "John").Return(nil).Once()

	tx := &Tx{ctx: ctx, logger: &Logger{connection: "listen-tx-select", level: logger.Silent}, txBuilder: mockBuilder}

	var users []TestUser
	require.NoError(t, tx.Select(&users, parameterizedSQL, "John"))

	got := events()
	require.Len(t, got, 1)
	assert.Equal(t, parameterizedSQL, got[0].Sql)
	assert.Equal(t, []any{"John"}, got[0].Bindings)
	assert.Equal(t, explainedSQL, got[0].RawSql)
	assert.Nil(t, got[0].Error)
}

func TestTxSelectWithListenersErrorCarried(t *testing.T) {
	events := collectEvents(t, "listen-tx-select-err")

	ctx := context.Background()
	queryErr := errors.New("select failed")

	mockBuilder := mocksdb.NewTxBuilder(t)
	mockBuilder.EXPECT().Explain("SELECT 1").Return("SELECT 1").Once()
	mockBuilder.EXPECT().GetContext(ctx, mock.Anything, "SELECT 1").Return(queryErr).Once()

	tx := &Tx{ctx: ctx, logger: &Logger{connection: "listen-tx-select-err", level: logger.Silent}, txBuilder: mockBuilder}

	var user TestUser
	assert.ErrorIs(t, tx.Select(&user, "SELECT 1"), queryErr)

	got := events()
	require.Len(t, got, 1)
	assert.ErrorIs(t, got[0].Error, queryErr)
}

func TestTxSelectSliceErrorCarried(t *testing.T) {
	events := collectEvents(t, "listen-tx-select-slice-err")

	ctx := context.Background()
	queryErr := errors.New("select slice failed")

	mockBuilder := mocksdb.NewTxBuilder(t)
	mockBuilder.EXPECT().Explain("SELECT * FROM users WHERE id = ?", 1).Return("SELECT * FROM users WHERE id = 1").Once()
	mockBuilder.EXPECT().SelectContext(ctx, mock.Anything, "SELECT * FROM users WHERE id = ?", 1).Return(queryErr).Once()

	tx := &Tx{ctx: ctx, logger: &Logger{connection: "listen-tx-select-slice-err", level: logger.Silent}, txBuilder: mockBuilder}

	var users []TestUser
	assert.ErrorIs(t, tx.Select(&users, "SELECT * FROM users WHERE id = ?", 1), queryErr)

	got := events()
	require.Len(t, got, 1)
	assert.ErrorIs(t, got[0].Error, queryErr)
	assert.Equal(t, "SELECT * FROM users WHERE id = ?", got[0].Sql)
	assert.Equal(t, []any{1}, got[0].Bindings)
}

func TestTxExecWithListenersCarriesBindings(t *testing.T) {
	events := collectEvents(t, "listen-tx-exec")

	ctx := context.Background()
	sqlStr := "INSERT INTO users (name) VALUES (?)"
	explainedSQL := `INSERT INTO users (name) VALUES ("John")`

	mockResult := &MockResult{}
	mockResult.On("RowsAffected").Return(int64(1), nil).Once()

	mockBuilder := mocksdb.NewTxBuilder(t)
	mockBuilder.EXPECT().Explain(sqlStr, "John").Return(explainedSQL).Once()
	mockBuilder.EXPECT().ExecContext(ctx, sqlStr, "John").Return(mockResult, nil).Once()

	tx := &Tx{ctx: ctx, logger: &Logger{connection: "listen-tx-exec", level: logger.Silent}, txBuilder: mockBuilder}

	result, err := tx.Insert(sqlStr, "John")
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.RowsAffected)

	got := events()
	require.Len(t, got, 1)
	assert.Equal(t, sqlStr, got[0].Sql)
	assert.Equal(t, []any{"John"}, got[0].Bindings)
	assert.Equal(t, explainedSQL, got[0].RawSql)
	assert.Nil(t, got[0].Error)
}

func TestTxExecWithListenersErrorCarried(t *testing.T) {
	events := collectEvents(t, "listen-tx-exec-err")

	ctx := context.Background()
	execErr := errors.New("exec failed")

	mockBuilder := mocksdb.NewTxBuilder(t)
	mockBuilder.EXPECT().Explain("DELETE FROM users").Return("DELETE FROM users").Once()
	mockBuilder.EXPECT().ExecContext(ctx, "DELETE FROM users").Return(nil, execErr).Once()

	tx := &Tx{ctx: ctx, logger: &Logger{connection: "listen-tx-exec-err", level: logger.Silent}, txBuilder: mockBuilder}

	_, err := tx.Delete("DELETE FROM users")
	assert.ErrorIs(t, err, execErr)

	got := events()
	require.Len(t, got, 1)
	assert.ErrorIs(t, got[0].Error, execErr)
}

func TestQueryTraceWithListeners(t *testing.T) {
	events := collectEvents(t, "listen-query-trace")

	ctx := context.Background()
	parameterizedSQL := "SELECT * FROM users WHERE id = ?"
	explainedSQL := "SELECT * FROM users WHERE id = 1"

	mockBuilder := mocksdb.NewCommonBuilder(t)
	mockBuilder.EXPECT().Explain(parameterizedSQL, 1).Return(explainedSQL).Once()

	r := &Query{ctx: ctx, logger: &Logger{connection: "listen-query-trace", level: logger.Silent}}
	r.trace(mockBuilder, parameterizedSQL, []any{1}, carbon.Now(), 1, nil)

	got := events()
	require.Len(t, got, 1)
	assert.Equal(t, parameterizedSQL, got[0].Sql)
	assert.Equal(t, []any{1}, got[0].Bindings)
	assert.Equal(t, explainedSQL, got[0].RawSql)
}

func TestQueryTraceWithListenersInTransaction(t *testing.T) {
	events := collectEvents(t, "listen-query-trace-tx")

	ctx := context.Background()
	parameterizedSQL := "UPDATE users SET name = ? WHERE id = ?"
	explainedSQL := "UPDATE users SET name = 'John' WHERE id = 1"

	mockBuilder := mocksdb.NewCommonBuilder(t)
	mockBuilder.EXPECT().Explain(parameterizedSQL, "John", 1).Return(explainedSQL).Once()

	txLogs := []TxLog{}
	r := &Query{ctx: ctx, logger: &Logger{connection: "listen-query-trace-tx", level: logger.Silent}, txLogs: &txLogs}
	r.trace(mockBuilder, parameterizedSQL, []any{"John", 1}, carbon.Now(), 2, nil)

	// Inside a transaction the trace is buffered; the event fires on Commit.
	assert.Empty(t, events())
	require.Len(t, txLogs, 1)

	sql, bindings, ok := utils.QueryBindingsFromContext(txLogs[0].ctx)
	require.True(t, ok)
	assert.Equal(t, parameterizedSQL, sql)
	assert.Equal(t, []any{"John", 1}, bindings)

	// Commit replays the buffered logs through Trace, firing the event.
	mockTxBuilder := mocksdb.NewTxBuilder(t)
	mockTxBuilder.EXPECT().Commit().Return(nil).Once()

	tx := &Tx{ctx: ctx, logger: &Logger{connection: "listen-query-trace-tx", level: logger.Silent}, txBuilder: mockTxBuilder, txLogs: &txLogs}
	require.NoError(t, tx.Commit())

	got := events()
	require.Len(t, got, 1)
	assert.Equal(t, parameterizedSQL, got[0].Sql)
	assert.Equal(t, []any{"John", 1}, got[0].Bindings)
	assert.Equal(t, explainedSQL, got[0].RawSql)
}
