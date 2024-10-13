package migration

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

func TestMigrateMakeCommand(t *testing.T) {
	var (
		mockConfig  *mocksconfig.Config
		mockContext *mocksconsole.Context
		mockSchema  *mocksmigration.Schema
	)

	now := carbon.Now()
	carbon.SetTestNow(now)

	beforeEach := func() {
		mockConfig = mocksconfig.NewConfig(t)
		mockContext = mocksconsole.NewContext(t)
		mockSchema = mocksmigration.NewSchema(t)
	}

	tests := []struct {
		name      string
		setup     func()
		assert    func()
		expectErr error
	}{
		{
			name: "the migration name is empty",
			setup: func() {
				mockContext.On("Argument", 0).Return("").Once()
				mockContext.On("Ask", "Enter the migration name", mock.Anything).Return("", errors.New("the migration name cannot be empty")).Once()
			},
			assert:    func() {},
			expectErr: errors.New("the migration name cannot be empty"),
		},
		{
			name: "default driver",
			setup: func() {
				mockContext.On("Argument", 0).Return("create_users_table").Once()
				mockConfig.On("GetString", "database.migrations.driver").Return(contractsmigration.DriverDefault).Once()
				mockConfig.On("GetString", "database.migrations.table").Return("migrations").Once()
			},
			assert: func() {
				migration := fmt.Sprintf("database/migrations/%s_%s.go", now.ToShortDateTimeString(), "create_users_table")

				assert.True(t, file.Exists(migration))
			},
		},
		{
			name: "sql driver",
			setup: func() {
				mockContext.On("Argument", 0).Return("create_users_table").Once()
				mockConfig.On("GetString", "database.default").Return("postgres").Once()
				mockConfig.On("GetString", "database.migrations.driver").Return(contractsmigration.DriverSql).Once()
				mockConfig.On("GetString", "database.connections.postgres.driver").Return("postgres").Once()
				mockConfig.On("GetString", "database.connections.postgres.charset").Return("utf8mb4").Once()
				mockConfig.On("GetString", "database.migrations.table").Return("migrations").Once()
			},
			assert: func() {
				up := fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.ToShortDateTimeString(), "create_users_table", "up")
				down := fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.ToShortDateTimeString(), "create_users_table", "down")

				assert.True(t, file.Exists(up))
				assert.True(t, file.Exists(down))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			migrateMakeCommand := NewMigrateMakeCommand(mockConfig, mockSchema)
			err := migrateMakeCommand.Handle(mockContext)
			assert.Equal(t, test.expectErr, err)

			test.assert()
		})
	}

	defer func() {
		assert.Nil(t, file.Remove("database"))
		carbon.UnsetTestNow()
	}()
}
