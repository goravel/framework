package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
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
		mockApp     *mocksfoundation.Application
		mockArtisan *mocksconsole.Artisan
		mockConfig  *mocksconfig.Config
		mockOrm     *mocksorm.Orm
	)

	beforeEach := func() {
		mockApp = mocksfoundation.NewApplication(t)
		mockArtisan = mocksconsole.NewArtisan(t)
		mockConfig = mocksconfig.NewConfig(t)
		mockOrm = mocksorm.NewOrm(t)
		mockApp.EXPECT().MakeArtisan().Return(mockArtisan).Once()
		mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
		mockApp.EXPECT().MakeOrm().Return(mockOrm).Once()
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
				mockConfig.EXPECT().GetString("database.default").Return("mysql").Once()
				mockConfig.EXPECT().GetString("database.connections.mysql.driver").Return(contractsdatabase.DriverMysql.String()).Once()
				mockConfig.EXPECT().GetString("database.connections.mysql.database").Return(testDatabase).Once()
				mockConfig.EXPECT().GetString("database.connections.mysql.username").Return(testUsername).Once()
				mockConfig.EXPECT().GetString("database.connections.mysql.password").Return(testPassword).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					artisan:        mockArtisan,
					config:         mockConfig,
					connection:     "mysql",
					orm:            mockOrm,
					DatabaseDriver: supportdocker.NewMysqlImpl(testDatabase, testUsername, testPassword),
				}
			},
		},
		{
			name:       "success when connection is mysql",
			connection: "mysql",
			setup: func() {
				mockConfig.EXPECT().GetString("database.connections.mysql.driver").Return(contractsdatabase.DriverMysql.String()).Once()
				mockConfig.EXPECT().GetString("database.connections.mysql.database").Return(testDatabase).Once()
				mockConfig.EXPECT().GetString("database.connections.mysql.username").Return(testUsername).Once()
				mockConfig.EXPECT().GetString("database.connections.mysql.password").Return(testPassword).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					artisan:        mockArtisan,
					config:         mockConfig,
					connection:     "mysql",
					orm:            mockOrm,
					DatabaseDriver: supportdocker.NewMysqlImpl(testDatabase, testUsername, testPassword),
				}
			},
		},
		{
			name:       "success when connection is postgres",
			connection: "postgres",
			setup: func() {
				mockConfig.EXPECT().GetString("database.connections.postgres.driver").Return(contractsdatabase.DriverPostgres.String()).Once()
				mockConfig.EXPECT().GetString("database.connections.postgres.database").Return(testDatabase).Once()
				mockConfig.EXPECT().GetString("database.connections.postgres.username").Return(testUsername).Once()
				mockConfig.EXPECT().GetString("database.connections.postgres.password").Return(testPassword).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					artisan:        mockArtisan,
					config:         mockConfig,
					connection:     "postgres",
					orm:            mockOrm,
					DatabaseDriver: supportdocker.NewPostgresImpl(testDatabase, testUsername, testPassword),
				}
			},
		},
		{
			name:       "success when connection is sqlserver",
			connection: "sqlserver",
			setup: func() {
				mockConfig.EXPECT().GetString("database.connections.sqlserver.driver").Return(contractsdatabase.DriverSqlserver.String()).Once()
				mockConfig.EXPECT().GetString("database.connections.sqlserver.database").Return(testDatabase).Once()
				mockConfig.EXPECT().GetString("database.connections.sqlserver.username").Return(testUsername).Once()
				mockConfig.EXPECT().GetString("database.connections.sqlserver.password").Return(testPassword).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					artisan:        mockArtisan,
					config:         mockConfig,
					connection:     "sqlserver",
					orm:            mockOrm,
					DatabaseDriver: supportdocker.NewSqlserverImpl(testDatabase, testUsername, testPassword),
				}
			},
		},
		{
			name:       "success when connection is sqlite",
			connection: "sqlite",
			setup: func() {
				mockConfig.EXPECT().GetString("database.connections.sqlite.driver").Return(contractsdatabase.DriverSqlite.String()).Once()
				mockConfig.EXPECT().GetString("database.connections.sqlite.database").Return(testDatabase).Once()
				mockConfig.EXPECT().GetString("database.connections.sqlite.username").Return(testUsername).Once()
				mockConfig.EXPECT().GetString("database.connections.sqlite.password").Return(testPassword).Once()
			},
			wantDatabase: func() *Database {
				return &Database{
					artisan:        mockArtisan,
					config:         mockConfig,
					connection:     "sqlite",
					orm:            mockOrm,
					DatabaseDriver: supportdocker.NewSqliteImpl(testDatabase),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()
			gotDatabase, err := NewDatabase(mockApp, tt.connection)

			assert.Nil(t, err)
			assert.Equal(t, tt.wantDatabase(), gotDatabase)
		})
	}
}

type DatabaseTestSuite struct {
	suite.Suite
	mockApp     *mocksfoundation.Application
	mockArtisan *mocksconsole.Artisan
	mockConfig  *mocksconfig.Config
	mockOrm     *mocksorm.Orm
	database    *Database
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}

func (s *DatabaseTestSuite) SetupTest() {
	s.mockApp = mocksfoundation.NewApplication(s.T())
	s.mockArtisan = mocksconsole.NewArtisan(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockOrm = mocksorm.NewOrm(s.T())
	s.database = &Database{
		artisan:        s.mockArtisan,
		config:         s.mockConfig,
		connection:     "postgres",
		orm:            s.mockOrm,
		DatabaseDriver: supportdocker.NewPostgresImpl(testDatabase, testUsername, testPassword),
	}
}

func (s *DatabaseTestSuite) TestBuild() {
	if env.IsWindows() {
		s.T().Skip("Skipping tests of using docker")
	}

	s.mockConfig.EXPECT().Add("database.connections.postgres.port", mock.Anything).Once()
	s.mockArtisan.EXPECT().Call("migrate").Once()
	s.mockOrm.EXPECT().Refresh().Once()

	s.Nil(s.database.Build())
	s.True(s.database.Config().Port > 0)
	s.Nil(s.database.Stop())
}

func (s *DatabaseTestSuite) TestConfig() {
	config := s.database.Config()
	s.Equal("127.0.0.1", config.Host)
	s.Equal(testDatabase, config.Database)
	s.Equal(testUsername, config.Username)
	s.Equal(testPassword, config.Password)
}

func (s *DatabaseTestSuite) TestSeed() {
	s.mockArtisan.EXPECT().Call("db:seed").Once()
	s.database.Seed()

	s.mockArtisan.EXPECT().Call("db:seed --seeder mock").Once()
	s.database.Seed(&MockSeeder{})
}

type MockSeeder struct{}

func (m *MockSeeder) Signature() string {
	return "mock"
}

func (m *MockSeeder) Run() error {
	return nil
}
