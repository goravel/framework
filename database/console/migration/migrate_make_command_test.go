package migration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	"github.com/goravel/framework/support/file"
)

func TestMigrateMakeCommand(t *testing.T) {
	var (
		mockContext  *mocksconsole.Context
		mockMigrator *mocksmigration.Migrator
	)

	beforeEach := func() {
		mockContext = mocksconsole.NewContext(t)
		mockMigrator = mocksmigration.NewMigrator(t)
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "Happy path",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("").Once()
				mockContext.EXPECT().Ask("Enter the migration name", mock.Anything).Return("create_users_table", nil).Once()
				mockMigrator.EXPECT().Create("create_users_table").Return("", nil).Once()
				mockContext.EXPECT().Success("Created Migration: create_users_table").Once()
				mockContext.EXPECT().Warning(mock.MatchedBy(func(msg string) bool {
					return strings.HasPrefix(msg, "migration register failed:")
				})).Once()
			},
		},
		{
			name: "Happy path - name is not empty",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("create_users_table").Once()
				mockMigrator.EXPECT().Create("create_users_table").Return("", nil).Once()
				mockContext.EXPECT().Success("Created Migration: create_users_table").Once()
				mockContext.EXPECT().Warning(mock.MatchedBy(func(msg string) bool {
					return strings.HasPrefix(msg, "migration register failed:")
				})).Once()
			},
		},
		{
			name: "Sad path - failed to ask",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("").Once()
				mockContext.EXPECT().Ask("Enter the migration name", mock.Anything).Return("", assert.AnError).Once()
				mockContext.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "Sad path - failed to create",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("create_users_table").Once()
				mockMigrator.EXPECT().Create("create_users_table").Return("", assert.AnError).Once()
				mockContext.EXPECT().Error(errors.MigrationCreateFailed.Args(assert.AnError).Error()).Once()
			},
		},
		{
			name: "Register success",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("create_users_table").Once()
				mockMigrator.EXPECT().Create("create_users_table").Return("20240915060148_create_users_table", nil).Once()
				mockContext.EXPECT().Success("Created Migration: create_users_table").Once()
				mockContext.EXPECT().Success("Migration registered successfully").Once()
				assert.NoError(t, file.PutContent("database/kernel.go", `package database

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/migrations"
)

type Kernel struct {
}

func (kernel Kernel) Migrations() []schema.Migration {
	return []schema.Migration{}
}`))
				t.Cleanup(func() {
					assert.True(t, file.Contain("database/kernel.go", `func (kernel Kernel) Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20240915060148CreateUsersTable{},
	}
}`))
					assert.NoError(t, file.Remove("database"))
				})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			migrateMakeCommand := NewMigrateMakeCommand(nil, mockMigrator)
			err := migrateMakeCommand.Handle(mockContext)

			assert.NoError(t, err)
		})
	}
}
