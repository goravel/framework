package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
)

func TestMigrateRefreshCommand(t *testing.T) {
	var (
		mockArtisan *mocksconsole.Artisan
		mockContext *mocksconsole.Context
	)

	beforeEach := func() {
		mockArtisan = mocksconsole.NewArtisan(t)
		mockContext = mocksconsole.NewContext(t)
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "step is 0, call migrate reset command failed",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(0).Once()
				mockArtisan.EXPECT().Call("migrate:reset").Return(assert.AnError).Once()
				mockContext.EXPECT().Error(errors.MigrationRefreshFailed.Args(assert.AnError).Error()).Once()
			},
		},
		{
			name: "step > 0, call migrate rollback command failed",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(2).Once()
				mockArtisan.EXPECT().Call("migrate:rollback --step 2").Return(assert.AnError).Once()
				mockContext.EXPECT().Error(errors.MigrationRefreshFailed.Args(assert.AnError).Error()).Once()
			},
		},
		{
			name: "call migrate command failed",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(0).Once()
				mockArtisan.EXPECT().Call("migrate:reset").Return(nil).Once()
				mockArtisan.EXPECT().Call("migrate").Return(assert.AnError).Once()
				mockContext.EXPECT().Error(errors.MigrationRefreshFailed.Args(assert.AnError).Error()).Once()
			},
		},
		{
			name: "call db:seed failed",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(0).Once()
				mockArtisan.EXPECT().Call("migrate:reset").Return(nil).Once()
				mockArtisan.EXPECT().Call("migrate").Return(nil).Once()
				mockContext.EXPECT().OptionBool("seed").Return(true).Once()
				mockContext.EXPECT().OptionSlice("seeder").Return([]string{"a", "b"}).Once()
				mockArtisan.EXPECT().Call("db:seed --seeder a,b").Return(assert.AnError).Once()
				mockContext.EXPECT().Error(errors.MigrationRefreshFailed.Args(assert.AnError).Error()).Once()
			},
		},
		{
			name: "success",
			setup: func() {
				mockContext.EXPECT().OptionInt("step").Return(0).Once()
				mockArtisan.EXPECT().Call("migrate:reset").Return(nil).Once()
				mockArtisan.EXPECT().Call("migrate").Return(nil).Once()
				mockContext.EXPECT().OptionBool("seed").Return(false).Once()
				mockContext.EXPECT().Success("Migration refresh success").Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			command := NewMigrateRefreshCommand(mockArtisan)
			assert.NoError(t, command.Handle(mockContext))
		})
	}
}
