package console

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	configmock "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

func TestMigrateMakeCommand(t *testing.T) {
	now := carbon.Now()
	up := fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.ToShortDateTimeString(), "create_users_table", "up")
	down := fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.ToShortDateTimeString(), "create_users_table", "down")

	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "database.default").Return("mysql").Times(3)
	mockConfig.On("GetString", "database.connections.mysql.driver").Return("mysql").Once()
	mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4").Twice()

	migrateMakeCommand := NewMigrateMakeCommand(mockConfig)
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the migration name", mock.Anything).Return("", errors.New("the migration name cannot be empty")).Once()
	err := migrateMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "the migration name cannot be empty")
	assert.False(t, file.Exists(up))
	assert.False(t, file.Exists(down))

	mockContext.On("Argument", 0).Return("create_users_table").Once()
	assert.Nil(t, migrateMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists(up))
	assert.True(t, file.Exists(down))
	assert.Nil(t, file.Remove("database"))

	mockConfig.AssertExpectations(t)
}
