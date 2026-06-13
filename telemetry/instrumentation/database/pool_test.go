package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type stubConnector struct{}

func (stubConnector) Connect(context.Context) (driver.Conn, error) { return nil, driver.ErrBadConn }
func (stubConnector) Driver() driver.Driver                        { return nil }

func TestInstrument_RegisterPoolMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	// Pool metrics need only the meter and base attributes, so build a partial
	// instrument rather than going through newInstrument and its facade gating.
	inst := &instrument{
		meter:     provider.Meter(instrumentationName),
		baseAttrs: baseAttributes(testPool(), "postgres"),
	}

	db := sql.OpenDB(stubConnector{})
	defer func() { _ = db.Close() }()

	assert.NoError(t, inst.registerPoolMetrics(db))

	var rm metricdata.ResourceMetrics
	assert.NoError(t, reader.Collect(context.Background(), &rm))

	names := map[string]bool{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			names[m.Name] = true
			if m.Name == metricConnectionMax {
				data := m.Data.(metricdata.Sum[int64])
				assert.NotEmpty(t, data.DataPoints)
				poolName, ok := data.DataPoints[0].Attributes.Value(semconv.DBClientConnectionPoolNameKey)
				assert.True(t, ok)
				assert.Equal(t, "postgres", poolName.AsString())
			}
			if m.Name == metricConnectionCount {
				data := m.Data.(metricdata.Sum[int64])
				assert.Len(t, data.DataPoints, 2)
			}
		}
	}
	for _, name := range []string{metricConnectionCount, metricConnectionMax, metricConnectionWaitTime, metricConnectionWaits} {
		assert.True(t, names[name], name)
	}
}
