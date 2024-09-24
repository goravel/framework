package docker

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	contractstesting "github.com/goravel/framework/contracts/testing"
	frameworkdatabase "github.com/goravel/framework/database"
	"github.com/goravel/framework/database/gorm"
	configmocks "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
	gormmocks "github.com/goravel/framework/mocks/database/gorm"
	foundationmocks "github.com/goravel/framework/mocks/foundation"
	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

var (
	testDatabase = "goravel"
	testUsername = "goravel"
	testPassword = "Framework!123"
)

func TestNewDatabase(t *testing.T) {
	var (
		mockApp            *foundationmocks.Application
		mockConfig         *configmocks.Config
		mockGormInitialize *gormmocks.Initialize
	)

	beforeEach := func() {
		mockConfig = &configmocks.Config{}
		mockApp = &foundationmocks.Application{}
		mockApp.On("MakeConfig").Return(mockConfig).Once()
		mockGormInitialize = &gormmocks.Initialize{}
	}

	tests := []struct {
		name         string
		connection   string
		setup        func()
		wantDatabase func() *Database
		wantErr      error
	}{
		{
			name: "success when connection is empty",
			setup: func() {
				mockConfig.On("GetString", "database.default").Return("mysql").Once()
				mockConfig.On("GetString", "database.connections.mysql.driver").Return(contractsorm.DriverMysql.String()).Once()
				mockConfig.On("GetString", "database.connections.mysql.database").Return(testDatabase).Once()
				mockConfig.On("GetString", "database.connections.mysql.username").Return(testUsername).Once()
				mockConfig.On("GetString", "database.connections.mysql.password").Return(testPassword).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					app:            mockApp,
					config:         mockConfig,
					connection:     "mysql",
					driver:         supportdocker.NewMysqlImpl(testDatabase, testUsername, testPassword),
					gormInitialize: mockGormInitialize,
				}
			},
		},
		{
			name:       "success when connection is mysql",
			connection: "mysql",
			setup: func() {
				mockConfig.On("GetString", "database.connections.mysql.driver").Return(contractsorm.DriverMysql.String()).Once()
				mockConfig.On("GetString", "database.connections.mysql.database").Return(testDatabase).Once()
				mockConfig.On("GetString", "database.connections.mysql.username").Return(testUsername).Once()
				mockConfig.On("GetString", "database.connections.mysql.password").Return(testPassword).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					app:            mockApp,
					config:         mockConfig,
					connection:     "mysql",
					driver:         supportdocker.NewMysqlImpl(testDatabase, testUsername, testPassword),
					gormInitialize: mockGormInitialize,
				}
			},
		},
		{
			name:       "success when connection is postgres",
			connection: "postgres",
			setup: func() {
				mockConfig.On("GetString", "database.connections.postgres.driver").Return(contractsorm.DriverPostgres.String()).Once()
				mockConfig.On("GetString", "database.connections.postgres.database").Return(testDatabase).Once()
				mockConfig.On("GetString", "database.connections.postgres.username").Return(testUsername).Once()
				mockConfig.On("GetString", "database.connections.postgres.password").Return(testPassword).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					app:            mockApp,
					config:         mockConfig,
					connection:     "postgres",
					driver:         supportdocker.NewPostgresImpl(testDatabase, testUsername, testPassword),
					gormInitialize: mockGormInitialize,
				}
			},
		},
		{
			name:       "success when connection is sqlserver",
			connection: "sqlserver",
			setup: func() {
				mockConfig.On("GetString", "database.connections.sqlserver.driver").Return(contractsorm.DriverSqlserver.String()).Once()
				mockConfig.On("GetString", "database.connections.sqlserver.database").Return(testDatabase).Once()
				mockConfig.On("GetString", "database.connections.sqlserver.username").Return(testUsername).Once()
				mockConfig.On("GetString", "database.connections.sqlserver.password").Return(testPassword).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					app:            mockApp,
					config:         mockConfig,
					connection:     "sqlserver",
					driver:         supportdocker.NewSqlserverImpl(testDatabase, testUsername, testPassword),
					gormInitialize: mockGormInitialize,
				}
			},
		},
		{
			name:       "success when connection is sqlite",
			connection: "sqlite",
			setup: func() {
				mockConfig.On("GetString", "database.connections.sqlite.driver").Return(contractsorm.DriverSqlite.String()).Once()
				mockConfig.On("GetString", "database.connections.sqlite.database").Return(testDatabase).Once()
				mockConfig.On("GetString", "database.connections.sqlite.username").Return(testUsername).Once()
				mockConfig.On("GetString", "database.connections.sqlite.password").Return(testPassword).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					app:            mockApp,
					config:         mockConfig,
					connection:     "sqlite",
					driver:         supportdocker.NewSqliteImpl(testDatabase),
					gormInitialize: mockGormInitialize,
				}
			},
		},
		{
			name:       "error when connection is not exist",
			connection: "mysql",
			setup: func() {
				mockConfig.On("GetString", "database.connections.mysql.driver").Return("").Once()
				mockConfig.On("GetString", "database.connections.mysql.database").Return(testDatabase).Once()
				mockConfig.On("GetString", "database.connections.mysql.username").Return(testUsername).Once()
				mockConfig.On("GetString", "database.connections.mysql.password").Return(testPassword).Once()
			},
			wantErr: fmt.Errorf("not found database connection: %s", "mysql"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()
			gotDatabase, err := NewDatabase(mockApp, tt.connection, mockGormInitialize)
			if tt.wantDatabase != nil {
				assert.Equal(t, tt.wantDatabase(), gotDatabase)
			}
			assert.Equal(t, tt.wantErr, err)

			mockApp.AssertExpectations(t)
			mockConfig.AssertExpectations(t)
			mockGormInitialize.AssertExpectations(t)
		})
	}
}

