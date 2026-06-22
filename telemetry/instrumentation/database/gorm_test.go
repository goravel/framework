package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"gorm.io/gorm"
	gormtests "gorm.io/gorm/utils/tests"
)

type testUser struct {
	ID   uint
	Name string
}

type GormPluginTestSuite struct {
	suite.Suite
	db       *gorm.DB
	exporter *recordingSpanExporter
}

func TestGormPluginTestSuite(t *testing.T) {
	suite.Run(t, &GormPluginTestSuite{})
}

func (s *GormPluginTestSuite) SetupTest() {
	exporter, resolver := setupTelemetry(s.T())

	instrument := NewInstrument(testPool(), "postgres", resolver)
	plugin := NewGormPlugin(instrument)
	s.Require().NotNil(plugin)

	db, err := gorm.Open(gormtests.DummyDialector{}, &gorm.Config{SkipDefaultTransaction: true, DryRun: true})
	s.Require().NoError(err)
	s.Require().NoError(db.Use(plugin))

	s.db = db
	s.exporter = exporter
}

func (s *GormPluginTestSuite) lastSpan() sdktrace.ReadOnlySpan {
	s.Require().NotEmpty(s.exporter.spans)
	return s.exporter.spans[len(s.exporter.spans)-1]
}

func (s *GormPluginTestSuite) TestInactiveWhenResolverIsNil() {
	instrument := NewInstrument(testPool(), "postgres", nil)
	plugin := NewGormPlugin(instrument)
	s.NotNil(plugin)
	s.False(plugin.instrument.active())
}

func (s *GormPluginTestSuite) TestQuerySpan() {
	var users []testUser
	s.db.WithContext(context.Background()).Where("name = ?", "Goravel").Find(&users)

	s.Require().Len(s.exporter.spans, 1)
	span := s.lastSpan()
	s.Equal("SELECT test_users", span.Name())

	for key, expected := range map[string]string{
		"db.collection.name":             "test_users",
		"db.namespace":                   "app",
		"db.client.connection.pool.name": "postgres",
	} {
		val, ok := attrValue(span, key)
		s.True(ok, key)
		s.Equal(expected, val, key)
	}

	query, ok := attrValue(span, "db.query.text")
	s.True(ok)
	s.Contains(query, "?")
	s.NotContains(query, "Goravel")
}

func (s *GormPluginTestSuite) TestCreateSpan() {
	s.db.WithContext(context.Background()).Create(&testUser{Name: "Goravel"})

	s.Require().Len(s.exporter.spans, 1)
	s.Equal("INSERT test_users", s.lastSpan().Name())
}

func (s *GormPluginTestSuite) TestSequentialQueriesAreSiblings() {
	ctx := context.Background()

	var users []testUser
	s.db.WithContext(ctx).Find(&users)
	s.db.WithContext(ctx).Find(&users)

	s.Require().Len(s.exporter.spans, 2)
	s.False(s.exporter.spans[1].Parent().IsValid())
}

func (s *GormPluginTestSuite) TestNestsUnderParentSpan() {
	ctx, parent := s.exporter.tracer.Start(context.Background(), "parent")
	var users []testUser
	s.db.WithContext(ctx).Find(&users)
	parent.End()

	s.Require().Len(s.exporter.spans, 2)
	s.Equal(parent.SpanContext().SpanID(), s.exporter.spans[0].Parent().SpanID())
}
