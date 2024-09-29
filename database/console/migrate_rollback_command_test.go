package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/database/gorm"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/env"
)

func TestMigrateRollbackCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	testQueries := gorm.NewTestQueries().Queries()
	for driver, testQuery := range testQueries {
		query := testQuery.Query()
		mockConfig := testQuery.MockConfig()
		createMigrations(driver)

		mockContext := mocksconsole.NewContext(t)
		mockContext.On("Option", "step").Return("1").Once()

		migrateCommand := NewMigrateCommand(mockConfig)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		var agent Agent
		err := query.Where("name", "goravel").FirstOrFail(&agent)
		assert.Nil(t, err)
		assert.True(t, agent.ID > 0)

		migrateRollbackCommand := NewMigrateRollbackCommand(mockConfig)
		assert.Nil(t, migrateRollbackCommand.Handle(mockContext))

		var agent1 Agent
		err = query.Where("name", "goravel").FirstOrFail(&agent1)
		assert.Error(t, err)
	}
}