type DatabaseTestSuite struct {
	suite.Suite
	mockApp            *foundationmocks.Application
	mockArtisan        *consolemocks.Artisan
	mockConfig         *configmocks.Config
	mockGormInitialize *gormmocks.Initialize
	database           *Database
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}

func (s *DatabaseTestSuite) SetupTest() {
	s.mockApp = &foundationmocks.Application{}
	s.mockArtisan = &consolemocks.Artisan{}
	s.mockConfig = &configmocks.Config{}
	s.mockGormInitialize = &gormmocks.Initialize{}
	s.database = &Database{
		app:            s.mockApp,
		config:         s.mockConfig,
		connection:     "mysql",
		driver:         supportdocker.NewMysqlImpl(testDatabase, testUsername, testPassword),
		gormInitialize: s.mockGormInitialize,
	}
}

func (s *DatabaseTestSuite) TestBuild() {
	if env.IsWindows() {
		s.T().Skip("Skipping tests of using docker")
	}

	s.mockConfig.On("Add", "database.connections.mysql.port", mock.Anything).Once()
	s.mockGormInitialize.On("InitializeQuery", context.Background(), s.mockConfig, s.database.driver.Name().String()).Return(&gorm.QueryImpl{}, nil).Once()
	s.mockApp.On("MakeArtisan").Return(s.mockArtisan).Once()
	s.mockArtisan.On("Call", "migrate").Once()
	s.mockApp.On("Singleton", frameworkdatabase.BindingOrm, mock.Anything).Once()

	s.Nil(s.database.Build())
	s.True(s.database.Config().Port > 0)
	s.Nil(s.database.Stop())

	s.mockConfig.AssertExpectations(s.T())
	s.mockGormInitialize.AssertExpectations(s.T())
	s.mockApp.AssertExpectations(s.T())
	s.mockArtisan.AssertExpectations(s.T())
}

func (s *DatabaseTestSuite) TestConfig() {
	config := s.database.Config()
	s.Equal("127.0.0.1", config.Host)
	s.Equal(testDatabase, config.Database)
	s.Equal(testUsername, config.Username)
	s.Equal(testPassword, config.Password)
}

func (s *DatabaseTestSuite) TestImage() {
	s.database.Image(contractstesting.Image{
		Repository: "mysql",
	})
	s.Equal(&contractstesting.Image{
		Repository: "mysql",
	}, s.database.image)
}

func (s *DatabaseTestSuite) TestSeed() {
	mockArtisan := &consolemocks.Artisan{}
	mockArtisan.On("Call", "db:seed").Once()
	s.mockApp.On("MakeArtisan").Return(mockArtisan).Once()

	s.database.Seed()

	mockArtisan = &consolemocks.Artisan{}
	mockArtisan.On("Call", "db:seed --seeder mock").Once()
	s.mockApp.On("MakeArtisan").Return(mockArtisan).Once()

	s.database.Seed(&MockSeeder{})

	s.mockApp.AssertExpectations(s.T())
	mockArtisan.AssertExpectations(s.T())
}

type MockSeeder struct{}

func (m *MockSeeder) Signature() string {
	return "mock"
}

func (m *MockSeeder) Run() error {
	return nil
}
