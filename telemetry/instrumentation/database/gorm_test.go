package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"gorm.io/gorm"
	gormtests "gorm.io/gorm/utils/tests"

	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
)

type testUser struct {
	ID   uint
	Name string
}

func setupTracedGorm(t *testing.T) (*gorm.DB, *recordingSpanExporter) {
	t.Helper()

	exporter, mockConfig, resolver := setupTelemetry(t, true)

	plugin := NewGormPlugin(testPool(), "postgres", mockConfig, resolver)
	assert.NotNil(t, plugin)

	db, err := gorm.Open(gormtests.DummyDialector{}, &gorm.Config{SkipDefaultTransaction: true, DryRun: true})
	assert.NoError(t, err)
	assert.NoError(t, db.Use(plugin))

	return db, exporter
}

func assertAttr(t *testing.T, span sdktrace.ReadOnlySpan, key, expected string) {
	t.Helper()

	value, ok := attrValue(span, key)
	assert.True(t, ok, key)
	assert.Equal(t, expected, value, key)
}

func TestNewGormPlugin(t *testing.T) {
	t.Run("inactive when config is nil", func(t *testing.T) {
		plugin := NewGormPlugin(testPool(), "postgres", nil, func() contractstelemetry.Telemetry { return nil })
		assert.NotNil(t, plugin)
		assert.False(t, plugin.instrument.active())
	})

	t.Run("inactive when disabled", func(t *testing.T) {
		_, mockConfig, resolver := setupTelemetry(t, false)
		plugin := NewGormPlugin(testPool(), "postgres", mockConfig, resolver)
		assert.NotNil(t, plugin)
		assert.False(t, plugin.instrument.active())
	})
}

func TestGormPlugin_QuerySpan(t *testing.T) {
	db, exporter := setupTracedGorm(t)

	var users []testUser
	db.WithContext(context.Background()).Where("name = ?", "Goravel").Find(&users)

	assert.Len(t, exporter.spans, 1)
	span := exporter.spans[0]
	assert.Equal(t, "SELECT test_users", span.Name())
	assertAttr(t, span, "db.collection.name", "test_users")
	assertAttr(t, span, "db.namespace", "app")
	assertAttr(t, span, "db.client.connection.pool.name", "postgres")

	query, ok := attrValue(span, "db.query.text")
	assert.True(t, ok)
	assert.Contains(t, query, "?", "query text should keep placeholders")
	assert.NotContains(t, query, "Goravel", "query text must not contain bound values")
}

func TestGormPlugin_CreateSpan(t *testing.T) {
	db, exporter := setupTracedGorm(t)

	db.WithContext(context.Background()).Create(&testUser{Name: "Goravel"})

	assert.Len(t, exporter.spans, 1)
	assert.Equal(t, "INSERT test_users", exporter.spans[0].Name())
}

func TestGormPlugin_SequentialQueriesAreSiblings(t *testing.T) {
	db, exporter := setupTracedGorm(t)
	ctx := context.Background()

	var users []testUser
	db.WithContext(ctx).Find(&users)
	db.WithContext(ctx).Find(&users)

	assert.Len(t, exporter.spans, 2)
	assert.False(t, exporter.spans[1].Parent().IsValid(), "second span must not be a child of the first")
}

func TestGormPlugin_NestsUnderParentSpan(t *testing.T) {
	db, exporter := setupTracedGorm(t)

	ctx, parent := exporter.tracer.Start(context.Background(), "parent")
	var users []testUser
	db.WithContext(ctx).Find(&users)
	parent.End()

	assert.Len(t, exporter.spans, 2)
	assert.Equal(t, parent.SpanContext().SpanID(), exporter.spans[0].Parent().SpanID())
}
