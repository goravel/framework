package database

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"

	contractsdb "github.com/goravel/framework/contracts/database/db"
)

type fakeBuilder struct {
	execErr error
}

func (f *fakeBuilder) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return fakeResult{}, f.execErr
}
func (f *fakeBuilder) GetContext(_ context.Context, _ any, _ string, _ ...any) error    { return nil }
func (f *fakeBuilder) SelectContext(_ context.Context, _ any, _ string, _ ...any) error { return nil }
func (f *fakeBuilder) QueryxContext(_ context.Context, _ string, _ ...any) (*sqlx.Rows, error) {
	return nil, nil
}
func (f *fakeBuilder) Explain(_ string, _ ...any) string { return "" }
func (f *fakeBuilder) Beginx() (*sqlx.Tx, error)         { return nil, nil }

type fakeResult struct{}

func (fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (fakeResult) RowsAffected() (int64, error) { return 2, nil }

func TestWrapBuilder_NilInstrumentPassesThrough(t *testing.T) {
	inner := &fakeBuilder{}
	assert.Equal(t, contractsdb.Builder(inner), WrapBuilder(inner, nil))
}

func TestWrapBuilder_StructuredQuerySpan(t *testing.T) {
	exporter := setupTelemetry(t, true)
	inst := NewInstrument(testPool(), "postgres")

	wrapped := WrapBuilder(&fakeBuilder{}, inst)
	ctx := ContextWithTable(context.Background(), "users")

	assert.NoError(t, wrapped.SelectContext(ctx, &[]any{}, "SELECT * FROM users WHERE id = ?", 1))

	assert.Len(t, exporter.spans, 1)
	span := exporter.spans[0]
	assert.Equal(t, "SELECT users", span.Name())
	collection, ok := attrValue(span, "db.collection.name")
	assert.True(t, ok)
	assert.Equal(t, "users", collection)
	query, ok := attrValue(span, "db.query.text")
	assert.True(t, ok)
	assert.Contains(t, query, "?")
}

func TestWrapBuilder_RawQueryHasNoCollection(t *testing.T) {
	exporter := setupTelemetry(t, true)
	inst := NewInstrument(testPool(), "postgres")

	wrapped := WrapBuilder(&fakeBuilder{}, inst)

	_, err := wrapped.ExecContext(context.Background(), "UPDATE users SET name = ?", "x")
	assert.NoError(t, err)

	assert.Len(t, exporter.spans, 1)
	span := exporter.spans[0]
	assert.Equal(t, "UPDATE", span.Name())
	_, ok := attrValue(span, "db.collection.name")
	assert.False(t, ok)
}

func TestWrapBuilder_ExecRecordsRows(t *testing.T) {
	exporter := setupTelemetry(t, true)
	inst := NewInstrument(testPool(), "postgres")

	wrapped := WrapBuilder(&fakeBuilder{}, inst)

	_, err := wrapped.ExecContext(context.Background(), "INSERT INTO users (name) VALUES (?)", "x")
	assert.NoError(t, err)

	assert.Len(t, exporter.spans, 1)
	rows, ok := attrValue(exporter.spans[0], "db.response.returned_rows")
	assert.True(t, ok)
	assert.Equal(t, "2", rows)
}

func TestWrapBuilder_RecordsError(t *testing.T) {
	exporter := setupTelemetry(t, true)
	inst := NewInstrument(testPool(), "postgres")

	wrapped := WrapBuilder(&fakeBuilder{execErr: assert.AnError}, inst)

	_, err := wrapped.ExecContext(context.Background(), "UPDATE users SET name = ?", "x")
	assert.ErrorIs(t, err, assert.AnError)

	assert.Len(t, exporter.spans, 1)
	assert.Equal(t, codes.Error, exporter.spans[0].Status().Code)
}

func TestWrapBuilder_QueryxSpan(t *testing.T) {
	exporter := setupTelemetry(t, true)
	inst := NewInstrument(testPool(), "postgres")

	wrapped := WrapBuilder(&fakeBuilder{}, inst)
	ctx := ContextWithTable(context.Background(), "users")

	_, err := wrapped.QueryxContext(ctx, "SELECT * FROM users")
	assert.NoError(t, err)

	assert.Len(t, exporter.spans, 1)
	assert.Equal(t, "SELECT users", exporter.spans[0].Name())
}

func TestWrapTxBuilder(t *testing.T) {
	t.Run("nil passes through", func(t *testing.T) {
		inner := &fakeTxBuilder{}
		assert.Equal(t, contractsdb.TxBuilder(inner), WrapTxBuilder(inner, nil))
	})

	t.Run("records span", func(t *testing.T) {
		exporter := setupTelemetry(t, true)
		inst := NewInstrument(testPool(), "postgres")

		wrapped := WrapTxBuilder(&fakeTxBuilder{}, inst)
		ctx := ContextWithTable(context.Background(), "users")

		_, err := wrapped.ExecContext(ctx, "UPDATE users SET name = ?", "x")
		assert.NoError(t, err)

		assert.Len(t, exporter.spans, 1)
		assert.Equal(t, "UPDATE users", exporter.spans[0].Name())
	})
}

type fakeTxBuilder struct{}

func (f *fakeTxBuilder) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return fakeResult{}, nil
}
func (f *fakeTxBuilder) GetContext(_ context.Context, _ any, _ string, _ ...any) error    { return nil }
func (f *fakeTxBuilder) SelectContext(_ context.Context, _ any, _ string, _ ...any) error { return nil }
func (f *fakeTxBuilder) QueryxContext(_ context.Context, _ string, _ ...any) (*sqlx.Rows, error) {
	return nil, nil
}
func (f *fakeTxBuilder) Explain(_ string, _ ...any) string { return "" }
func (f *fakeTxBuilder) Commit() error                     { return nil }
func (f *fakeTxBuilder) Rollback() error                   { return nil }
