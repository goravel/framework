package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database"
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksdriver "github.com/goravel/framework/mocks/database/driver"
	mockslog "github.com/goravel/framework/mocks/log"
)

func TestBuildDB(t *testing.T) {
	var (
		mockConfig *mocksconfig.Config
		mockDriver *mocksdriver.Driver
	)

	tests := []struct {
		name          string
		connection    string
		setup         func()
		expectedError error
	}{
		{
			name:       "Success",
			connection: "mysql",
			setup: func() {
				driverCallback := func() (contractsdriver.Driver, error) {
					return mockDriver, nil
				}
				mockConfig.EXPECT().Get("database.connections.mysql.via").Return(driverCallback).Once()
				mockDriver.EXPECT().DB().Return(&sql.DB{}, nil).Once()
				mockDriver.EXPECT().Config().Return(database.Config{Driver: "mysql"}).Once()
			},
			expectedError: nil,
		},
		{
			name:       "Config Not Found",
			connection: "invalid",
			setup: func() {
				mockConfig.EXPECT().Get("database.connections.invalid.via").Return(nil).Once()
			},
			expectedError: errors.DatabaseConfigNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = mocksconfig.NewConfig(t)
			mockDriver = mocksdriver.NewDriver(t)
			test.setup()

			db, err := BuildDB(context.Background(), mockConfig, nil, test.connection)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}
}

func TestConnection(t *testing.T) {
	var (
		mockConfig *mocksconfig.Config
		mockDriver *mocksdriver.Driver
		mockLog    *mockslog.Log
	)

	tests := []struct {
		name          string
		connection    string
		setup         func(*DB)
		expectedPanic bool
	}{
		{
			name:       "Success with empty connection name",
			connection: "",
			setup: func(db *DB) {
				mockConfig.EXPECT().GetString("database.default").Return("mysql").Once()
				driverCallback := func() (contractsdriver.Driver, error) {
					return mockDriver, nil
				}
				mockConfig.EXPECT().Get("database.connections.mysql.via").Return(driverCallback).Once()
				mockDriver.EXPECT().DB().Return(&sql.DB{}, nil).Once()
				mockDriver.EXPECT().Config().Return(database.Config{Driver: "mysql"}).Once()
			},
			expectedPanic: false,
		},
		{
			name:       "Success with specific connection",
			connection: "postgres",
			setup: func(db *DB) {
				driverCallback := func() (contractsdriver.Driver, error) {
					return mockDriver, nil
				}
				mockConfig.EXPECT().Get("database.connections.postgres.via").Return(driverCallback).Once()
				mockDriver.EXPECT().DB().Return(&sql.DB{}, nil).Once()
				mockDriver.EXPECT().Config().Return(database.Config{Driver: "postgres"}).Once()
			},
			expectedPanic: false,
		},
		{
			name:       "Return cached connection",
			connection: "mysql",
			setup: func(db *DB) {
				driverCallback := func() (contractsdriver.Driver, error) {
					return mockDriver, nil
				}
				mockConfig.EXPECT().Get("database.connections.mysql.via").Return(driverCallback).Once()
				mockDriver.EXPECT().DB().Return(&sql.DB{}, nil).Once()
				mockDriver.EXPECT().Config().Return(database.Config{Driver: "mysql"}).Once()

				cachedDB, _ := BuildDB(context.Background(), mockConfig, mockLog, "mysql")
				db.queries = map[string]contractsdb.DB{"mysql": cachedDB}
			},
			expectedPanic: false,
		},
		{
			name:       "Panic on BuildDB error",
			connection: "invalid",
			setup: func(db *DB) {
				mockConfig.EXPECT().Get("database.connections.invalid.via").Return(nil).Once()
				mockLog.EXPECT().Panic(errors.DatabaseConfigNotFound.Error()).Once()
			},
			expectedPanic: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = mocksconfig.NewConfig(t)
			mockDriver = mocksdriver.NewDriver(t)
			mockLog = mockslog.NewLog(t)

			db := NewDB(context.Background(), mockConfig, mockDriver, mockLog, nil)
			test.setup(db)

			if test.expectedPanic {
				assert.NotPanics(t, func() {
					result := db.Connection(test.connection)
					assert.Nil(t, result)
				})
			} else {
				result := db.Connection(test.connection)
				assert.NotNil(t, result)
			}
		})
	}
}
