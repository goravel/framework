package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
)

func TestMigrateFreshCommand(t *testing.T) {
	var (
		mockArtisan  *mocksconsole.Artisan
		mockContext  *mocksconsole.Context
		mockMigrator *mocksmigration.Migrator
	)

	beforeEach := func() {
		mockArtisan = mocksconsole.NewArtisan(t)
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
				mockMigrator.EXPECT().Fresh().Return(nil).Once()
				mockContext.EXPECT().OptionBool("seed").Return(true).Once()
				mockContext.EXPECT().OptionSlice("seeder").Return([]string{"UserSeeder", "AgentSeeder"}).Once()
				mockArtisan.EXPECT().Call("db:seed --seeder UserSeeder,AgentSeeder").Once()
			},
		},
		{
			name: "Sad path - Fresh failed",
			setup: func() {
				mockMigrator.EXPECT().Fresh().Return(assert.AnError).Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			command := NewMigrateFreshCommand(mockArtisan, mockMigrator)
			err := command.Handle(mockContext)

			assert.NoError(t, err)
		})
	}
}
