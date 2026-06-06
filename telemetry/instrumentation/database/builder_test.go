package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/codes"

	mocksdb "github.com/goravel/framework/mocks/database/db"
	"github.com/goravel/framework/telemetry"
)

func TestTracedBuilder_SelectContext(t *testing.T) {
	exporter := setupRecordingTelemetry(t)

	inner := mocksdb.NewCommonBuilder(t)
	inner.EXPECT().SelectContext(mock.Anything, mock.Anything, "SELECT * FROM users WHERE id = ?", 1).Return(nil).Once()

	builder := WrapBuilder(inner, "postgres")
	var dest []any
	assert.NoError(t, builder.SelectContext(context.Background(), &dest, "SELECT * FROM users WHERE id = ?", 1))

	assert.Len(t, exporter.spans, 1)
	assert.Equal(t, "SELECT", exporter.spans[0].Name())
}

func TestTracedBuilder_ExecContextError(t *testing.T) {
	exporter := setupRecordingTelemetry(t)

	inner := mocksdb.NewCommonBuilder(t)
	inner.EXPECT().ExecContext(mock.Anything, "DELETE FROM users").Return(nil, assert.AnError).Once()

	builder := WrapBuilder(inner, "postgres")
	_, err := builder.ExecContext(context.Background(), "DELETE FROM users")
	assert.Equal(t, assert.AnError, err)

	assert.Len(t, exporter.spans, 1)
	assert.Equal(t, codes.Error, exporter.spans[0].Status().Code)
}

func TestTracedBuilder_ExecContextSuccess(t *testing.T) {
	exporter := setupRecordingTelemetry(t)

	inner := mocksdb.NewCommonBuilder(t)
	inner.EXPECT().ExecContext(mock.Anything, "INSERT INTO users (name) VALUES (?)", "Alice").Return(&stubResult{rowsAffected: 1}, nil).Once()

	builder := WrapBuilder(inner, "mysql")
	result, err := builder.ExecContext(context.Background(), "INSERT INTO users (name) VALUES (?)", "Alice")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Len(t, exporter.spans, 1)
	assert.Equal(t, "INSERT", exporter.spans[0].Name())
}

func TestTracedBuilder_QueryxContext(t *testing.T) {
	exporter := setupRecordingTelemetry(t)

	inner := mocksdb.NewCommonBuilder(t)
	inner.EXPECT().QueryxContext(mock.Anything, "SELECT id FROM orders", mock.Anything).Return(nil, nil).Once()

	builder := WrapBuilder(inner, "postgres")
	rows, err := builder.QueryxContext(context.Background(), "SELECT id FROM orders")
	assert.NoError(t, err)
	assert.Nil(t, rows)

	assert.Len(t, exporter.spans, 1)
	assert.Equal(t, "SELECT", exporter.spans[0].Name())
}

func TestTracedBuilder_ExplainPassthrough(t *testing.T) {
	inner := mocksdb.NewCommonBuilder(t)
	inner.EXPECT().Explain("SELECT ?", 1).Return("SELECT 1").Once()

	builder := WrapBuilder(inner, "postgres")
	result := builder.Explain("SELECT ?", 1)
	assert.Equal(t, "SELECT 1", result)
}

func TestTracedBuilder_PassthroughWithoutFacade(t *testing.T) {
	original := telemetry.Facade
	telemetry.Facade = nil
	t.Cleanup(func() { telemetry.Facade = original })

	inner := mocksdb.NewCommonBuilder(t)
	inner.EXPECT().GetContext(mock.Anything, mock.Anything, "SELECT 1").Return(nil).Once()

	builder := WrapBuilder(inner, "postgres")
	assert.NoError(t, builder.GetContext(context.Background(), nil, "SELECT 1"))
}

type stubResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (s *stubResult) LastInsertId() (int64, error) { return s.lastInsertID, nil }
func (s *stubResult) RowsAffected() (int64, error) { return s.rowsAffected, nil }
