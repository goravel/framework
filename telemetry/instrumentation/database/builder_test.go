package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/codes"

	mocksdb "github.com/goravel/framework/mocks/database/db"
)

// stubResult is a minimal sql.Result for exec returns. sql.Result is a standard
// library interface, so it has no generated mock to use here.
type stubResult struct{}

func (stubResult) LastInsertId() (int64, error) { return 0, nil }
func (stubResult) RowsAffected() (int64, error) { return 2, nil }

func TestWrapBuilder_StructuredQuerySpan(t *testing.T) {
	exporter := setupTelemetry(t, true)
	inst := NewInstrument(testPool(), "postgres")

	inner := mocksdb.NewBuilder(t)
	inner.EXPECT().SelectContext(mock.Anything, mock.Anything, "SELECT * FROM users WHERE id = ?", 1).Return(nil).Once()

	wrapped := WrapBuilder(inner, inst)
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

	inner := mocksdb.NewBuilder(t)
	inner.EXPECT().ExecContext(mock.Anything, "UPDATE users SET name = ?", "x").Return(stubResult{}, nil).Once()

	wrapped := WrapBuilder(inner, inst)

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

	inner := mocksdb.NewBuilder(t)
	inner.EXPECT().ExecContext(mock.Anything, "INSERT INTO users (name) VALUES (?)", "x").Return(stubResult{}, nil).Once()

	wrapped := WrapBuilder(inner, inst)

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

	inner := mocksdb.NewBuilder(t)
	inner.EXPECT().ExecContext(mock.Anything, "UPDATE users SET name = ?", "x").Return(nil, assert.AnError).Once()

	wrapped := WrapBuilder(inner, inst)

	_, err := wrapped.ExecContext(context.Background(), "UPDATE users SET name = ?", "x")
	assert.ErrorIs(t, err, assert.AnError)

	assert.Len(t, exporter.spans, 1)
	assert.Equal(t, codes.Error, exporter.spans[0].Status().Code)
}

func TestWrapBuilder_QueryxSpan(t *testing.T) {
	exporter := setupTelemetry(t, true)
	inst := NewInstrument(testPool(), "postgres")

	inner := mocksdb.NewBuilder(t)
	inner.EXPECT().QueryxContext(mock.Anything, "SELECT * FROM users").Return(nil, nil).Once()

	wrapped := WrapBuilder(inner, inst)
	ctx := ContextWithTable(context.Background(), "users")

	_, err := wrapped.QueryxContext(ctx, "SELECT * FROM users")
	assert.NoError(t, err)

	assert.Len(t, exporter.spans, 1)
	assert.Equal(t, "SELECT users", exporter.spans[0].Name())
}

func TestWrapTxBuilder(t *testing.T) {
	exporter := setupTelemetry(t, true)
	inst := NewInstrument(testPool(), "postgres")

	inner := mocksdb.NewTxBuilder(t)
	inner.EXPECT().ExecContext(mock.Anything, "UPDATE users SET name = ?", "x").Return(stubResult{}, nil).Once()

	wrapped := WrapTxBuilder(inner, inst)
	ctx := ContextWithTable(context.Background(), "users")

	_, err := wrapped.ExecContext(ctx, "UPDATE users SET name = ?", "x")
	assert.NoError(t, err)

	assert.Len(t, exporter.spans, 1)
	assert.Equal(t, "UPDATE users", exporter.spans[0].Name())
}
