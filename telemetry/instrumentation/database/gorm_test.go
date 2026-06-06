package database

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"gorm.io/gorm"
	gormtests "gorm.io/gorm/utils/tests"

	"github.com/goravel/framework/telemetry"
)

type testUser struct {
	ID   uint
	Name string
}

func setupTracedGorm(t *testing.T) (*gorm.DB, *recordingSpanExporter) {
	t.Helper()

	exporter := setupRecordingTelemetry(t)

	db, err := gorm.Open(gormtests.DummyDialector{}, &gorm.Config{SkipDefaultTransaction: true, DryRun: true})
	assert.NoError(t, err)
	assert.NoError(t, db.Use(NewGormPlugin()))

	return db, exporter
}

func assertAttr(t *testing.T, span sdktrace.ReadOnlySpan, key string) {
	t.Helper()

	for _, attr := range span.Attributes() {
		if string(attr.Key) == key {
			return
		}
	}

	t.Fatalf("expected span to have attribute %q", key)
}

func assertNoBoundValues(t *testing.T, span sdktrace.ReadOnlySpan) {
	t.Helper()

	for _, attr := range span.Attributes() {
		if string(attr.Key) != "db.query.text" {
			continue
		}

		query := attr.Value.AsString()
		assert.Contains(t, query, "?", "query text should keep placeholders")
		assert.NotContains(t, query, "Goravel", "query text must not contain interpolated values")
		return
	}

	t.Fatal("expected span to have db.query.text attribute")
}

func TestGormPlugin_QuerySpan(t *testing.T) {
	db, exporter := setupTracedGorm(t)

	var users []testUser
	db.WithContext(context.Background()).Where("name = ?", "Goravel").Find(&users)

	assert.Len(t, exporter.spans, 1)
	span := exporter.spans[0]
	assert.Equal(t, "SELECT test_users", span.Name())
	assertAttr(t, span, "db.query.text")
	assertNoBoundValues(t, span)
}

func TestGormPlugin_CreateSpan(t *testing.T) {
	db, exporter := setupTracedGorm(t)

	db.WithContext(context.Background()).Create(&testUser{Name: "Goravel"})

	assert.Len(t, exporter.spans, 1)
	span := exporter.spans[0]
	assert.Equal(t, "INSERT test_users", span.Name())
	assert.True(t, strings.HasPrefix(span.Name(), "INSERT "))
	assertAttr(t, span, "db.query.text")
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

func TestGormPlugin_NoFacadeNoSpans(t *testing.T) {
	original := telemetry.Facade
	telemetry.Facade = nil
	t.Cleanup(func() { telemetry.Facade = original })

	db, err := gorm.Open(gormtests.DummyDialector{}, &gorm.Config{SkipDefaultTransaction: true, DryRun: true})
	assert.NoError(t, err)
	assert.NoError(t, db.Use(NewGormPlugin()))

	var users []testUser
	assert.NotPanics(t, func() { db.WithContext(context.Background()).Find(&users) })
}
