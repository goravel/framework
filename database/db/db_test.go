package db

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksdriver "github.com/goravel/framework/mocks/database/driver"
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
				mockConfig.On("Get", "database.connections.mysql.via").Return(driverCallback)
				mockDriver.On("DB").Return(&sql.DB{}, nil)
				mockDriver.On("Config").Return(database.Config{Driver: "mysql"})
			},
			expectedError: nil,
		},
		{
			name:       "Config Not Found",
			connection: "invalid",
			setup: func() {
				mockConfig.On("Get", "database.connections.invalid.via").Return(nil)
			},
			expectedError: errors.DatabaseConfigNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = mocksconfig.NewConfig(t)
			mockDriver = mocksdriver.NewDriver(t)
			test.setup()

			db, err := BuildDB(mockConfig, test.connection)
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
