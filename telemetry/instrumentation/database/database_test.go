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
	otellog "go.opentelemetry.io/otel/log"
	lognoop "go.opentelemetry.io/otel/log/noop"
	otelmetric "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/telemetry"
)

type recordingSpanExporter struct {
	mu    sync.Mutex
	spans []sdktrace.ReadOnlySpan
	tracer oteltrace.Tracer
}

func (r *recordingSpanExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spans = append(r.spans, spans...)
	return nil
}

func (r *recordingSpanExporter) Shutdown(_ context.Context) error { return nil }

type telemetryStub struct {
	traceProvider *sdktrace.TracerProvider
	meterProvider otelmetric.MeterProvider
}

func (s *telemetryStub) ForceFlush(_ context.Context) error                                          { return nil }
func (s *telemetryStub) Shutdown(_ context.Context) error                                            { return nil }
func (s *telemetryStub) Logger(_ string, _ ...otellog.LoggerOption) otellog.Logger                   { return lognoop.NewLoggerProvider().Logger("") }
func (s *telemetryStub) Meter(name string, _ ...otelmetric.MeterOption) otelmetric.Meter             { return s.meterProvider.Meter(name) }
func (s *telemetryStub) MeterProvider() otelmetric.MeterProvider                                     { return s.meterProvider }
func (s *telemetryStub) Propagator() propagation.TextMapPropagator                                   { return propagation.NewCompositeTextMapPropagator() }
func (s *telemetryStub) Tracer(name string, _ ...oteltrace.TracerOption) oteltrace.Tracer            { return s.traceProvider.Tracer(name) }
func (s *telemetryStub) TracerProvider() oteltrace.TracerProvider                                    { return s.traceProvider }

func setupTelemetry(t *testing.T, enabled bool) (*recordingSpanExporter, *mocksconfig.Config, contractstelemetry.Resolver) {
	t.Helper()

	exporter := &recordingSpanExporter{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	exporter.tracer = provider.Tracer(instrumentationName)
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetBool(enabledConfigKey, true).Return(enabled).Maybe()

	tel := &telemetryStub{traceProvider: provider, meterProvider: metricnoop.NewMeterProvider()}
	resolver := func() contractstelemetry.Telemetry { return tel }

	return exporter, mockConfig, resolver
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
			return attr.Value.String(), true
		}
	}
	return "", false
}

func TestNewInstrument(t *testing.T) {
	t.Run("inactive when config is nil", func(t *testing.T) {
		inst := NewInstrument(testPool(), "postgres", nil, func() contractstelemetry.Telemetry { return nil })
		assert.False(t, inst.active())
	})

	t.Run("inactive when resolver is nil", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetBool(enabledConfigKey, true).Return(true).Maybe()
		inst := NewInstrument(testPool(), "postgres", mockConfig, nil)
		assert.False(t, inst.active())
	})

	t.Run("inactive when disabled", func(t *testing.T) {
		_, mockConfig, resolver := setupTelemetry(t, false)
		inst := NewInstrument(testPool(), "postgres", mockConfig, resolver)
		assert.False(t, inst.active())
	})

	t.Run("captures base attributes", func(t *testing.T) {
		inst := NewInstrument(testPool(), "postgres", nil, nil)
		assert.Equal(t, []telemetry.KeyValue{
			semconv.DBSystemNamePostgreSQL,
			semconv.DBNamespace("app"),
			semconv.ServerAddress("db.local"),
			semconv.ServerPort(5432),
			semconv.DBClientConnectionPoolName("postgres"),
		}, inst.baseAttrs)
	})

	t.Run("skips empty connection details", func(t *testing.T) {
		inst := NewInstrument(contractsdatabase.Pool{Writers: []contractsdatabase.Config{{Driver: "sqlite"}}}, "", nil, nil)
		assert.Equal(t, []telemetry.KeyValue{semconv.DBSystemNameSQLite}, inst.baseAttrs)
	})

	t.Run("retries resolution when resolver returns nil then succeeds", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetBool(enabledConfigKey, true).Return(true).Maybe()

		var tel contractstelemetry.Telemetry
		resolver := func() contractstelemetry.Telemetry { return tel }

		inst := NewInstrument(testPool(), "postgres", mockConfig, resolver)
		assert.False(t, inst.active())

		exporter := &recordingSpanExporter{}
		provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
		t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
		tel = &telemetryStub{traceProvider: provider, meterProvider: metricnoop.NewMeterProvider()}

		assert.True(t, inst.active())
	})
}

func TestInstrument_EndSpan(t *testing.T) {
	t.Run("records query attributes and renames span", func(t *testing.T) {
		exporter, mockConfig, resolver := setupTelemetry(t, true)
		inst := NewInstrument(testPool(), "postgres", mockConfig, resolver)
		assert.True(t, inst.active())

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
		exporter, mockConfig, resolver := setupTelemetry(t, true)
		inst := NewInstrument(testPool(), "postgres", mockConfig, resolver)
		assert.True(t, inst.active())

		ctx, span := inst.startSpan(context.Background(), "gorm.Query")
		inst.endSpan(ctx, span, time.Now(), "SELECT * FROM users", "users", -1, assert.AnError)

		assert.Equal(t, codes.Error, exporter.spans[0].Status().Code)
	})

	t.Run("ignores record not found", func(t *testing.T) {
		exporter, mockConfig, resolver := setupTelemetry(t, true)
		inst := NewInstrument(testPool(), "postgres", mockConfig, resolver)
		assert.True(t, inst.active())

		ctx, span := inst.startSpan(context.Background(), "gorm.Query")
		inst.endSpan(ctx, span, time.Now(), "SELECT * FROM users", "users", -1, gorm.ErrRecordNotFound)

		assert.Equal(t, codes.Unset, exporter.spans[0].Status().Code)
	})

	t.Run("raw query without table omits collection name", func(t *testing.T) {
		exporter, mockConfig, resolver := setupTelemetry(t, true)
		inst := NewInstrument(testPool(), "postgres", mockConfig, resolver)
		assert.True(t, inst.active())

		ctx, span := inst.startSpan(context.Background(), "SELECT")
		inst.endSpan(ctx, span, time.Now(), "SELECT 1", "", -1, nil)

		assert.Len(t, exporter.spans, 1)
		recorded := exporter.spans[0]
		assert.Equal(t, "SELECT", recorded.Name())
		_, ok := attrValue(recorded, "db.collection.name")
		assert.False(t, ok)
		_, ok = attrValue(recorded, "db.query.summary")
		assert.False(t, ok)
	})

	t.Run("negative rows omits returned_rows attribute", func(t *testing.T) {
		exporter, mockConfig, resolver := setupTelemetry(t, true)
		inst := NewInstrument(testPool(), "postgres", mockConfig, resolver)
		assert.True(t, inst.active())

		ctx, span := inst.startSpan(context.Background(), "gorm.Query")
		inst.endSpan(ctx, span, time.Now(), "SELECT * FROM users", "users", -1, nil)

		_, ok := attrValue(exporter.spans[0], "db.response.returned_rows")
		assert.False(t, ok)
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
		name     string
		query    string
		expected string
	}{
		{"select", "SELECT * FROM users", "SELECT"},
		{"leading spaces", "  insert into users values (?)", "INSERT"},
		{"update", "UPDATE users SET name = ?", "UPDATE"},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"leading newline", "\n\tDELETE FROM users", "DELETE"},
		{"single word", "COMMIT", "COMMIT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
