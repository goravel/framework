package migration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/file"
)

func TestMigrateMakeCommand(t *testing.T) {
	var (
		mockApp      *mocksfoundation.Application
		mockContext  *mocksconsole.Context
		mockMigrator *mocksmigration.Migrator
	)

	beforeEach := func() {
		mockApp = mocksfoundation.NewApplication(t)
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
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockMigrator.EXPECT().Create("create_users_table", "").Return("", nil).Once()
				mockContext.EXPECT().Success("Created Migration: create_users_table").Once()
				mockApp.EXPECT().DatabasePath("kernel.go").Return("database/kernel.go").Once()
				mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, errors.MigrationRegisterFailed.Error())
				})).Once()
			},
		},
		{
			name: "Happy path - name is not empty",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("create_users_table").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockMigrator.EXPECT().Create("create_users_table", "").Return("", nil).Once()
				mockContext.EXPECT().Success("Created Migration: create_users_table").Once()
				mockApp.EXPECT().DatabasePath("kernel.go").Return("database/kernel.go").Once()
				mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, errors.MigrationRegisterFailed.Error())
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
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockMigrator.EXPECT().Create("create_users_table", "").Return("", assert.AnError).Once()
				mockContext.EXPECT().Error(errors.MigrationCreateFailed.Args(assert.AnError).Error()).Once()
			},
		},
		{
			name: "Register success",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("create_users_table").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockMigrator.EXPECT().Create("create_users_table", "").Return("20240915060148_create_users_table", nil).Once()
				mockContext.EXPECT().Success("Created Migration: create_users_table").Once()
				mockApp.EXPECT().DatabasePath("kernel.go").Return("database/kernel.go").Once()
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

			migrateMakeCommand := NewMigrateMakeCommand(mockApp, mockMigrator)
			err := migrateMakeCommand.Handle(mockContext)

			assert.NoError(t, err)
		})
	}
}

func TestMigrateMakeCommand_WithBootstrapSetup(t *testing.T) {
	var (
		mockContext  *mocksconsole.Context
		mockMigrator *mocksmigration.Migrator
	)

	// Create bootstrap/app.go to trigger IsBootstrapSetup() == true
	bootstrapContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/contracts/database/schema"
)

func Boot() {
	foundation.Setup().WithMigrations([]schema.Migration{}).Run()
}
`
	assert.NoError(t, file.PutContent("bootstrap/app.go", bootstrapContent))

	// Cleanup after test
	defer func() {
		assert.NoError(t, file.Remove("bootstrap"))
		assert.NoError(t, file.Remove("database"))
	}()

	t.Run("Bootstrap setup - successful registration", func(t *testing.T) {
		mockContext = mocksconsole.NewContext(t)
		mockMigrator = mocksmigration.NewMigrator(t)
		mockApp := mocksfoundation.NewApplication(t)

		mockContext.EXPECT().Argument(0).Return("create_posts_table").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockMigrator.EXPECT().Create("create_posts_table", "").Return("20240915060148_create_posts_table", nil).Once()
		mockContext.EXPECT().Success("Created Migration: create_posts_table").Once()
		mockContext.EXPECT().Success("Migration registered successfully").Once()

		migrateMakeCommand := NewMigrateMakeCommand(mockApp, mockMigrator)
		err := migrateMakeCommand.Handle(mockContext)

		assert.NoError(t, err)

		// Verify bootstrap/app.go was updated with the migration
		bootstrapUpdated, err := file.GetContent("bootstrap/app.go")
		assert.NoError(t, err)
		assert.Contains(t, bootstrapUpdated, "migrations.M20240915060148CreatePostsTable{}")
	})

	t.Run("Bootstrap setup - registration failed", func(t *testing.T) {
		// Reset bootstrap/app.go with invalid syntax
		invalidBootstrapContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
}
`
		assert.NoError(t, file.PutContent("bootstrap/app.go", invalidBootstrapContent))

		mockContext = mocksconsole.NewContext(t)
		mockMigrator = mocksmigration.NewMigrator(t)
		mockApp := mocksfoundation.NewApplication(t)

		mockContext.EXPECT().Argument(0).Return("create_comments_table").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockMigrator.EXPECT().Create("create_comments_table", "").Return("20240915060149_create_comments_table", nil).Once()
		mockContext.EXPECT().Success("Created Migration: create_comments_table").Once()
		mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
			return strings.Contains(msg, errors.MigrationRegisterFailed.Error())
		})).Once()

		migrateMakeCommand := NewMigrateMakeCommand(mockApp, mockMigrator)
		err := migrateMakeCommand.Handle(mockContext)

		assert.NoError(t, err)
	})
}
