package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksdriver "github.com/goravel/framework/mocks/database/driver"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksseeder "github.com/goravel/framework/mocks/database/seeder"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksdocker "github.com/goravel/framework/mocks/testing/docker"
)

func TestNewDatabase(t *testing.T) {
	var (
		mockArtisan        *mocksconsole.Artisan
		mockConfig         *mocksconfig.Config
		mockOrm            *mocksorm.Orm
		mockDatabaseDriver *mocksdriver.Driver
		mockDockerDriver   *mocksdocker.DatabaseDriver
	)

	beforeEach := func() {
		mockArtisan = mocksconsole.NewArtisan(t)
		mockConfig = mocksconfig.NewConfig(t)
		mockOrm = mocksorm.NewOrm(t)
		mockDatabaseDriver = mocksdriver.NewDriver(t)
		mockDockerDriver = mocksdocker.NewDatabaseDriver(t)
	}

	tests := []struct {
		name       string
		connection string
		setup      func()
		wantErr    error
	}{
		{
			name: "success when connection is empty",
			setup: func() {
				mockDatabaseDriver.EXPECT().Docker().Return(mockDockerDriver, nil).Once()
				mockConfig.EXPECT().GetString("database.default").Return("mysql").Once()
				mockConfig.EXPECT().Get("database.connections.mysql.via").Return(func() (contractsdriver.Driver, error) {
					return mockDatabaseDriver, nil
				}).Once()
			},
		},
		{
			name:       "success when connection is not empty",
			connection: "mysql",
			setup: func() {
				mockDatabaseDriver.EXPECT().Docker().Return(mockDockerDriver, nil).Once()
				mockConfig.EXPECT().Get("database.connections.mysql.via").Return(func() (contractsdriver.Driver, error) {
					return mockDatabaseDriver, nil
				}).Once()
			},
		},
		{
			name: "error when Docker returns an error",
			setup: func() {
				mockDatabaseDriver.EXPECT().Docker().Return(nil, assert.AnError).Once()
				mockConfig.EXPECT().GetString("database.default").Return("mysql").Once()
				mockConfig.EXPECT().Get("database.connections.mysql.via").Return(func() (contractsdriver.Driver, error) {
					return mockDatabaseDriver, nil
				}).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "error when init database driver returns an error",
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("mysql").Once()
				mockConfig.EXPECT().Get("database.connections.mysql.via").Return(func() (contractsdriver.Driver, error) {
					return nil, assert.AnError
				}).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "error when database driver doesn't exist",
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("mysql").Once()
				mockConfig.EXPECT().Get("database.connections.mysql.via").Return(func() error {
					return nil
				}).Once()
			},
			wantErr: errors.DatabaseConfigNotFound,
		},
		{
			name: "error when artisan facade is not set",
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("mysql").Once()
			},
			wantErr: errors.ArtisanFacadeNotSet,
		},
		{
			name: "error when config facade is not set",
			setup: func() {
				mockApp.EXPECT().MakeConfig().Return(nil).Once()
			},
			wantErr: errors.ConfigFacadeNotSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()
			gotDatabase, err := NewDatabase(mockArtisan, mockConfig, mockOrm, tt.connection)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				assert.Nil(t, gotDatabase)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, gotDatabase)
			}
		})
	}
}

type DatabaseTestSuite struct {
	suite.Suite
	mockApp            *mocksfoundation.Application
	mockArtisan        *mocksconsole.Artisan
	mockConfig         *mocksconfig.Config
	mockOrm            *mocksorm.Orm
	mockDatabaseDriver *mocksdocker.DatabaseDriver
	database           *Database
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}

func (s *DatabaseTestSuite) SetupTest() {
	s.mockApp = mocksfoundation.NewApplication(s.T())
	s.mockArtisan = mocksconsole.NewArtisan(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockOrm = mocksorm.NewOrm(s.T())
	s.mockDatabaseDriver = mocksdocker.NewDatabaseDriver(s.T())
	s.database = &Database{
		artisan:        s.mockArtisan,
		config:         s.mockConfig,
		connection:     "postgres",
		orm:            s.mockOrm,
		DatabaseDriver: s.mockDatabaseDriver,
	}
}

func (s *DatabaseTestSuite) TestReady() {
	s.mockOrm.EXPECT().Fresh().Once()
	s.mockDatabaseDriver.EXPECT().Ready().Return(nil).Once()

	s.Nil(s.database.Ready())
}

func (s *DatabaseTestSuite) TestSeed() {
	s.mockArtisan.EXPECT().Call("db:seed").Return(nil).Once()
	s.NoError(s.database.Seed())

	s.mockArtisan.EXPECT().Call("db:seed --seeder mock").Return(nil).Once()
	mockSeeder := mocksseeder.NewSeeder(s.T())
	mockSeeder.EXPECT().Signature().Return("mock").Once()
	s.NoError(s.database.Seed(mockSeeder))

	s.mockArtisan.EXPECT().Call("db:seed").Return(assert.AnError).Once()
	s.EqualError(s.database.Seed(), assert.AnError.Error())
}
