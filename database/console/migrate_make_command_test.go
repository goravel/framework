package console

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
	supporttime "github.com/goravel/framework/support/time"
)

func TestMigrateMakeCommand(t *testing.T) {
	now := time.Now()
	supporttime.SetTestNow(now)
	up := fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.Format("20060102150405"), "create_users_table", "up")
	down := fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.Format("20060102150405"), "create_users_table", "down")

	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "database.default").Return("mysql").Times(3)
	mockConfig.On("GetString", "database.connections.mysql.driver").Return("mysql").Once()
	mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4").Twice()

	migrateMakeCommand := NewMigrateMakeCommand(mockConfig)
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	assert.Nil(t, migrateMakeCommand.Handle(mockContext))
	assert.False(t, file.Exists(up))
	assert.False(t, file.Exists(down))

	mockContext.On("Argument", 0).Return("create_users_table").Once()
	assert.Nil(t, migrateMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists(up))
	assert.True(t, file.Exists(down))
	assert.True(t, file.Remove("database"))

	mockConfig.AssertExpectations(t)
}
