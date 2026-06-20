package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/suite"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type stubConnector struct{}

func (stubConnector) Connect(context.Context) (driver.Conn, error) { return nil, driver.ErrBadConn }
func (stubConnector) Driver() driver.Driver                        { return stubDriver{} }

type stubDriver struct{}

func (stubDriver) Open(string) (driver.Conn, error) { return nil, driver.ErrBadConn }

type PoolMetricsTestSuite struct {
	suite.Suite
	reader *sdkmetric.ManualReader
	inst   *Instrument
	db     *sql.DB
}

func TestPoolMetricsTestSuite(t *testing.T) {
	suite.Run(t, &PoolMetricsTestSuite{})
}

func (s *PoolMetricsTestSuite) SetupTest() {
	s.reader = sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(s.reader))
	s.T().Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	s.inst = &Instrument{
		baseAttrs: baseAttributes(testPool(), "postgres"),
		meter:     provider.Meter(instrumentationName),
	}

	s.db = sql.OpenDB(stubConnector{})
	s.T().Cleanup(func() { _ = s.db.Close() })
}

func (s *PoolMetricsTestSuite) collect() map[string]metricdata.Metrics {
	var rm metricdata.ResourceMetrics
	s.Require().NoError(s.reader.Collect(context.Background(), &rm))

	result := map[string]metricdata.Metrics{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			result[m.Name] = m
		}
	}
	return result
}

func (s *PoolMetricsTestSuite) TestRegistersAllMetrics() {
	s.Require().NoError(s.inst.registerPoolMetrics(s.db))

	metrics := s.collect()
	for _, name := range []string{metricConnectionCount, metricConnectionMax, metricConnectionWaitTime, metricConnectionTimeouts} {
		s.Contains(metrics, name)
	}
}

func (s *PoolMetricsTestSuite) TestConnectionCountHasIdleAndUsedStates() {
	s.Require().NoError(s.inst.registerPoolMetrics(s.db))

	metrics := s.collect()
	data := metrics[metricConnectionCount].Data.(metricdata.Sum[int64])
	s.Require().Len(data.DataPoints, 2)

	states := map[string]bool{}
	for _, dp := range data.DataPoints {
		val, ok := dp.Attributes.Value(semconv.DBClientConnectionStateKey)
		s.True(ok)
		states[val.AsString()] = true
	}
	s.Contains(states, "idle")
	s.Contains(states, "used")
}

func (s *PoolMetricsTestSuite) TestPoolNameAttribute() {
	s.Require().NoError(s.inst.registerPoolMetrics(s.db))

	metrics := s.collect()
	data := metrics[metricConnectionMax].Data.(metricdata.Sum[int64])
	s.Require().NotEmpty(data.DataPoints)

	poolName, ok := data.DataPoints[0].Attributes.Value(semconv.DBClientConnectionPoolNameKey)
	s.True(ok)
	s.Equal("postgres", poolName.AsString())
}
