package docker

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	foundationmocks "github.com/goravel/framework/contracts/foundation/mocks"
	"github.com/goravel/framework/database"
	gormmocks "github.com/goravel/framework/database/gorm/mocks"
)

func TestNewDatabase(t *testing.T) {
	var (
		mockApp            *foundationmocks.Application
		mockConfig         *configmocks.Config
		mockGormInitialize *gormmocks.Initialize
	)

	beforeEach := func() {
		mockConfig = configmocks.NewConfig(t)
		mockApp = foundationmocks.NewApplication(t)
		mockApp.On("MakeConfig").Return(mockConfig).Once()
		mockGormInitialize = gormmocks.NewInitialize(t)
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
				mockConfig.On("GetString", "database.connections.mysql.driver").Return("mysql").Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					app:            mockApp,
					config:         mockConfig,
					connection:     "mysql",
					driver:         NewMysql(mockConfig, "mysql"),
					gormInitialize: mockGormInitialize,
				}
			},
		},
		{
			name:       "success when connection is mysql",
			connection: "mysql",
			setup: func() {
				mockConfig.On("GetString", "database.connections.mysql.driver").Return(contractsorm.DriverMysql.String()).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					app:            mockApp,
					config:         mockConfig,
					connection:     "mysql",
					driver:         NewMysql(mockConfig, "mysql"),
					gormInitialize: mockGormInitialize,
				}
			},
		},
		{
			name:       "success when connection is postgresql",
			connection: "postgresql",
			setup: func() {
				mockConfig.On("GetString", "database.connections.postgresql.driver").Return(contractsorm.DriverPostgresql.String()).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					app:            mockApp,
					config:         mockConfig,
					connection:     "postgresql",
					driver:         NewPostgresql(mockConfig, "postgresql"),
					gormInitialize: mockGormInitialize,
				}
			},
		},
		{
			name:       "success when connection is sqlserver",
			connection: "sqlserver",
			setup: func() {
				mockConfig.On("GetString", "database.connections.sqlserver.driver").Return(contractsorm.DriverSqlserver.String()).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					app:            mockApp,
					config:         mockConfig,
					connection:     "sqlserver",
					driver:         NewSqlserver(mockConfig, "sqlserver"),
					gormInitialize: mockGormInitialize,
				}
			},
		},
		{
			name:       "success when connection is sqlite",
			connection: "sqlite",
			setup: func() {
				mockConfig.On("GetString", "database.connections.sqlite.driver").Return(contractsorm.DriverSqlite.String()).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					app:            mockApp,
					config:         mockConfig,
					connection:     "sqlite",
					driver:         NewSqlite(mockConfig, "sqlite"),
					gormInitialize: mockGormInitialize,
				}
			},
		},
		{
			name:       "error when connection is not exist",
			connection: "mysql",
			setup: func() {
				mockConfig.On("GetString", "database.connections.mysql.driver").Return("").Once()
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
	s.mockApp = foundationmocks.NewApplication(s.T())
	s.mockArtisan = consolemocks.NewArtisan(s.T())
	s.mockConfig = configmocks.NewConfig(s.T())
	s.mockGormInitialize = gormmocks.NewInitialize(s.T())
	s.database = &Database{
		app:            s.mockApp,
		config:         s.mockConfig,
		connection:     "mysql",
		driver:         NewMysql(s.mockConfig, "mysql"),
		gormInitialize: s.mockGormInitialize,
	}
}

func (s *DatabaseTestSuite) TestBuild() {
	if testing.Short() {
		s.T().Skip("Skipping tests of using docker")
	}

	s.mockConfig.On("GetString", "database.connections.mysql.database").Return("goravel").Twice()
	s.mockConfig.On("GetString", "database.connections.mysql.username").Return("root").Twice()
	s.mockConfig.On("GetString", "database.connections.mysql.password").Return("123123").Twice()
	s.mockConfig.On("Add", "database.connections.mysql.host", "127.0.0.1").Once()
	s.mockConfig.On("Add", "database.connections.mysql.port", mock.Anything).Once()
	s.mockConfig.On("Add", "database.connections.mysql.database", "goravel").Once()
	s.mockConfig.On("Add", "database.connections.mysql.username", "root").Once()
	s.mockConfig.On("Add", "database.connections.mysql.password", "123123").Once()
	s.mockGormInitialize.On("InitializeQuery", context.Background(), s.mockConfig, s.database.driver.Name().String()).Return(nil, nil).Once()
	s.mockApp.On("MakeArtisan").Return(s.mockArtisan).Once()
	s.mockArtisan.On("Call", "migrate").Once()
	s.mockApp.On("Singleton", database.BindingOrm, mock.Anything).Once()

	s.Nil(s.database.Build())
}

func (s *DatabaseTestSuite) TestSeed() {
	mockArtisan := consolemocks.NewArtisan(s.T())
	mockArtisan.On("Call", "db:seed").Once()
	s.mockApp.On("MakeArtisan").Return(mockArtisan).Once()

	s.database.Seed()

	mockArtisan = consolemocks.NewArtisan(s.T())
	mockArtisan.On("Call", "db:seed --seeder mock").Once()
	s.mockApp.On("MakeArtisan").Return(mockArtisan).Once()

	s.database.Seed(&MockSeeder{})
}

type MockSeeder struct{}

func (m *MockSeeder) Signature() string {
	return "mock"
}

func (m *MockSeeder) Run() error {
	return nil
}
