package migration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/migration"
	consolemocks "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

func TestMigrateResetCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	testQueries := gorm.NewTestQueries().Queries()
	for driver, testQuery := range testQueries {
		query := testQuery.Query()

		mockConfig := testQuery.MockConfig()
		mockConfig.EXPECT().GetString("database.migrations.table").Return("migrations").Once()
		mockConfig.EXPECT().GetString("database.migrations.driver").Return(contractsmigration.DriverSql).Once()
		mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.charset", testQuery.Docker().Driver().String())).Return("utf8bm4").Once()

		migration.CreateTestMigrations(driver)

		mockContext := consolemocks.NewContext(t)
		mockSchema := mocksmigration.NewSchema(t)

		migrateCommand := NewMigrateCommand(mockConfig, mockSchema)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		migrateResetCommand := NewMigrateResetCommand(mockConfig)
		assert.Nil(t, migrateResetCommand.Handle(mockContext))

		var agent migration.Agent
		err := query.Where("name", "goravel").FirstOrFail(&agent)
		assert.Error(t, err)
	}

	defer assert.Nil(t, file.Remove("database"))
}
