package database

import (
	"context"
	"database/sql"
	"slices"

	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const (
	metricConnectionCount = "db.client.connection.count"
	metricConnectionMax   = "db.client.connection.max"

	unitConnections = "{connection}"
)

func (r *Instrument) observePool(db *sql.DB) error {
	count, err := r.meter.Int64ObservableUpDownCounter(metricConnectionCount,
		metric.WithUnit(unitConnections),
		metric.WithDescription("Open connections by state"),
	)
	if err != nil {
		return err
	}

	maxConns, err := r.meter.Int64ObservableUpDownCounter(metricConnectionMax,
		metric.WithUnit(unitConnections),
		metric.WithDescription("Maximum open connections allowed"),
	)
	if err != nil {
		return err
	}

	idle := metric.WithAttributes(append(slices.Clone(r.baseAttrs), semconv.DBClientConnectionStateIdle)...)
	used := metric.WithAttributes(append(slices.Clone(r.baseAttrs), semconv.DBClientConnectionStateUsed)...)
	base := metric.WithAttributes(r.baseAttrs...)

	registration, err := r.meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		stats := db.Stats()
		observer.ObserveInt64(count, int64(stats.InUse), used)
		observer.ObserveInt64(count, int64(stats.Idle), idle)
		observer.ObserveInt64(maxConns, int64(stats.MaxOpenConnections), base)
		return nil
	}, count, maxConns)
	if err != nil {
		return err
	}

	r.poolRegistration = registration

	return nil
}
