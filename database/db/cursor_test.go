package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksdriver "github.com/goravel/framework/mocks/database/driver"
	mockslogger "github.com/goravel/framework/mocks/database/logger"
)

// cursorDriver is a database/sql driver whose queries return `rows` rows with an
// "id" column, and that records when the rows are closed.
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

func newCursorRows(t *testing.T, rows int) (*sqlx.Rows, *cursorDriver) {
	fake := &cursorDriver{rows: rows}
	name := fmt.Sprintf("goravel-db-cursor-%d", cursorDriverID.Add(1))
	sql.Register(name, fake)

	sqlDB, err := sql.Open(name, "")
	assert.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	result, err := sqlx.NewDb(sqlDB, name).Queryx("SELECT id FROM users")
	assert.NoError(t, err)

	return result, fake
}

func newCursorQuery(t *testing.T, ctx context.Context, rows *sqlx.Rows, queryErr error) *Query {
	mockGrammar := mocksdriver.NewGrammar(t)
	mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	mockReadBuilder := mocksdb.NewBuilder(t)
	mockReadBuilder.EXPECT().QueryxContext(ctx, "SELECT * FROM users").Return(rows, queryErr).Once()
	mockReadBuilder.EXPECT().Explain("SELECT * FROM users").Return("SELECT * FROM users").Once()
	mockLogger := mockslogger.NewLogger(t)
	mockLogger.EXPECT().Trace(ctx, mock.Anything, "SELECT * FROM users", mock.Anything, queryErr).Return().Once()

	return NewQuery(ctx, mockReadBuilder, mocksdb.NewBuilder(t), mockGrammar, mockLogger, "users", nil)
}

func TestCursor(t *testing.T) {
	rows, fake := newCursorRows(t, 3)
	query := newCursorQuery(t, context.Background(), rows, nil)

	var ids []any
	for row := range query.Cursor() {
		assert.NoError(t, row.Err())
		var data map[string]any
		assert.NoError(t, row.Scan(&data))
		ids = append(ids, data["id"])
	}

	assert.Equal(t, []any{int64(1), int64(2), int64(3)}, ids)
	assert.True(t, fake.closed.Load())
}

func TestCursorQueryError(t *testing.T) {
	query := newCursorQuery(t, context.Background(), nil, assert.AnError)

	var errs []error
	for row := range query.Cursor() {
		errs = append(errs, row.Err())
	}

	assert.Equal(t, []error{assert.AnError}, errs)
}

func TestCursorReleasesTheRowsWhenTheContextIsCanceled(t *testing.T) {
	rows, fake := newCursorRows(t, 3)
	ctx, cancel := context.WithCancel(context.Background())
	query := newCursorQuery(t, ctx, rows, nil)

	ch := query.Cursor()
	first := <-ch
	assert.NoError(t, first.Err())

	// The consumer stops early: canceling the context lets the producer exit,
	// close the rows (releasing the connection) and close the channel.
	cancel()

	assert.Eventually(t, fake.closed.Load, time.Second, time.Millisecond)
	assert.Eventually(t, func() bool {
		select {
		case _, ok := <-ch:
			return !ok
		default:
			return false
		}
	}, time.Second, time.Millisecond)
}
