package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/goravel/framework/contracts/config"
	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
)

const (
	instrumentationName = "github.com/goravel/framework/telemetry/instrumentation/database"

	enabledConfigKey = "telemetry.instrumentation.database.enabled"

	metricOperationDuration = "db.client.operation.duration"
	unitSeconds             = "s"
)

var durationBuckets = []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}

func Enabled(config config.Config) bool {
	return config != nil && config.GetBool(enabledConfigKey, true)
}

type Instrument struct {
	baseAttrs []telemetry.KeyValue
	resolver  contractstelemetry.Resolver

	mu           sync.Mutex
	tracer       trace.Tracer
	meter        metric.Meter
	durationHist metric.Float64Histogram

	sqlDB        *sql.DB
	poolObserved bool
}

// SetDB stores the primary writer's *sql.DB for pool metrics.
// Pool metrics cover the writer pool only; dbresolver replica pools are
// internal gorm.ConnPool instances with no public sql.DBStats access.
func (r *Instrument) SetDB(db *sql.DB) {
	if r != nil {
		r.sqlDB = db
	}
}

func NewInstrument(pool contractsdatabase.Pool, connection string, resolver contractstelemetry.Resolver) *Instrument {
	return &Instrument{
		baseAttrs: baseAttributes(pool, connection),
		resolver:  resolver,
	}
}

func (r *Instrument) active() bool {
	if r == nil || r.resolver == nil {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tracer != nil {
		r.startPoolObservation()
		return true
	}

	tel := r.resolver()
	if tel == nil {
		return false
	}

	r.tracer = tel.Tracer(instrumentationName)
	r.meter = tel.Meter(instrumentationName)
	r.durationHist, _ = r.meter.Float64Histogram(metricOperationDuration,
		metric.WithUnit(unitSeconds),
		metric.WithDescription("Duration of database client operations"),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
	)

	r.startPoolObservation()

	return true
}

func (r *Instrument) startPoolObservation() {
	if r.poolObserved || r.sqlDB == nil {
		return
	}
	r.poolObserved = true

	if err := r.observePool(r.sqlDB); err != nil {
		color.Warningln(err.Error())
	}
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
	return r.tracer.Start(ctx, name, telemetry.WithSpanKind(telemetry.SpanKindClient))
}

func (r *Instrument) endSpan(ctx context.Context, span trace.Span, start time.Time, query, table string, rows int64, err error) {
	operation := operationName(query)

	attrs := append([]telemetry.KeyValue{}, r.baseAttrs...)
	attrs = append(attrs, semconv.DBOperationName(operation), semconv.DBQueryText(query))
	if table != "" {
		name := operation + " " + table
		attrs = append(attrs, semconv.DBCollectionName(table), semconv.DBQuerySummary(name))
		span.SetName(name)
	} else {
		span.SetName(operation)
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
	query = strings.TrimLeft(query, " \t\n\r")
	if query == "" {
		return ""
	}
	if i := strings.IndexByte(query, ' '); i > 0 {
		return strings.ToUpper(query[:i])
	}
	return strings.ToUpper(query)
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
