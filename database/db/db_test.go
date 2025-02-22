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
				mockConfig.EXPECT().Get("database.connections.mysql.via").Return(driverCallback).Once()
				mockDriver.EXPECT().DB().Return(&sql.DB{}, nil).Once()
				mockDriver.EXPECT().Config().Return(database.Config{Driver: "mysql"}).Once()
				mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()
				mockConfig.EXPECT().GetInt("database.slow_threshold", 200).Return(200).Once()
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

			db, err := BuildDB(mockConfig, nil, test.connection)
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
