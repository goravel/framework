package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/database/gorm"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/env"
)

func TestMigrateResetCommand(t *testing.T) {
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

		migrateResetCommand := NewMigrateResetCommand(mockConfig)
		assert.Nil(t, migrateResetCommand.Handle(mockContext))

		var agent Agent
		err := query.Where("name", "goravel").FirstOrFail(&agent)
		assert.Error(t, err)
	}
}
