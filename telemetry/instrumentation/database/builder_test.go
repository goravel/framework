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

type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

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
	var users []any
	inner := mocksdb.NewBuilder(s.T())
	// Context uses mock.Anything because the tracing wrapper injects a span into it.
	inner.EXPECT().SelectContext(mock.Anything, &users, "SELECT * FROM users WHERE id = ?", 1).Return(nil).Once()

	wrapped := WrapBuilder(inner, s.instrument)
	ctx := ContextWithTable(context.Background(), "users")

	s.NoError(wrapped.SelectContext(ctx, &users, "SELECT * FROM users WHERE id = ?", 1))

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
	mockResult := &MockResult{}
	mockResult.On("RowsAffected").Return(int64(1), nil)

	inner := mocksdb.NewBuilder(s.T())
	inner.EXPECT().ExecContext(mock.Anything, "UPDATE users SET name = ?", "x").Return(mockResult, nil).Once()

	wrapped := WrapBuilder(inner, s.instrument)

	_, err := wrapped.ExecContext(context.Background(), "UPDATE users SET name = ?", "x")
	s.NoError(err)

	span := s.lastSpan()
	s.Equal("UPDATE", span.Name())
	_, ok := attrValue(span, "db.collection.name")
	s.False(ok)
	mockResult.AssertExpectations(s.T())
}

func (s *BuilderTestSuite) TestExecContext_RecordsRowsAffected() {
	mockResult := &MockResult{}
	mockResult.On("RowsAffected").Return(int64(2), nil)

	inner := mocksdb.NewBuilder(s.T())
	inner.EXPECT().ExecContext(mock.Anything, "INSERT INTO users (name) VALUES (?)", "x").Return(mockResult, nil).Once()

	wrapped := WrapBuilder(inner, s.instrument)

	_, err := wrapped.ExecContext(context.Background(), "INSERT INTO users (name) VALUES (?)", "x")
	s.NoError(err)

	rows, ok := attrValue(s.lastSpan(), "db.response.returned_rows")
	s.True(ok)
	s.Equal("2", rows)
	mockResult.AssertExpectations(s.T())
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
	mockResult := &MockResult{}
	mockResult.On("RowsAffected").Return(int64(1), nil)

	inner := mocksdb.NewTxBuilder(s.T())
	inner.EXPECT().ExecContext(mock.Anything, "UPDATE users SET name = ?", "x").Return(mockResult, nil).Once()

	wrapped := WrapTxBuilder(inner, s.instrument)
	ctx := ContextWithTable(context.Background(), "users")

	_, err := wrapped.ExecContext(ctx, "UPDATE users SET name = ?", "x")
	s.NoError(err)
	s.Equal("UPDATE users", s.lastSpan().Name())
	mockResult.AssertExpectations(s.T())
}
