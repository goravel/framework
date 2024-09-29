package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/orm"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/env"
)

type Agent struct {
	orm.Model
	Name string
}

func TestMigrateCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	testQueries := gorm.NewTestQueries().Queries()
	for driver, testQuery := range testQueries {
		query := testQuery.Query()
		mockConfig := testQuery.MockConfig()
		createMigrations(driver)

		migrateCommand := NewMigrateCommand(mockConfig)
		mockContext := &mocksconsole.Context{}
		assert.Nil(t, migrateCommand.Handle(mockContext))

		var agent Agent
		assert.Nil(t, query.Where("name", "goravel").First(&agent))
		assert.True(t, agent.ID > 0)
	}
}
