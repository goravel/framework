package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
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
	mu     sync.Mutex
	spans  []sdktrace.ReadOnlySpan
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

func (s *telemetryStub) ForceFlush(_ context.Context) error { return nil }
func (s *telemetryStub) Shutdown(_ context.Context) error   { return nil }
func (s *telemetryStub) Logger(_ string, _ ...otellog.LoggerOption) otellog.Logger {
	return lognoop.NewLoggerProvider().Logger("")
}
func (s *telemetryStub) Meter(name string, _ ...otelmetric.MeterOption) otelmetric.Meter {
	return s.meterProvider.Meter(name)
}
func (s *telemetryStub) MeterProvider() otelmetric.MeterProvider { return s.meterProvider }
func (s *telemetryStub) Propagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator()
}
func (s *telemetryStub) Tracer(name string, _ ...oteltrace.TracerOption) oteltrace.Tracer {
	return s.traceProvider.Tracer(name)
}
func (s *telemetryStub) TracerProvider() oteltrace.TracerProvider { return s.traceProvider }

func setupTelemetry(t *testing.T) (*recordingSpanExporter, contractstelemetry.Resolver) {
	t.Helper()

	exporter := &recordingSpanExporter{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	exporter.tracer = provider.Tracer(instrumentationName)
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	tel := &telemetryStub{traceProvider: provider, meterProvider: metricnoop.NewMeterProvider()}
	resolver := func() contractstelemetry.Telemetry { return tel }

	return exporter, resolver
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

// InstrumentTestSuite covers Enabled, NewInstrument, endSpan, dbSystem, operationName, and isRecordableError.
type InstrumentTestSuite struct {
	suite.Suite
	exporter   *recordingSpanExporter
	instrument *Instrument
}

func TestInstrumentTestSuite(t *testing.T) {
	suite.Run(t, &InstrumentTestSuite{})
}

func (s *InstrumentTestSuite) SetupTest() {
	exporter, resolver := setupTelemetry(s.T())
	s.exporter = exporter
	s.instrument = NewInstrument(testPool(), "postgres", resolver)
}

func (s *InstrumentTestSuite) TestEnabled() {
	s.Run("nil config", func() {
		s.False(Enabled(nil))
	})

	s.Run("enabled by default", func() {
		m := mocksconfig.NewConfig(s.T())
		m.EXPECT().GetBool(enabledConfigKey, true).Return(true)
		s.True(Enabled(m))
	})

	s.Run("explicitly disabled", func() {
		m := mocksconfig.NewConfig(s.T())
		m.EXPECT().GetBool(enabledConfigKey, true).Return(false)
		s.False(Enabled(m))
	})
}

func (s *InstrumentTestSuite) TestActive() {
	tests := []struct {
		name     string
		inst     *Instrument
		expected bool
	}{
		{"nil receiver", nil, false},
		{"nil resolver", NewInstrument(testPool(), "postgres", nil), false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.inst.active())
		})
	}
}

func (s *InstrumentTestSuite) TestActive_RetriesResolution() {
	var tel contractstelemetry.Telemetry
	inst := NewInstrument(testPool(), "postgres", func() contractstelemetry.Telemetry { return tel })
	s.False(inst.active())

	exporter := &recordingSpanExporter{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	s.T().Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	tel = &telemetryStub{traceProvider: provider, meterProvider: metricnoop.NewMeterProvider()}

	s.True(inst.active())
}

func (s *InstrumentTestSuite) TestBaseAttributes() {
	tests := []struct {
		name     string
		pool     contractsdatabase.Pool
		conn     string
		expected []telemetry.KeyValue
	}{
		{
			name: "all fields",
			pool: testPool(),
			conn: "postgres",
			expected: []telemetry.KeyValue{
				semconv.DBSystemNamePostgreSQL,
				semconv.DBNamespace("app"),
				semconv.ServerAddress("db.local"),
				semconv.ServerPort(5432),
				semconv.DBClientConnectionPoolName("postgres"),
			},
		},
		{
			name:     "minimal",
			pool:     contractsdatabase.Pool{Writers: []contractsdatabase.Config{{Driver: "sqlite"}}},
			conn:     "",
			expected: []telemetry.KeyValue{semconv.DBSystemNameSQLite},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			inst := NewInstrument(tt.pool, tt.conn, nil)
			s.Equal(tt.expected, inst.baseAttrs)
		})
	}
}

