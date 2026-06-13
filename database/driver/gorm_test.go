package driver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	gormtests "gorm.io/gorm/utils/tests"

	"github.com/goravel/framework/contracts/database"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/telemetry"
	instrumentationdatabase "github.com/goravel/framework/telemetry/instrumentation/database"
)

func TestBuildGorm_TelemetryPlugin(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T)
		registered bool
	}{
		{
			name:       "registered when telemetry enabled",
			setup:      func(t *testing.T) { setupTelemetry(t, true) },
			registered: true,
		},
		{
			name:  "skipped when disabled",
			setup: func(t *testing.T) { setupTelemetry(t, false) },
		},
		{
			name: "skipped when facade is not set",
			setup: func(t *testing.T) {
				original := telemetry.Facade
				telemetry.Facade = nil
				t.Cleanup(func() { telemetry.Facade = original })
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(ResetConnections)
			tt.setup(t)

			instance, err := BuildGorm(stubGormConfig(t), gormlogger.Discard, stubPool(), tt.name)
			assert.NoError(t, err)
			assert.NotNil(t, instance)
			t.Cleanup(func() {
				if db, err := instance.DB(); err == nil {
					_ = db.Close()
				}
			})

			_, registered := instance.Plugins[instrumentationdatabase.PluginName]
			assert.Equal(t, tt.registered, registered)
		})
	}
}

type stubConnector struct{}

func (stubConnector) Connect(context.Context) (driver.Conn, error) { return nil, driver.ErrBadConn }
func (stubConnector) Driver() driver.Driver                        { return nil }

type stubDialector struct {
	gormtests.DummyDialector
}

func (stubDialector) Initialize(db *gorm.DB) error {
	if err := (gormtests.DummyDialector{}).Initialize(db); err != nil {
		return err
	}

	db.ConnPool = sql.OpenDB(stubConnector{})

	return nil
}

func stubPool() database.Pool {
	return database.Pool{
		Writers: []database.Config{
			{Driver: "postgres", Database: "app", Dialector: stubDialector{}},
		},
	}
}

func stubGormConfig(t *testing.T) *mocksconfig.Config {
	// Pool sizing is incidental to plugin registration; accept whatever reads
	// BuildGorm makes without pinning their exact count.
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetInt(mock.Anything, mock.Anything).Return(10).Maybe()
	mockConfig.EXPECT().GetDuration(mock.Anything, mock.Anything).Return(time.Duration(3600)).Maybe()

	return mockConfig
}

func setupTelemetry(t *testing.T, enabled bool) {
	t.Helper()

	mockTelemetry := mockstelemetry.NewTelemetry(t)
	mockTelemetry.EXPECT().Tracer(mock.Anything).Return(tracenoop.NewTracerProvider().Tracer("test")).Maybe()
	mockTelemetry.EXPECT().Meter(mock.Anything).Return(metricnoop.NewMeterProvider().Meter("test")).Maybe()

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetBool("telemetry.instrumentation.database.enabled", true).Return(enabled).Maybe()

	originalFacade, originalConfig := telemetry.Facade, telemetry.ConfigFacade
	telemetry.Facade, telemetry.ConfigFacade = mockTelemetry, mockConfig
	t.Cleanup(func() { telemetry.Facade, telemetry.ConfigFacade = originalFacade, originalConfig })
}
