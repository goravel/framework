package console

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/file"
	supporttime "github.com/goravel/framework/support/time"
	"github.com/goravel/framework/testing/mock"
)

func TestCreate(t *testing.T) {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "database.default").Return("mysql").Times(3)
	mockConfig.On("GetString", "database.connections.mysql.driver").Return("mysql").Once()
	mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4").Twice()
	now := time.Now()
	supporttime.SetTestNow(now)

	MigrateCreator{}.Create("create_users_table", "users", true)
	assert.True(t, file.Exists(fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.Format("20060102150405"), "create_users_table", "up")))
	assert.True(t, file.Exists(fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.Format("20060102150405"), "create_users_table", "down")))
	assert.True(t, file.Remove("database"))
	mockConfig.AssertExpectations(t)
}

func TestPopulateStub(t *testing.T) {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "database.default").Return("mysql").Twice()
	mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4").Twice()

	assert.Equal(t, "DummyTable ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;", MigrateCreator{}.populateStub("DummyTable ENGINE = InnoDB DEFAULT CHARSET = DummyDatabaseCharset;", ""))
	assert.Equal(t, "users ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;", MigrateCreator{}.populateStub("DummyTable ENGINE = InnoDB DEFAULT CHARSET = DummyDatabaseCharset;", "users"))
	mockConfig.AssertExpectations(t)
}
