package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sync/atomic"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsdb "github.com/goravel/framework/contracts/database/db"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksdriver "github.com/goravel/framework/mocks/database/driver"
	mockslogger "github.com/goravel/framework/mocks/database/logger"
)

// eachDriver is a database/sql driver whose queries return `rows` rows with an "id" column, and that
// records when the rows are closed.
type eachDriver struct {
	rows   int
	closed atomic.Bool
}

func (r *eachDriver) Open(string) (driver.Conn, error) { return &eachConn{driver: r}, nil }

type eachConn struct{ driver *eachDriver }

func (r *eachConn) Prepare(string) (driver.Stmt, error) { return &eachStmt{driver: r.driver}, nil }
func (r *eachConn) Close() error                        { return nil }
func (r *eachConn) Begin() (driver.Tx, error)           { return nil, fmt.Errorf("not supported") }

type eachStmt struct{ driver *eachDriver }

func (r *eachStmt) Close() error  { return nil }
func (r *eachStmt) NumInput() int { return -1 }
func (r *eachStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, fmt.Errorf("not supported")
}
func (r *eachStmt) Query([]driver.Value) (driver.Rows, error) {
	return &eachRows{driver: r.driver}, nil
}

type eachRows struct {
	driver *eachDriver
	next   int
}

func (r *eachRows) Columns() []string { return []string{"id"} }
func (r *eachRows) Close() error {
	r.driver.closed.Store(true)
	return nil
}
func (r *eachRows) Next(dest []driver.Value) error {
	if r.next >= r.driver.rows {
		return io.EOF
	}
	r.next++
	dest[0] = int64(r.next)
	return nil
}

var eachDriverID atomic.Int64

// newEachRows returns the result of a query on a new eachDriver, and a function that closes its database.
func newEachRows(t *testing.T, rows int) (*sqlx.Rows, *eachDriver, func()) {
	fake := &eachDriver{rows: rows}
	name := fmt.Sprintf("goravel-db-each-%d", eachDriverID.Add(1))
	sql.Register(name, fake)

	sqlDB, err := sql.Open(name, "")
	assert.NoError(t, err)

	result, err := sqlx.NewDb(sqlDB, name).Queryx("SELECT id FROM users")
	assert.NoError(t, err)

	return result, fake, func() { _ = sqlDB.Close() }
}

// newEachQuery expects a single query fetching the first chunk of 1000 rows.
func newEachQuery(t *testing.T, rows int) *Query {
	result, _, closeDB := newEachRows(t, rows)
	t.Cleanup(closeDB)

	ctx := context.Background()
	query := "SELECT * FROM users LIMIT 1000 OFFSET 0"

	mockGrammar := mocksdriver.NewGrammar(t)
	mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	mockReadBuilder := mocksdb.NewBuilder(t)
	mockReadBuilder.EXPECT().QueryxContext(ctx, query).Return(result, nil).Once()
	mockReadBuilder.EXPECT().Explain(query).Return(query).Once()
	mockLogger := mockslogger.NewLogger(t)
	mockLogger.EXPECT().Trace(ctx, mock.Anything, query, mock.Anything, nil).Return().Once()

	return NewQuery(ctx, mockReadBuilder, mocksdb.NewBuilder(t), mockGrammar, mockLogger, "users", nil)
}

func TestEachFetchesRowsInChunks(t *testing.T) {
	var ids []any
	err := newEachQuery(t, 3).Each(func(row contractsdb.Row) error {
		var data map[string]any
		if err := row.Scan(&data); err != nil {
			return err
		}
		ids = append(ids, data["id"])
		return nil
	})

	// One query (the mock expects it once) for all the rows, in order.
	assert.NoError(t, err)
	assert.Equal(t, []any{int64(1), int64(2), int64(3)}, ids)
}

func TestEachStopsOnCallbackError(t *testing.T) {
	calls := 0
	err := newEachQuery(t, 3).Each(func(row contractsdb.Row) error {
		calls++
		return assert.AnError
	})

	assert.ErrorIs(t, err, assert.AnError)
	assert.Equal(t, 1, calls)
}
