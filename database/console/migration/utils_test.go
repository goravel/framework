package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/migration"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
)

func TestGetDriver(t *testing.T) {
	var (
		mockConfig *mocksconfig.Config
		mockSchema *mocksmigration.Schema
	)

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
				mockConfig.EXPECT().GetString("database.migrations.table").Return("migrations").Once()
			},
			expectDriver: &migration.DefaultDriver{},
		},
		{
			name: "sql driver",
			setup: func() {
				mockConfig.EXPECT().GetString("database.migrations.driver").Return(contractsmigration.DriverSql).Once()
				mockConfig.EXPECT().GetString("database.migrations.table").Return("migrations").Once()
				mockConfig.EXPECT().GetString("database.default").Return("postgres").Once()
				mockConfig.EXPECT().GetString("database.connections.postgres.driver").Return("postgres").Once()
				mockConfig.EXPECT().GetString("database.connections.postgres.charset").Return("utf8mb4").Once()
			},
			expectDriver: &migration.SqlDriver{},
		},
		{
			name: "unsupported driver",
			setup: func() {
				mockConfig.EXPECT().GetString("database.migrations.driver").Return("unsupported").Once()
			},
			expectError: errors.MigrationUnsupportedDriver.Args("unsupported").SetModule(errors.ModuleMigration).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = mocksconfig.NewConfig(t)
			mockSchema = mocksmigration.NewSchema(t)

			test.setup()
			driver, err := GetDriver(mockConfig, mockSchema)
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
