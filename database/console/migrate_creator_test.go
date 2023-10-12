package console

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

func TestCreate(t *testing.T) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "database.default").Return("mysql").Times(3)
	mockConfig.On("GetString", "database.connections.mysql.driver").Return("mysql").Once()
	mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4").Twice()
	now := carbon.Now()
	carbon.SetTestNow(carbon.Now())

	migrateCreator := NewMigrateCreator(mockConfig)
	assert.Nil(t, migrateCreator.Create("create_users_table", "users", true))
	assert.True(t, file.Exists(fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.ToShortDateTimeString(), "create_users_table", "up")))
	assert.True(t, file.Exists(fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.ToShortDateTimeString(), "create_users_table", "down")))
	assert.Nil(t, file.Remove("database"))
	carbon.UnsetTestNow()

	mockConfig.AssertExpectations(t)
}

func TestPopulateStub(t *testing.T) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "database.default").Return("mysql").Twice()
	mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4").Twice()

	migrateCreator := NewMigrateCreator(mockConfig)
	assert.Equal(t, "DummyTable ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;", migrateCreator.populateStub("DummyTable ENGINE = InnoDB DEFAULT CHARSET = DummyDatabaseCharset;", ""))
	assert.Equal(t, "users ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;", migrateCreator.populateStub("DummyTable ENGINE = InnoDB DEFAULT CHARSET = DummyDatabaseCharset;", "users"))
	mockConfig.AssertExpectations(t)
}