func (s *InstrumentTestSuite) TestEndSpan() {
	tests := []struct {
		name           string
		query          string
		table          string
		rows           int64
		err            error
		expectedName   string
		expectedStatus codes.Code
		hasCollection  bool
		hasRows        bool
	}{
		{
			name:           "query with table",
			query:          "SELECT * FROM users WHERE id = ?",
			table:          "users",
			rows:           3,
			expectedName:   "SELECT users",
			expectedStatus: codes.Unset,
			hasCollection:  true,
			hasRows:        true,
		},
		{
			name:           "error status",
			query:          "SELECT * FROM users",
			table:          "users",
			rows:           -1,
			err:            io.ErrUnexpectedEOF,
			expectedName:   "SELECT users",
			expectedStatus: codes.Error,
			hasCollection:  true,
		},
		{
			name:           "record not found is not an error",
			query:          "SELECT * FROM users",
			table:          "users",
			rows:           -1,
			err:            gorm.ErrRecordNotFound,
			expectedName:   "SELECT users",
			expectedStatus: codes.Unset,
			hasCollection:  true,
		},
		{
			name:           "raw query without table",
			query:          "SELECT 1",
			table:          "",
			rows:           -1,
			expectedName:   "SELECT",
			expectedStatus: codes.Unset,
		},
		{
			name:           "negative rows omits attribute",
			query:          "SELECT * FROM users",
			table:          "users",
			rows:           -1,
			expectedName:   "SELECT users",
			expectedStatus: codes.Unset,
			hasCollection:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			exporter, resolver := setupTelemetry(s.T())
			inst := NewInstrument(testPool(), "postgres", resolver)
			s.True(inst.active())

			ctx, span := inst.startSpan(context.Background(), "db")
			inst.endSpan(ctx, span, time.Now(), tt.query, tt.table, tt.rows, tt.err, "")

			s.Require().Len(exporter.spans, 1)
			recorded := exporter.spans[0]
			s.Equal(tt.expectedName, recorded.Name())
			s.Equal(tt.expectedStatus, recorded.Status().Code)

			_, ok := attrValue(recorded, "db.collection.name")
			s.Equal(tt.hasCollection, ok)
			_, ok = attrValue(recorded, "db.response.returned_rows")
			s.Equal(tt.hasRows, ok)
		})
	}
}

func (s *InstrumentTestSuite) TestEndSpan_ResolverMode() {
	exporter, resolver := setupTelemetry(s.T())
	inst := NewInstrument(testPool(), "postgres", resolver)
	s.True(inst.active())

	ctx, span := inst.startSpan(context.Background(), "db")
	inst.endSpan(ctx, span, time.Now(), "SELECT * FROM users", "users", 1, nil, "replica")

	s.Require().Len(exporter.spans, 1)
	val, ok := attrValue(exporter.spans[0], attrResolverMode)
	s.True(ok)
	s.Equal("replica", val)
}

func (s *InstrumentTestSuite) TestEndSpan_EmptyResolverModeOmitsAttribute() {
	exporter, resolver := setupTelemetry(s.T())
	inst := NewInstrument(testPool(), "postgres", resolver)
	s.True(inst.active())

	ctx, span := inst.startSpan(context.Background(), "db")
	inst.endSpan(ctx, span, time.Now(), "SELECT * FROM users", "users", 1, nil, "")

	s.Require().Len(exporter.spans, 1)
	_, ok := attrValue(exporter.spans[0], attrResolverMode)
	s.False(ok)
}

func (s *InstrumentTestSuite) TestDBSystem() {
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
		s.Run(tt.driver, func() {
			kv := dbSystem(tt.driver)
			s.Equal(semconv.DBSystemNameKey, kv.Key)
			s.Equal(tt.expected, kv.Value.AsString())
		})
	}
}

func (s *InstrumentTestSuite) TestOperationName() {
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
		s.Run(tt.name, func() {
			s.Equal(tt.expected, operationName(tt.query))
		})
	}
}

func (s *InstrumentTestSuite) TestIsRecordableError() {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil", nil, false},
		{"record not found", gorm.ErrRecordNotFound, false},
		{"no rows", sql.ErrNoRows, false},
		{"driver skip", driver.ErrSkip, false},
		{"eof", io.EOF, false},
		{"real error", io.ErrUnexpectedEOF, true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, isRecordableError(tt.err))
		})
	}
}
