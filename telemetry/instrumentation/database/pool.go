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
	metricConnectionTimeouts = "db.client.connection.timeouts"

	unitConnections = "{connection}"
	unitTimeout     = "{timeout}"
)

func (r *Instrument) registerPoolMetrics(db *sql.DB) error {
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

	waitTime, err := r.meter.Float64ObservableCounter(metricConnectionWaitTime,
		metric.WithUnit(unitSeconds),
		metric.WithDescription("Cumulative time blocked waiting for an available connection"),
	)
	if err != nil {
		return err
	}

	timeouts, err := r.meter.Int64ObservableCounter(metricConnectionTimeouts,
		metric.WithUnit(unitTimeout),
		metric.WithDescription("Cumulative count of connection pool timeouts (idle, lifetime, idle time)"),
	)
	if err != nil {
		return err
	}

	idle := metric.WithAttributes(append(slices.Clone(r.baseAttrs), semconv.DBClientConnectionStateIdle)...)
	used := metric.WithAttributes(append(slices.Clone(r.baseAttrs), semconv.DBClientConnectionStateUsed)...)
	base := metric.WithAttributes(r.baseAttrs...)

	_, err = r.meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		stats := db.Stats()
		observer.ObserveInt64(count, int64(stats.InUse), used)
		observer.ObserveInt64(count, int64(stats.Idle), idle)
		observer.ObserveInt64(maxConns, int64(stats.MaxOpenConnections), base)
		observer.ObserveFloat64(waitTime, stats.WaitDuration.Seconds(), base)
		observer.ObserveInt64(timeouts, stats.MaxIdleClosed+stats.MaxIdleTimeClosed+stats.MaxLifetimeClosed, base)
		return nil
	}, count, maxConns, waitTime, timeouts)

	return err
}
