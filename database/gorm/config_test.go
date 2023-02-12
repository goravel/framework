package gorm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/testing/mock"
)

func TestFillDefaultForConfigs(t *testing.T) {
	var mockConfig *configmocks.Config
	connection := "mysql"
	host := "localhost"
	port := 3306
	database := "forge"
	username := "root"
	password := "123123"

	tests := []struct {
		description string
		setup       func()
	}{
		{
			description: "success when configs is empty",
			setup: func() {
				mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", connection)).Return("mysql").Once()
				configs := fillDefaultForConfigs(connection, []contractsdatabase.Config{})
				assert.Equal(t, 0, len(configs))
			},
		},
		{
			description: "success when configs have item but key is empty",
			setup: func() {
				mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", connection)).Return("mysql").Once()
				mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.host", connection)).Return(host).Once()
				mockConfig.On("GetInt", fmt.Sprintf("database.connections.%s.port", connection)).Return(port).Once()
				mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", connection)).Return(database).Once()
				mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.username", connection)).Return(username).Once()
				mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.password", connection)).Return(password).Once()

				configs := fillDefaultForConfigs(connection, []contractsdatabase.Config{{}})
				assert.Equal(t, []contractsdatabase.Config{
					{
						Host:     host,
						Port:     port,
						Database: database,
						Username: username,
						Password: password,
					},
				}, configs)
			},
		},
		{
			description: "success when configs have item",
			setup: func() {
				mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", connection)).Return("mysql").Once()
				configs := []contractsdatabase.Config{
					{
						Host:     "localhost",
						Port:     3306,
						Database: "forge",
						Username: "root",
						Password: "123123",
					},
				}
				newConfigs := fillDefaultForConfigs(connection, configs)
				assert.Equal(t, configs, newConfigs)
			},
		},
		{
			description: "success when sqlite",
			setup: func() {
				mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", connection)).Return("sqlite").Once()
				configs := []contractsdatabase.Config{
					{
						Database: "forge",
					},
				}
				newConfigs := fillDefaultForConfigs(connection, configs)
				assert.Equal(t, configs, newConfigs)
			},
		},
	}

	for _, test := range tests {
		mockConfig = mock.Config()
		test.setup()
		mockConfig.AssertExpectations(t)
	}
}
