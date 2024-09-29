package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/database/gorm"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/env"
)

func TestMigrateStatusCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	testQueries := gorm.NewTestQueries().Queries()
	for driver, testQuery := range testQueries {
		query := testQuery.Query()
		mockConfig := testQuery.MockConfig()
		createMigrations(driver)

		mockContext := consolemocks.NewContext(t)

		migrateCommand := NewMigrateCommand(mockConfig)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		migrateStatusCommand := NewMigrateStatusCommand(mockConfig)
		assert.Nil(t, migrateStatusCommand.Handle(mockContext))

		res, err := query.Table("migrations").Where("dirty", false).Update("dirty", true)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), res.RowsAffected)

		assert.Nil(t, migrateStatusCommand.Handle(mockContext))
	}
}
