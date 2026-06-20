package driver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	gormtests "gorm.io/gorm/utils/tests"

	"github.com/goravel/framework/contracts/database"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	mocksconfig "github.com/goravel/framework/mocks/config"
	instrumentationdatabase "github.com/goravel/framework/telemetry/instrumentation/database"
)

func TestBuildGorm_TelemetryPlugin(t *testing.T) {
	t.Cleanup(ResetConnections)

	resolver := func() contractstelemetry.Telemetry { return nil }
	instance, err := BuildGorm(stubGormConfig(t), gormlogger.Discard, stubPool(), "primary", resolver)
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	t.Cleanup(func() {
		if db, err := instance.DB(); err == nil {
			_ = db.Close()
		}
	})

	_, registered := instance.Plugins[instrumentationdatabase.PluginName]
	assert.True(t, registered)
}

type stubConnector struct{}

func (stubConnector) Connect(context.Context) (driver.Conn, error) { return nil, driver.ErrBadConn }
func (stubConnector) Driver() driver.Driver                        { return stubDriver{} }

type stubDriver struct{}

func (stubDriver) Open(string) (driver.Conn, error) { return nil, driver.ErrBadConn }

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
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetInt(mock.Anything, mock.Anything).Return(10).Maybe()
	mockConfig.EXPECT().GetDuration(mock.Anything, mock.Anything).Return(time.Duration(3600)).Maybe()

	return mockConfig
}
