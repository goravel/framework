package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"gorm.io/gorm"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/telemetry"
)

type recordingSpanExporter struct {
	mu    sync.Mutex
	spans []sdktrace.ReadOnlySpan
}

func (r *recordingSpanExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spans = append(r.spans, spans...)
	return nil
}

func (r *recordingSpanExporter) Shutdown(_ context.Context) error { return nil }

func setupTelemetry(t *testing.T, enabled bool) *recordingSpanExporter {
	t.Helper()

	exporter := &recordingSpanExporter{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	mockTelemetry := mockstelemetry.NewTelemetry(t)
	mockTelemetry.EXPECT().Tracer(instrumentationName).Return(provider.Tracer(instrumentationName)).Maybe()
	mockTelemetry.EXPECT().Meter(instrumentationName).Return(metricnoop.NewMeterProvider().Meter(instrumentationName)).Maybe()

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetBool(enabledConfigKey, true).Return(enabled).Maybe()

	originalFacade, originalConfig := telemetry.Facade, telemetry.ConfigFacade
	telemetry.Facade, telemetry.ConfigFacade = mockTelemetry, mockConfig
	t.Cleanup(func() { telemetry.Facade, telemetry.ConfigFacade = originalFacade, originalConfig })

	return exporter
}

func testPool() contractsdatabase.Pool {
	return contractsdatabase.Pool{
		Writers: []contractsdatabase.Config{
			{Driver: "postgres", Database: "app", Host: "db.local", Port: 5432},
		},
	}
}

func attrValue(span sdktrace.ReadOnlySpan, key string) (string, bool) {
	for _, attr := range span.Attributes() {
		if string(attr.Key) == key {
			return attr.Value.Emit(), true
		}
	}
	return "", false
}

func TestNewInstrument(t *testing.T) {
	t.Run("nil when facade is not set", func(t *testing.T) {
		original := telemetry.Facade
		telemetry.Facade = nil
		t.Cleanup(func() { telemetry.Facade = original })

		assert.Nil(t, newInstrument(testPool(), "postgres"))
	})

	t.Run("nil when disabled", func(t *testing.T) {
		setupTelemetry(t, false)

		assert.Nil(t, newInstrument(testPool(), "postgres"))
	})

	t.Run("captures base attributes when enabled", func(t *testing.T) {
		setupTelemetry(t, true)

		inst := newInstrument(testPool(), "postgres")

		assert.NotNil(t, inst)
		assert.Equal(t, []telemetry.KeyValue{
			semconv.DBSystemNamePostgreSQL,
			semconv.DBNamespace("app"),
			semconv.ServerAddress("db.local"),
			semconv.ServerPort(5432),
			semconv.DBClientConnectionPoolName("postgres"),
		}, inst.baseAttrs)
	})

	t.Run("skips empty connection details", func(t *testing.T) {
		setupTelemetry(t, true)

		inst := newInstrument(contractsdatabase.Pool{Writers: []contractsdatabase.Config{{Driver: "sqlite"}}}, "")

		assert.Equal(t, []telemetry.KeyValue{semconv.DBSystemNameSQLite}, inst.baseAttrs)
	})
}

func TestInstrument_EndSpan(t *testing.T) {
	t.Run("records query attributes and renames span", func(t *testing.T) {
		exporter := setupTelemetry(t, true)
		inst := newInstrument(testPool(), "postgres")

		ctx, span := inst.startSpan(context.Background(), "gorm.Query")
		inst.endSpan(ctx, span, time.Now(), "SELECT * FROM users WHERE id = ?", "users", 3, nil)

		assert.Len(t, exporter.spans, 1)
		recorded := exporter.spans[0]
		assert.Equal(t, "SELECT users", recorded.Name())
		assert.Equal(t, codes.Unset, recorded.Status().Code)

		for key, expected := range map[string]string{
			"db.system.name":            "postgresql",
			"db.namespace":              "app",
			"server.address":            "db.local",
			"server.port":               "5432",
			"db.operation.name":         "SELECT",
			"db.query.text":             "SELECT * FROM users WHERE id = ?",
			"db.collection.name":        "users",
			"db.query.summary":          "SELECT users",
			"db.response.returned_rows": "3",
		} {
			value, ok := attrValue(recorded, key)
			assert.True(t, ok, key)
			assert.Equal(t, expected, value, key)
		}
	})

	t.Run("records error status", func(t *testing.T) {
		exporter := setupTelemetry(t, true)
		inst := newInstrument(testPool(), "postgres")

		ctx, span := inst.startSpan(context.Background(), "gorm.Query")
		inst.endSpan(ctx, span, time.Now(), "SELECT * FROM users", "users", -1, assert.AnError)

		assert.Equal(t, codes.Error, exporter.spans[0].Status().Code)
	})

	t.Run("ignores record not found", func(t *testing.T) {
		exporter := setupTelemetry(t, true)
		inst := newInstrument(testPool(), "postgres")

		ctx, span := inst.startSpan(context.Background(), "gorm.Query")
		inst.endSpan(ctx, span, time.Now(), "SELECT * FROM users", "users", -1, gorm.ErrRecordNotFound)

		assert.Equal(t, codes.Unset, exporter.spans[0].Status().Code)
	})
}

func TestDBSystem(t *testing.T) {
	tests := []struct {
		driver   string
		expected string
	}{
		{"postgres", "postgresql"},
		{"postgresql", "postgresql"},
		{"mysql", "mysql"},
		{"sqlite", "sqlite"},
		{"sqlserver", "microsoft.sql_server"},
		{"alien", "alien"},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			assert.Equal(t, semconv.DBSystemNameKey, dbSystem(tt.driver).Key)
			assert.Equal(t, tt.expected, dbSystem(tt.driver).Value.AsString())
		})
	}
}

func TestOperationName(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{"SELECT * FROM users", "SELECT"},
		{"  insert into users values (?)", "INSERT"},
		{"UPDATE users SET name = ?", "UPDATE"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, operationName(tt.query))
		})
	}
}

func TestIsRecordableError(t *testing.T) {
	assert.False(t, isRecordableError(nil))
	assert.False(t, isRecordableError(gorm.ErrRecordNotFound))
	assert.False(t, isRecordableError(sql.ErrNoRows))
	assert.False(t, isRecordableError(driver.ErrSkip))
	assert.False(t, isRecordableError(io.EOF))
	assert.True(t, isRecordableError(assert.AnError))
}
