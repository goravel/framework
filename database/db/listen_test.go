package db

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormtests "gorm.io/gorm/utils/tests"

	contractsdb "github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/database/logger"
	"github.com/goravel/framework/database/utils"
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
