package console

import (
	"testing"

	"github.com/stretchr/testify/suite"

	frameworkerrors "github.com/goravel/framework/errors"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
)

type MaintenanceModeTestSuite struct {
	suite.Suite
	mockCache   *mockscache.Cache
	mockConfig  *mocksconfig.Config
	mockStorage *mocksfilesystem.Storage
}

func TestMaintenanceModeTestSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceModeTestSuite))
}

func (s *MaintenanceModeTestSuite) SetupTest() {
	s.mockCache = mockscache.NewCache(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockStorage = mocksfilesystem.NewStorage(s.T())
}

func (s *MaintenanceModeTestSuite) maintenance() *MaintenanceMode {
	return NewMaintenanceMode(s.mockConfig, s.mockCache, s.mockStorage)
}

func (s *MaintenanceModeTestSuite) TestGetWithFileDriver() {
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("file").Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(true).Once()
	s.mockStorage.EXPECT().GetBytes("framework/maintenance.json").Return([]byte(`{"status":503}`), nil).Once()

	content, exists, err := s.maintenance().Get()

	s.Nil(err)
	s.True(exists)
	s.Equal([]byte(`{"status":503}`), content)
}

func (s *MaintenanceModeTestSuite) TestPutWithCacheDriver() {
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("cache").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_STORE").Return("").Once()
	s.mockCache.EXPECT().Forever("framework:maintenance", `{"status":503}`).Return(true).Once()

	err := s.maintenance().Put(`{"status":503}`)

	s.Nil(err)
}

func (s *MaintenanceModeTestSuite) TestDeleteWithNamedCacheStore() {
	mockCacheDriver := mockscache.NewDriver(s.T())
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("cache").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_STORE").Return("redis").Once()
	s.mockCache.EXPECT().Store("redis").Return(mockCacheDriver).Once()
	mockCacheDriver.EXPECT().Has("framework:maintenance").Return(true).Once()
	mockCacheDriver.EXPECT().Forget("framework:maintenance").Return(true).Once()

	deleted, err := s.maintenance().Delete()

	s.Nil(err)
	s.True(deleted)
}

func (s *MaintenanceModeTestSuite) TestGetWithUnsupportedDriver() {
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("redis").Once()

	content, exists, err := s.maintenance().Get()

	s.Nil(content)
	s.False(exists)
	s.ErrorIs(err, frameworkerrors.MaintenanceDriverNotSupported)
}
