package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
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
				mockMigrator.EXPECT().Create("create_users_table").Return(nil).Once()
				mockContext.EXPECT().Info("Created Migration: create_users_table").Once()
			},
		},
		{
			name: "Happy path - name is not empty",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("create_users_table").Once()
				mockMigrator.EXPECT().Create("create_users_table").Return(nil).Once()
				mockContext.EXPECT().Info("Created Migration: create_users_table").Once()
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
				mockMigrator.EXPECT().Create("create_users_table").Return(assert.AnError).Once()
				mockContext.EXPECT().Error(errors.MigrationCreateFailed.Args(assert.AnError).Error()).Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			migrateMakeCommand := NewMigrateMakeCommand(mockMigrator)
			err := migrateMakeCommand.Handle(mockContext)

			assert.NoError(t, err)
		})
	}
}
