package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
)

type UpCommandTestSuite struct {
	suite.Suite
	mockCache   *mockscache.Cache
	mockConfig  *mocksconfig.Config
	mockStorage *mocksfilesystem.Storage
}

func TestUpCommandTestSuite(t *testing.T) {
	suite.Run(t, new(UpCommandTestSuite))
}

func (s *UpCommandTestSuite) SetupTest() {
	s.mockCache = mockscache.NewCache(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockStorage = mocksfilesystem.NewStorage(s.T())
}

func (s *UpCommandTestSuite) maintenance() *MaintenanceMode {
	return NewMaintenanceMode(s.mockConfig, s.mockCache, s.mockStorage)
}

func (s *UpCommandTestSuite) TestSignature() {
	expected := "up"
	s.Require().Equal(expected, NewUpCommand(s.maintenance()).Signature())
}

func (s *UpCommandTestSuite) TestDescription() {
	expected := "Bring the application out of maintenance mode"
	s.Require().Equal(expected, NewUpCommand(s.maintenance()).Description())
}

func (s *UpCommandTestSuite) TestExtend() {
	cmd := NewUpCommand(s.maintenance())
	got := cmd.Extend()

	s.Empty(got)
}

func (s *UpCommandTestSuite) TestHandle() {
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("file").Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(true).Once()
	s.mockStorage.EXPECT().Delete("framework/maintenance.json").Return(nil).Once()

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Success("The application is up and live now").Once()

	cmd := NewUpCommand(s.maintenance())
	err := cmd.Handle(mockContext)
	assert.Nil(s.T(), err)
}

func (s *UpCommandTestSuite) TestHandleWhenNotDown() {
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("file").Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(false).Once()
	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Error("The application is not in maintenance mode").Once()

	cmd := NewUpCommand(s.maintenance())
	err := cmd.Handle(mockContext)
	assert.Nil(s.T(), err)
}

func (s *UpCommandTestSuite) TestHandleWithCacheDriver() {
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("cache").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_STORE").Return("").Once()
	s.mockCache.EXPECT().Has("framework:maintenance").Return(true).Once()
	s.mockCache.EXPECT().Forget("framework:maintenance").Return(true).Once()

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Success("The application is up and live now").Once()

	cmd := NewUpCommand(s.maintenance())
	err := cmd.Handle(mockContext)
	assert.Nil(s.T(), err)
}

func (s *UpCommandTestSuite) TestHandleWithNamedCacheStore() {
	mockCacheDriver := mockscache.NewDriver(s.T())
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("cache").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_STORE").Return("redis").Once()
	s.mockCache.EXPECT().Store("redis").Return(mockCacheDriver).Once()
	mockCacheDriver.EXPECT().Has("framework:maintenance").Return(true).Once()
	mockCacheDriver.EXPECT().Forget("framework:maintenance").Return(true).Once()

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Success("The application is up and live now").Once()

	cmd := NewUpCommand(s.maintenance())
	err := cmd.Handle(mockContext)
	assert.Nil(s.T(), err)
}
