package gorm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"
	gormtests "gorm.io/gorm/utils/tests"

	contractsdatabase "github.com/goravel/framework/contracts/database"
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

func newCursorQuery(t *testing.T, ctx context.Context, rows int) (*Query, *cursorDriver) {
	fake := &cursorDriver{rows: rows}
	name := fmt.Sprintf("goravel-orm-cursor-%d", cursorDriverID.Add(1))
	sql.Register(name, fake)

	sqlDB, err := sql.Open(name, "")
	assert.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{ConnPool: sqlDB})
	assert.NoError(t, err)

	query := NewQuery(ctx, nil, contractsdatabase.Config{}, db.WithContext(ctx), nil, nil, nil, nil, nil)

	return query.Table("users").(*Query), fake
}

func TestCursor(t *testing.T) {
	query, fake := newCursorQuery(t, context.Background(), 3)

	count := 0
	for row := range query.Cursor() {
		assert.NoError(t, row.Err())
		count++
	}

	assert.Equal(t, 3, count)
	assert.True(t, fake.closed.Load())
}

func TestCursorReleasesTheRowsWhenTheContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	query, fake := newCursorQuery(t, ctx, 3)

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
