package console

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	configmock "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

func TestMigrateMakeCommand(t *testing.T) {
	now := carbon.Now()
	up := fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.ToShortDateTimeString(), "create_failed_jobs_table", "up")
	down := fmt.Sprintf("database/migrations/%s_%s.%s.sql", now.ToShortDateTimeString(), "create_failed_jobs_table", "down")

	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "database.default").Return("mysql").Once()
	mockConfig.On("GetString", "database.connections.mysql.driver").Return("mysql").Once()

	migrateMakeCommand := NewMigrateMakeCommand(mockConfig)
	mockContext := &consolemocks.Context{}
	assert.Nil(t, migrateMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists(up))
	assert.True(t, file.Exists(down))
	assert.Nil(t, file.Remove("database"))

	mockConfig.AssertExpectations(t)
}
