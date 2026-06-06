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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/goravel/framework/telemetry"
)

const (
	instrumentationName = "github.com/goravel/framework/telemetry/instrumentation/database"

	metricOperationDuration = "db.client.operation.duration"
	unitSeconds             = "s"
)

var durationBuckets = []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}

var (
	durationOnce sync.Once
	durationHist metric.Float64Histogram
)

func operationDuration() metric.Float64Histogram {
	durationOnce.Do(func() {
		if telemetry.Facade == nil {
			return
		}

		meter := telemetry.Facade.Meter(instrumentationName)
		durationHist, _ = meter.Float64Histogram(metricOperationDuration,
			metric.WithUnit(unitSeconds),
			metric.WithDescription("Duration of database client operations"),
			metric.WithExplicitBucketBoundaries(durationBuckets...),
		)
	})

	return durationHist
}

func dbSystem(driverName string) attribute.KeyValue {
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

func isRecordableError(err error) bool {
	if err == nil {
		return false
	}

	for _, ignored := range []error{gorm.ErrRecordNotFound, sql.ErrNoRows, driver.ErrSkip, io.EOF} {
		if errors.Is(err, ignored) {
			return false
		}
	}

	return true
}

func startSpan(ctx context.Context, name string) (context.Context, oteltrace.Span, bool) {
	if telemetry.Facade == nil {
		return ctx, nil, false
	}

	spanCtx, span := telemetry.Facade.Tracer(instrumentationName).Start(ctx, name, oteltrace.WithSpanKind(oteltrace.SpanKindClient))

	return spanCtx, span, true
}

func endSpan(ctx context.Context, span oteltrace.Span, start time.Time, system attribute.KeyValue, query, table string, rows int64, err error) {
	operation := operationName(query)

	attrs := []attribute.KeyValue{
		system,
		semconv.DBOperationName(operation),
		semconv.DBQueryText(query),
	}
	if table != "" {
		summary := operation + " " + table
		attrs = append(attrs, semconv.DBCollectionName(table), semconv.DBQuerySummary(summary))
		span.SetName(summary)
	}
	if rows >= 0 {
		attrs = append(attrs, semconv.DBResponseReturnedRows(int(rows)))
	}

	metricAttrs := []attribute.KeyValue{system, semconv.DBOperationName(operation)}
	if table != "" {
		metricAttrs = append(metricAttrs, semconv.DBCollectionName(table))
	}

	if isRecordableError(err) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		metricAttrs = append(metricAttrs, semconv.ErrorType(err))
	}

	span.SetAttributes(attrs...)
	span.End()

	if hist := operationDuration(); hist != nil {
		hist.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(metricAttrs...))
	}
}
