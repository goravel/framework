package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	mocksconfig "github.com/goravel/framework/mocks/config"
)

type ConfigTestSuite struct {
	suite.Suite
	configBuilder *ConfigBuilder
	connection    string
	mockConfig    *mocksconfig.Config
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, &ConfigTestSuite{
		connection: "mysql",
	})
}

func (s *ConfigTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.configBuilder = NewConfigBuilder(s.mockConfig, s.connection)
}

func (s *ConfigTestSuite) TestReads() {
	database := "forge"
	prefix := "goravel_"
	singular := false

	// Test when configs is empty
	s.mockConfig.EXPECT().Get("database.connections.mysql.read").Return(nil).Once()
	s.Nil(s.configBuilder.Reads())

	// Test when configs is not empty
	s.mockConfig.EXPECT().Get("database.connections.mysql.read").Return([]contractsdatabase.Config{
		{
			Database: database,
		},
	}).Once()
	s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", s.connection)).Return(prefix).Once()
	s.mockConfig.EXPECT().GetBool(fmt.Sprintf("database.connections.%s.singular", s.connection)).Return(singular).Once()
	s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", s.connection)).Return(contractsdatabase.DriverSqlite.String()).Once()

	s.Equal([]contractsdatabase.FullConfig{
		{
			Connection: s.connection,
			Driver:     contractsdatabase.DriverSqlite,
			Prefix:     prefix,
			Config: contractsdatabase.Config{
				Database: database,
			},
		},
	}, s.configBuilder.Reads())
}

func (s *ConfigTestSuite) TestWrites() {
	database := "forge"
	prefix := "goravel_"
	singular := false

	// Test when configBuilder is empty
	s.mockConfig.EXPECT().Get("database.connections.mysql.write").Return(nil).Once()
	s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", s.connection)).Return(contractsdatabase.DriverSqlite.String()).Once()
	s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.database", s.connection)).Return(database).Once()
	s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", s.connection)).Return(prefix).Once()
	s.mockConfig.EXPECT().GetBool(fmt.Sprintf("database.connections.%s.singular", s.connection)).Return(singular).Once()

	s.Equal([]contractsdatabase.FullConfig{
		{
			Connection: s.connection,
			Driver:     contractsdatabase.DriverSqlite,
			Prefix:     prefix,
			Config: contractsdatabase.Config{
				Database: database,
			},
		},
	}, s.configBuilder.Writes())

	// Test when configBuilder is not empty
	s.mockConfig.EXPECT().Get("database.connections.mysql.write").Return([]contractsdatabase.Config{
		{
			Database: database,
		},
	}).Once()
	s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", s.connection)).Return(contractsdatabase.DriverSqlite.String()).Once()
	s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", s.connection)).Return(prefix).Once()
	s.mockConfig.EXPECT().GetBool(fmt.Sprintf("database.connections.%s.singular", s.connection)).Return(singular).Once()

	s.Equal([]contractsdatabase.FullConfig{
		{
			Connection: s.connection,
			Driver:     contractsdatabase.DriverSqlite,
			Prefix:     prefix,
			Config: contractsdatabase.Config{
				Database: database,
			},
		},
	}, s.configBuilder.Writes())
}

func (s *ConfigTestSuite) TestFillDefault() {
	host := "localhost"
	port := 3306
	database := "forge"
	username := "root"
	password := "123123"
	prefix := "goravel_"
	singular := false
	charset := "utf8mb4"
	loc := "Local"

	tests := []struct {
		name          string
		configs       []contractsdatabase.Config
		setup         func()
		expectConfigs []contractsdatabase.FullConfig
	}{
		{
			name:    "success when configs is empty",
			setup:   func() {},
			configs: []contractsdatabase.Config{},
		},
		{
			name:    "success when configs have item but key is empty",
			configs: []contractsdatabase.Config{{}},
			setup: func() {
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", s.connection)).Return(prefix).Once()
				s.mockConfig.EXPECT().GetBool(fmt.Sprintf("database.connections.%s.singular", s.connection)).Return(singular).Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", s.connection)).Return("mysql").Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.host", s.connection)).Return(host).Once()
				s.mockConfig.EXPECT().GetInt(fmt.Sprintf("database.connections.%s.port", s.connection)).Return(port).Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.database", s.connection)).Return(database).Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.username", s.connection)).Return(username).Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.password", s.connection)).Return(password).Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.charset", s.connection)).Return(charset).Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.loc", s.connection)).Return(loc).Once()
			},
			expectConfigs: []contractsdatabase.FullConfig{
				{
					Connection: s.connection,
					Driver:     contractsdatabase.DriverMysql,
					Prefix:     prefix,
					Singular:   singular,
					Charset:    charset,
					Loc:        loc,
					Config: contractsdatabase.Config{
						Host:     host,
						Port:     port,
						Database: database,
						Username: username,
						Password: password,
					},
				},
			},
		},
		{
			name: "success when configs have item",
			configs: []contractsdatabase.Config{
				{
					Host:     host,
					Port:     port,
					Database: database,
					Username: username,
					Password: password,
				},
			},
			setup: func() {
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", s.connection)).Return("mysql").Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", s.connection)).Return(prefix).Once()
				s.mockConfig.EXPECT().GetBool(fmt.Sprintf("database.connections.%s.singular", s.connection)).Return(singular).Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.charset", s.connection)).Return(charset).Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.loc", s.connection)).Return(loc).Once()
			},
			expectConfigs: []contractsdatabase.FullConfig{
				{
					Connection: s.connection,
					Driver:     contractsdatabase.DriverMysql,
					Prefix:     prefix,
					Singular:   singular,
					Charset:    charset,
					Loc:        loc,
					Config: contractsdatabase.Config{
						Database: database,
						Host:     host,
						Port:     port,
						Username: username,
						Password: password,
					},
				},
			},
		},
		{
			name: "success when sqlite",
			configs: []contractsdatabase.Config{
				{
					Database: database,
				},
			},
			setup: func() {
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", s.connection)).Return(prefix).Once()
				s.mockConfig.EXPECT().GetBool(fmt.Sprintf("database.connections.%s.singular", s.connection)).Return(singular).Once()
				s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", s.connection)).Return("sqlite").Once()
			},
			expectConfigs: []contractsdatabase.FullConfig{
				{
					Connection: s.connection,
					Driver:     contractsdatabase.DriverSqlite,
					Prefix:     prefix,
					Singular:   singular,
					Config: contractsdatabase.Config{
						Database: database,
					},
				},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			configs := s.configBuilder.fillDefault(test.configs)

			s.Equal(test.expectConfigs, configs)
		})
	}
}
