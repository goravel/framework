package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
)

func TestMigrateRollbackCommand(t *testing.T) {
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
			name: "Default value",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(0).Once()
				mockContext.EXPECT().OptionInt("batch").Return(0).Once()
				mockMigrator.EXPECT().Rollback(0, 1).Return(nil).Once()
				mockContext.EXPECT().Success("Migration rollback success").Once()
			},
		},
		{
			name: "step < 0",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(-1).Once()
				mockContext.EXPECT().OptionInt("batch").Return(0).Once()
				mockContext.EXPECT().Error("The step option should be a positive integer").Once()
			},
		},
		{
			name: "batch < 0",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(0).Once()
				mockContext.EXPECT().OptionInt("batch").Return(-1).Once()
				mockContext.EXPECT().Error("The batch option should be a positive integer").Once()
			},
		},
		{
			name: "step > 0 && batch > 0",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(1).Once()
				mockContext.EXPECT().OptionInt("batch").Return(1).Once()
				mockContext.EXPECT().Error("The step and batch options cannot be used together").Once()
			},
		},
		{
			name: "With step",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(2).Once()
				mockContext.EXPECT().OptionInt("batch").Return(0).Once()
				mockMigrator.EXPECT().Rollback(2, 0).Return(nil).Once()
				mockContext.EXPECT().Success("Migration rollback success").Once()
			},
		},
		{
			name: "With batch",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(0).Once()
				mockContext.EXPECT().OptionInt("batch").Return(2).Once()
				mockMigrator.EXPECT().Rollback(0, 2).Return(nil).Once()
				mockContext.EXPECT().Success("Migration rollback success").Once()
			},
		},
		{
			name: "Rollback failed",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(0).Once()
				mockContext.EXPECT().OptionInt("batch").Return(0).Once()
				mockMigrator.EXPECT().Rollback(0, 1).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(errors.MigrationMigrateFailed.Args(assert.AnError).Error()).Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			command := NewMigrateRollbackCommand(mockMigrator)
			err := command.Handle(mockContext)

			assert.NoError(t, err)
		})
	}
}
