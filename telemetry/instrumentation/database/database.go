package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/telemetry"
)

const (
	instrumentationName = "github.com/goravel/framework/telemetry/instrumentation/database"

	enabledConfigKey = "telemetry.instrumentation.database.enabled"

	metricOperationDuration = "db.client.operation.duration"
	unitSeconds             = "s"
)

var durationBuckets = []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}

// Instrument builds the spans and metrics shared by the gorm plugin and the
// query-builder decorator. It resolves the telemetry facade lazily on first use
// via resolver, so a connection built before telemetry is ready still ends up
// instrumented once telemetry becomes available.
type Instrument struct {
	resolver  contractstelemetry.Resolver
	baseAttrs []telemetry.KeyValue

	mu           sync.Mutex
	ready        atomic.Bool
	disabled     atomic.Bool
	tracer       trace.Tracer
	meter        metric.Meter
	durationHist metric.Float64Histogram
}

// NewInstrument returns the shared instrumentation core. It never returns nil:
// resolver is called lazily (see active), so callers always wrap and the wrapper
// no-ops until telemetry is available and enabled.
func NewInstrument(resolver contractstelemetry.Resolver, pool contractsdatabase.Pool, connection string) *Instrument {
	return &Instrument{
		resolver:  resolver,
		baseAttrs: baseAttributes(pool, connection),
	}
}

// FacadeResolver resolves the process-wide telemetry facade. It is the default
// resolver the database layer passes to the instrumentation.
func FacadeResolver() contractstelemetry.Telemetry {
	return telemetry.Facade
}

// active reports whether instrumentation is on, resolving telemetry and building
// the tracer and metric instruments once on the first call that finds it ready.
func (r *Instrument) active() bool {
	if r.ready.Load() {
		return true
	}
	if r.disabled.Load() {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ready.Load() {
		return true
	}
	if r.disabled.Load() {
		return false
	}

	if r.resolver == nil || telemetry.ConfigFacade == nil || !telemetry.ConfigFacade.GetBool(enabledConfigKey, true) {
		r.disabled.Store(true)
		return false
	}

	tel := r.resolver()
	if tel == nil {
		// Telemetry is not configured for this application; give up permanently so
		// instrumented queries pay only a couple of atomic loads from here on.
		r.disabled.Store(true)
		return false
	}

	meter := tel.Meter(instrumentationName)
	durationHist, _ := meter.Float64Histogram(metricOperationDuration,
		metric.WithUnit(unitSeconds),
		metric.WithDescription("Duration of database client operations"),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
	)

	r.tracer = tel.Tracer(instrumentationName)
	r.meter = meter
	r.durationHist = durationHist
	r.ready.Store(true)

	return true
}

func baseAttributes(pool contractsdatabase.Pool, connection string) []telemetry.KeyValue {
	if len(pool.Writers) == 0 {
		return nil
	}

	writer := pool.Writers[0]
	attrs := []telemetry.KeyValue{dbSystem(writer.Driver)}
	if writer.Database != "" {
		attrs = append(attrs, semconv.DBNamespace(writer.Database))
	}
	if writer.Host != "" {
		attrs = append(attrs, semconv.ServerAddress(writer.Host))
	}
	if writer.Port > 0 {
		attrs = append(attrs, semconv.ServerPort(writer.Port))
	}
	if connection != "" {
		attrs = append(attrs, semconv.DBClientConnectionPoolName(connection))
	}

	return attrs
}

func (r *Instrument) startSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	if !r.active() {
		return ctx, nil
	}

	return r.tracer.Start(ctx, name, telemetry.WithSpanKind(telemetry.SpanKindClient))
}

func (r *Instrument) endSpan(ctx context.Context, span trace.Span, start time.Time, query, table string, rows int64, err error) {
	if span == nil {
		return
	}

	operation := operationName(query)

	attrs := append([]telemetry.KeyValue{}, r.baseAttrs...)
	attrs = append(attrs, semconv.DBOperationName(operation), semconv.DBQueryText(query))
	if table != "" {
		summary := operation + " " + table
		attrs = append(attrs, semconv.DBCollectionName(table), semconv.DBQuerySummary(summary))
		span.SetName(summary)
	}
	if rows >= 0 {
		attrs = append(attrs, semconv.DBResponseReturnedRows(int(rows)))
	}

	metricAttrs := append([]telemetry.KeyValue{}, r.baseAttrs...)
	metricAttrs = append(metricAttrs, semconv.DBOperationName(operation))
	if table != "" {
		metricAttrs = append(metricAttrs, semconv.DBCollectionName(table))
	}

	if isRecordableError(err) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metricAttrs = append(metricAttrs, semconv.ErrorType(err))
	}

	span.SetAttributes(attrs...)

	// Record the metric while the span is still active so the SDK can attach an
	// exemplar correlating it to this span, then end the span.
	r.durationHist.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(metricAttrs...))

	span.End()
}

func dbSystem(driverName string) telemetry.KeyValue {
	switch driverName {
	case "postgres", "postgresql":
		return semconv.DBSystemNamePostgreSQL
	case "mysql":
		return semconv.DBSystemNameMySQL
	case "sqlite":
		return semconv.DBSystemNameSQLite
	case "sqlserver":
		return semconv.DBSystemNameMicrosoftSQLServer
	default:
		return semconv.DBSystemNameKey.String(driverName)
	}
}

func operationName(query string) string {
	fields := strings.Fields(query)
	if len(fields) == 0 {
		return ""
	}

	return strings.ToUpper(fields[0])
}

var ignoredErrors = []error{gorm.ErrRecordNotFound, sql.ErrNoRows, driver.ErrSkip, io.EOF}

func isRecordableError(err error) bool {
	if err == nil {
		return false
	}

	for _, ignored := range ignoredErrors {
		if errors.Is(err, ignored) {
			return false
		}
	}

	return true
}
