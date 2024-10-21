package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
)

func TestMigrateCommand(t *testing.T) {
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
				mockMigrator.EXPECT().Run().Return(nil).Once()
				mockContext.EXPECT().Info("Migration success").Once()
			},
		},
		{
			name: "Sad path - run failed",
			setup: func() {
				mockMigrator.EXPECT().Run().Return(assert.AnError).Once()
				mockContext.EXPECT().Error(errors.MigrationMigrateFailed.Args(assert.AnError).Error()).Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			command := NewMigrateCommand(mockMigrator)
			err := command.Handle(mockContext)

			assert.NoError(t, err)
		})
	}
}
