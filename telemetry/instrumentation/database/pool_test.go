package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/telemetry"
)

type stubConnector struct{}

func (stubConnector) Connect(_ context.Context) (driver.Conn, error) { return nil, driver.ErrBadConn }
func (stubConnector) Driver() driver.Driver                          { return nil }

func openStubDB() *sql.DB {
	return sql.OpenDB(stubConnector{})
}

func TestRegisterPoolMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	mockTelemetry := mockstelemetry.NewTelemetry(t)
	mockTelemetry.EXPECT().Meter(instrumentationName).RunAndReturn(provider.Meter).Once()

	original := telemetry.Facade
	telemetry.Facade = mockTelemetry
	t.Cleanup(func() { telemetry.Facade = original })

	db := openStubDB()
	defer db.Close()

	assert.NoError(t, RegisterPoolMetrics(db, "postgres"))

	var rm metricdata.ResourceMetrics
	assert.NoError(t, reader.Collect(context.Background(), &rm))

	names := map[string]bool{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			names[m.Name] = true
		}
	}
	assert.True(t, names["db.client.connection.count"])
	assert.True(t, names["db.client.connection.max"])
	assert.True(t, names["db.client.connection.wait_time"])
	assert.True(t, names["db.client.connection.waits"])
}

func TestRegisterPoolMetrics_NilFacade(t *testing.T) {
	original := telemetry.Facade
	telemetry.Facade = nil
	t.Cleanup(func() { telemetry.Facade = original })

	db := openStubDB()
	defer db.Close()

	assert.NoError(t, RegisterPoolMetrics(db, "postgres"))
}
