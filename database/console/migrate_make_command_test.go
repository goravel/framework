package console

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	configmock "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

func TestMigrateMakeCommand(t *testing.T) {
	var (
		mockConfig  *configmock.Config
		mockContext *consolemocks.Context
	)

	now := carbon.Now()
	carbon.SetTestNow(now)

	beforeEach := func() {
		mockConfig = &configmock.Config{}
		mockContext = &consolemocks.Context{}
	}

	afterEach := func() {
		mockConfig.AssertExpectations(t)
		mockContext.AssertExpectations(t)
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
				mockConfig.On("GetString", "database.migration.driver").Return("default").Once()
				mockContext.On("Argument", 0).Return("create_users_table").Once()
			},
			assert: func() {
				migration := fmt.Sprintf("database/migrations/%s_%s.go", now.ToShortDateTimeString(), "create_users_table")

				assert.True(t, file.Exists(migration))
			},
		},
		{
			name: "sql driver",
			setup: func() {
				mockConfig.On("GetString", "database.migration.driver").Return("sql").Once()
				mockConfig.On("GetString", "database.default").Return("mysql").Times(3)
				mockConfig.On("GetString", "database.connections.mysql.driver").Return("mysql").Once()
				mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4").Twice()
				mockContext.On("Argument", 0).Return("create_users_table").Once()
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

			migrateMakeCommand := NewMigrateMakeCommand(mockConfig)
			err := migrateMakeCommand.Handle(mockContext)
			assert.Equal(t, test.expectErr, err)

			test.assert()
			afterEach()
		})
	}

	assert.Nil(t, file.Remove("database"))
}
