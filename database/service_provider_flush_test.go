package database

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

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	databasedriver "github.com/goravel/framework/database/driver"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

var _ contractsfoundation.ServiceProviderWithFlushState = (*ServiceProvider)(nil)

func TestServiceProviderFlushState(t *testing.T) {
	t.Run("closes cached connections and builds a new one afterwards", func(t *testing.T) {
		t.Cleanup(databasedriver.CloseConnections)

		first, sqlDB := buildFlushGorm(t)
		(&ServiceProvider{}).FlushState(mocksfoundation.NewApplication(t))

		assert.ErrorContains(t, sqlDB.Ping(), "database is closed")

		second, _ := buildFlushGorm(t)
		assert.NotSame(t, first, second)
	})

	t.Run("is safe on empty state and when repeated", func(t *testing.T) {
		databasedriver.CloseConnections()
		provider := &ServiceProvider{}
		app := mocksfoundation.NewApplication(t)

		assert.NotPanics(t, func() {
			provider.FlushState(app)
			provider.FlushState(app)
		})
	})
}

func TestServiceProviderRegister_ClosesConnections(t *testing.T) {
	t.Cleanup(databasedriver.CloseConnections)

	_, sqlDB := buildFlushGorm(t)
	app := mocksfoundation.NewApplication(t)
	app.EXPECT().Singleton(mock.Anything, mock.Anything).Return().Times(4)

	(&ServiceProvider{}).Register(app)

	assert.ErrorContains(t, sqlDB.Ping(), "database is closed")
}

func buildFlushGorm(t *testing.T) (*gorm.DB, *sql.DB) {
	t.Helper()

	config := mocksconfig.NewConfig(t)
	config.EXPECT().GetBool("telemetry.instrumentation.database.enabled", true).Return(true).Maybe()
	config.EXPECT().GetInt("database.pool.max_idle_conns", 10).Return(10).Once()
	config.EXPECT().GetInt("database.pool.max_open_conns", 100).Return(100).Once()
	config.EXPECT().GetDuration("database.pool.conn_max_idletime", time.Duration(3600)).Return(time.Duration(3600)).Once()
	config.EXPECT().GetDuration("database.pool.conn_max_lifetime", time.Duration(3600)).Return(time.Duration(3600)).Once()

	pool := contractsdatabase.Pool{
		Writers: []contractsdatabase.Config{
			{Driver: "postgres", Database: "app", Dialector: flushDialector{}},
		},
	}
	instance, _, err := databasedriver.BuildGorm(config, gormlogger.Discard, pool, "flush", nil)
	assert.NoError(t, err)
	sqlDB, err := instance.DB()
	assert.NoError(t, err)

	return instance, sqlDB
}

type flushConnector struct{}

func (flushConnector) Connect(context.Context) (driver.Conn, error) { return nil, driver.ErrBadConn }
func (flushConnector) Driver() driver.Driver                        { return flushDriver{} }

type flushDriver struct{}

func (flushDriver) Open(string) (driver.Conn, error) { return nil, driver.ErrBadConn }

type flushDialector struct {
	gormtests.DummyDialector
}

func (flushDialector) Initialize(db *gorm.DB) error {
	if err := (gormtests.DummyDialector{}).Initialize(db); err != nil {
		return err
	}

	db.ConnPool = sql.OpenDB(flushConnector{})

	return nil
}
