package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	mocksdb "github.com/goravel/framework/mocks/database/db"
)

type stubResult struct{}

func (stubResult) LastInsertId() (int64, error) { return 0, nil }
func (stubResult) RowsAffected() (int64, error) { return 2, nil }

type BuilderTestSuite struct {
	suite.Suite
	exporter   *recordingSpanExporter
	instrument *Instrument
}

func TestBuilderTestSuite(t *testing.T) {
	suite.Run(t, &BuilderTestSuite{})
}

func (s *BuilderTestSuite) SetupTest() {
	exporter, resolver := setupTelemetry(s.T())
	s.exporter = exporter
	s.instrument = NewInstrument(testPool(), "postgres", resolver)
}

func (s *BuilderTestSuite) lastSpan() sdktrace.ReadOnlySpan {
	s.Require().NotEmpty(s.exporter.spans)
	return s.exporter.spans[len(s.exporter.spans)-1]
}

func (s *BuilderTestSuite) TestSelectContext_StructuredQuerySpan() {
	inner := mocksdb.NewBuilder(s.T())
	inner.EXPECT().SelectContext(mock.Anything, mock.Anything, "SELECT * FROM users WHERE id = ?", 1).Return(nil).Once()

	wrapped := WrapBuilder(inner, s.instrument)
	ctx := ContextWithTable(context.Background(), "users")

	s.NoError(wrapped.SelectContext(ctx, &[]any{}, "SELECT * FROM users WHERE id = ?", 1))

	span := s.lastSpan()
	s.Equal("SELECT users", span.Name())
	collection, ok := attrValue(span, "db.collection.name")
	s.True(ok)
	s.Equal("users", collection)
	query, ok := attrValue(span, "db.query.text")
	s.True(ok)
	s.Contains(query, "?")
}

func (s *BuilderTestSuite) TestExecContext_RawQueryHasNoCollection() {
	inner := mocksdb.NewBuilder(s.T())
	inner.EXPECT().ExecContext(mock.Anything, "UPDATE users SET name = ?", "x").Return(stubResult{}, nil).Once()

	wrapped := WrapBuilder(inner, s.instrument)

	_, err := wrapped.ExecContext(context.Background(), "UPDATE users SET name = ?", "x")
	s.NoError(err)

	span := s.lastSpan()
	s.Equal("UPDATE", span.Name())
	_, ok := attrValue(span, "db.collection.name")
	s.False(ok)
}

func (s *BuilderTestSuite) TestExecContext_RecordsRowsAffected() {
	inner := mocksdb.NewBuilder(s.T())
	inner.EXPECT().ExecContext(mock.Anything, "INSERT INTO users (name) VALUES (?)", "x").Return(stubResult{}, nil).Once()

	wrapped := WrapBuilder(inner, s.instrument)

	_, err := wrapped.ExecContext(context.Background(), "INSERT INTO users (name) VALUES (?)", "x")
	s.NoError(err)

	rows, ok := attrValue(s.lastSpan(), "db.response.returned_rows")
	s.True(ok)
	s.Equal("2", rows)
}

func (s *BuilderTestSuite) TestExecContext_RecordsError() {
	inner := mocksdb.NewBuilder(s.T())
	inner.EXPECT().ExecContext(mock.Anything, "UPDATE users SET name = ?", "x").Return(nil, assert.AnError).Once()

	wrapped := WrapBuilder(inner, s.instrument)

	_, err := wrapped.ExecContext(context.Background(), "UPDATE users SET name = ?", "x")
	s.ErrorIs(err, assert.AnError)
	s.Equal(codes.Error, s.lastSpan().Status().Code)
}

func (s *BuilderTestSuite) TestQueryxContext_Span() {
	inner := mocksdb.NewBuilder(s.T())
	inner.EXPECT().QueryxContext(mock.Anything, "SELECT * FROM users").Return(nil, nil).Once()

	wrapped := WrapBuilder(inner, s.instrument)
	ctx := ContextWithTable(context.Background(), "users")

	_, err := wrapped.QueryxContext(ctx, "SELECT * FROM users")
	s.NoError(err)
	s.Equal("SELECT users", s.lastSpan().Name())
}

func (s *BuilderTestSuite) TestTxBuilder_ExecContext() {
	inner := mocksdb.NewTxBuilder(s.T())
	inner.EXPECT().ExecContext(mock.Anything, "UPDATE users SET name = ?", "x").Return(stubResult{}, nil).Once()

	wrapped := WrapTxBuilder(inner, s.instrument)
	ctx := ContextWithTable(context.Background(), "users")

	_, err := wrapped.ExecContext(ctx, "UPDATE users SET name = ?", "x")
	s.NoError(err)
	s.Equal("UPDATE users", s.lastSpan().Name())
}
