package database

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/goravel/framework/telemetry"
)

const (
	metricConnectionCount    = "db.client.connection.count"
	metricConnectionMax      = "db.client.connection.max"
	metricConnectionWaitTime = "db.client.connection.wait_time"
	metricConnectionWaits    = "db.client.connection.waits"

	stateKey    = attribute.Key("db.client.connection.state")
	poolNameKey = attribute.Key("db.client.connection.pool.name")
)

// RegisterPoolMetrics exports sql.DBStats as observable metrics. Call once per *sql.DB.
func RegisterPoolMetrics(db *sql.DB, driverName, poolName string) error {
	if telemetry.Facade == nil {
		return nil
	}

	meter := telemetry.Facade.Meter(instrumentationName)
	system := dbSystem(driverName)
	pool := poolNameKey.String(poolName)

	count, err := meter.Int64ObservableUpDownCounter(metricConnectionCount, metric.WithUnit("{connection}"), metric.WithDescription("Open connections by state"))
	if err != nil {
		return err
	}
	maxConns, err := meter.Int64ObservableUpDownCounter(metricConnectionMax, metric.WithUnit("{connection}"), metric.WithDescription("Maximum open connections allowed"))
	if err != nil {
		return err
	}
	waitTime, err := meter.Float64ObservableCounter(metricConnectionWaitTime, metric.WithUnit(unitSeconds), metric.WithDescription("Cumulative time waiting for a connection"))
	if err != nil {
		return err
	}
	waits, err := meter.Int64ObservableCounter(metricConnectionWaits, metric.WithUnit("{wait}"), metric.WithDescription("Cumulative count of connection waits"))
	if err != nil {
		return err
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		stats := db.Stats()
		observer.ObserveInt64(count, int64(stats.InUse), metric.WithAttributes(system, pool, stateKey.String("used")))
		observer.ObserveInt64(count, int64(stats.Idle), metric.WithAttributes(system, pool, stateKey.String("idle")))
		observer.ObserveInt64(maxConns, int64(stats.MaxOpenConnections), metric.WithAttributes(system, pool))
		observer.ObserveFloat64(waitTime, stats.WaitDuration.Seconds(), metric.WithAttributes(system, pool))
		observer.ObserveInt64(waits, stats.WaitCount, metric.WithAttributes(system, pool))
		return nil
	}, count, maxConns, waitTime, waits)

	return err
}
