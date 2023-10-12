package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	databasecontract "github.com/goravel/framework/contracts/database"
)

type ConfigTestSuite struct {
	suite.Suite
	config     *ConfigImpl
	connection string
	mockConfig *configmock.Config
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, &ConfigTestSuite{
		connection: "mysql",
	})
}

func (s *ConfigTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	s.config = NewConfigImpl(s.mockConfig, s.connection)
}

func (s *ConfigTestSuite) TestFillDefaultForConfigs() {
	host := "localhost"
	port := 3306
	database := "forge"
	username := "root"
	password := "123123"

	tests := []struct {
		name          string
		configs       []databasecontract.Config
		setup         func()
		expectConfigs []databasecontract.Config
	}{
		{
			name:    "success when configs is empty",
			configs: []databasecontract.Config{},
			setup: func() {
				s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", s.connection)).Return("mysql").Once()
			},
		},
		{
			name:    "success when configs have item but key is empty",
			configs: []databasecontract.Config{{}},
			setup: func() {
				s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", s.connection)).Return("mysql").Once()
				s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.host", s.connection)).Return(host).Once()
				s.mockConfig.On("GetInt", fmt.Sprintf("database.connections.%s.port", s.connection)).Return(port).Once()
				s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", s.connection)).Return(database).Once()
				s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.username", s.connection)).Return(username).Once()
				s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.password", s.connection)).Return(password).Once()
			},
			expectConfigs: []databasecontract.Config{
				{
					Host:     host,
					Port:     port,
					Database: database,
					Username: username,
					Password: password,
				},
			},
		},
		{
			name: "success when configs have item",
			configs: []databasecontract.Config{
				{
					Host:     "localhost",
					Port:     3306,
					Database: "forge",
					Username: "root",
					Password: "123123",
				},
			},
			setup: func() {
				s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", s.connection)).Return("mysql").Once()
			},
			expectConfigs: []databasecontract.Config{
				{
					Host:     "localhost",
					Port:     3306,
					Database: "forge",
					Username: "root",
					Password: "123123",
				},
			},
		},
		{
			name: "success when sqlite",
			configs: []databasecontract.Config{
				{
					Database: "forge",
				},
			},
			setup: func() {
				s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", s.connection)).Return("sqlite").Once()
			},
			expectConfigs: []databasecontract.Config{
				{
					Database: "forge",
				},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			configs := s.config.fillDefault(test.configs)
			s.Equal(test.expectConfigs, configs)
			s.mockConfig.AssertExpectations(s.T())
		})
	}
}
