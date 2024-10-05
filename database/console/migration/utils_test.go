package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/migration"
	mocksconfig "github.com/goravel/framework/mocks/config"
)

func TestGetDriver(t *testing.T) {
	var mockConfig *mocksconfig.Config

	tests := []struct {
		name         string
		setup        func()
		expectDriver contractsmigration.Driver
		expectError  string
	}{
		{
			name: "default driver",
			setup: func() {
				mockConfig.EXPECT().GetString("database.migrations.driver").Return(contractsmigration.DriverDefault).Once()
			},
			expectDriver: migration.NewDefaultDriver(),
		},
		{
			name: "sql driver",
			setup: func() {
				mockConfig.EXPECT().GetString("database.migrations.driver").Return(contractsmigration.DriverSql).Once()
				mockConfig.EXPECT().GetString("database.default").Return("postgres").Once()
				mockConfig.EXPECT().GetString("database.connections.postgres.driver").Return("postgres").Once()
				mockConfig.EXPECT().GetString("database.connections.postgres.charset").Return("utf8mb4").Once()
			},
			expectDriver: migration.NewSqlDriver("postgres", "utf8mb4"),
		},
		{
			name: "unsupported driver",
			setup: func() {
				mockConfig.EXPECT().GetString("database.migrations.driver").Return("unsupported").Once()
			},
			expectError: "unsupported migration driver: unsupported",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = mocksconfig.NewConfig(t)

			test.setup()
			driver, err := GetDriver(mockConfig)
			if test.expectError != "" {
				assert.EqualError(t, err, test.expectError)
				assert.Nil(t, driver)
			} else {
				assert.Nil(t, err)
				assert.IsType(t, test.expectDriver, driver)
			}
		})
	}
}
