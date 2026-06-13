package database

import (
	"context"
	"database/sql"
	"slices"

	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const (
	metricConnectionCount    = "db.client.connection.count"
	metricConnectionMax      = "db.client.connection.max"
	metricConnectionWaitTime = "db.client.connection.wait_time"
	metricConnectionWaits    = "db.client.connection.waits"

	unitConnections = "{connection}"
	unitWaits       = "{wait}"
)

// registerPoolMetrics exports sql.DBStats as observable metrics. Call once per *sql.DB.
func (r *instrument) registerPoolMetrics(db *sql.DB) error {
	count, err := r.meter.Int64ObservableUpDownCounter(metricConnectionCount, metric.WithUnit(unitConnections), metric.WithDescription("Open connections by state"))
	if err != nil {
		return err
	}
	maxConns, err := r.meter.Int64ObservableUpDownCounter(metricConnectionMax, metric.WithUnit(unitConnections), metric.WithDescription("Maximum open connections allowed"))
	if err != nil {
		return err
	}
	waitTime, err := r.meter.Float64ObservableCounter(metricConnectionWaitTime, metric.WithUnit(unitSeconds), metric.WithDescription("Cumulative time waiting for a connection"))
	if err != nil {
		return err
	}
	waits, err := r.meter.Int64ObservableCounter(metricConnectionWaits, metric.WithUnit(unitWaits), metric.WithDescription("Cumulative count of connection waits"))
	if err != nil {
		return err
	}

	base := metric.WithAttributes(r.baseAttrs...)
	idle := metric.WithAttributes(append(slices.Clone(r.baseAttrs), semconv.DBClientConnectionStateIdle)...)
	used := metric.WithAttributes(append(slices.Clone(r.baseAttrs), semconv.DBClientConnectionStateUsed)...)

	_, err = r.meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		stats := db.Stats()
		observer.ObserveInt64(count, int64(stats.InUse), used)
		observer.ObserveInt64(count, int64(stats.Idle), idle)
		observer.ObserveInt64(maxConns, int64(stats.MaxOpenConnections), base)
		observer.ObserveFloat64(waitTime, stats.WaitDuration.Seconds(), base)
		observer.ObserveInt64(waits, stats.WaitCount, base)
		return nil
	}, count, maxConns, waitTime, waits)

	return err
}
