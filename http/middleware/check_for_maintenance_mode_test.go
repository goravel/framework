package middleware

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http"

	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockshash "github.com/goravel/framework/mocks/hash"
	mockshttp "github.com/goravel/framework/mocks/http"
)

type MaintenanceTestSuite struct {
	suite.Suite
	mockApp               *mocksfoundation.Application
	mockCache             *mockscache.Cache
	mockConfig            *mocksconfig.Config
	mockCtx               *mockshttp.Context
	mockFile              *mocksfilesystem.File
	mockHash              *mockshash.Hash
	mockStorage           *mocksfilesystem.Storage
	mockRequest           *mockshttp.ContextRequest
	mockResponse          *mockshttp.ContextResponse
	mockAbortableResponse *mockshttp.AbortableResponse
}

func TestMaintenanceModeSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceTestSuite))
}

func (s *MaintenanceTestSuite) SetupTest() {
	s.mockApp = mocksfoundation.NewApplication(s.T())
	s.mockCache = mockscache.NewCache(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockCtx = mockshttp.NewContext(s.T())
	s.mockFile = mocksfilesystem.NewFile(s.T())
	s.mockHash = mockshash.NewHash(s.T())
	s.mockRequest = mockshttp.NewContextRequest(s.T())
	s.mockResponse = mockshttp.NewContextResponse(s.T())
	s.mockAbortableResponse = mockshttp.NewAbortableResponse(s.T())
	s.mockStorage = mocksfilesystem.NewStorage(s.T())
	http.App = s.mockApp
}

func (s *MaintenanceTestSuite) expectFileMaintenanceDriver() {
	s.mockApp.EXPECT().MakeConfig().Return(s.mockConfig).Once()
	s.mockApp.EXPECT().MakeCache().Return(s.mockCache).Once()
	s.mockApp.EXPECT().MakeStorage().Return(s.mockStorage).Once()
	s.mockApp.EXPECT().MakeHash().Return(s.mockHash).Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("file").Once()
}

func (s *MaintenanceTestSuite) expectCacheMaintenanceDriver() {
	s.mockApp.EXPECT().MakeConfig().Return(s.mockConfig).Once()
	s.mockApp.EXPECT().MakeCache().Return(s.mockCache).Once()
	s.mockApp.EXPECT().MakeStorage().Return(s.mockStorage).Once()
	s.mockApp.EXPECT().MakeHash().Return(s.mockHash).Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("cache").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_STORE").Return("").Once()
}

func (s *MaintenanceTestSuite) expectMissingMaintenanceDependency(configMissing, cacheMissing, storageMissing, hashMissing bool) {
	if configMissing {
		s.mockApp.EXPECT().MakeConfig().Return(nil).Once()
	} else {
		s.mockApp.EXPECT().MakeConfig().Return(s.mockConfig).Once()
	}

	if cacheMissing {
		s.mockApp.EXPECT().MakeCache().Return(nil).Once()
	} else {
		s.mockApp.EXPECT().MakeCache().Return(s.mockCache).Once()
	}

	if storageMissing {
		s.mockApp.EXPECT().MakeStorage().Return(nil).Once()
	} else {
		s.mockApp.EXPECT().MakeStorage().Return(s.mockStorage).Once()
	}

	if hashMissing {
		s.mockApp.EXPECT().MakeHash().Return(nil).Once()
	} else {
		s.mockApp.EXPECT().MakeHash().Return(s.mockHash).Once()
	}
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_NotUnderMaintenance() {
	s.expectFileMaintenanceDriver()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Next().Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(false).Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_MissingConfigPassesThrough() {
	s.expectMissingMaintenanceDependency(true, false, false, false)
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Next().Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_MissingCachePassesThrough() {
	s.expectMissingMaintenanceDependency(false, true, false, false)
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Next().Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_MissingStoragePassesThrough() {
	s.expectMissingMaintenanceDependency(false, false, true, false)
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Next().Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_MissingHashPassesThrough() {
	s.expectMissingMaintenanceDependency(false, false, false, true)
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Next().Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_StorageFilePermissionIssue() {
	err := errors.New("permission denied")
	abortableResponse := mockshttp.NewAbortableResponse(s.T())
	abortableResponse.EXPECT().Abort().Return(err).Once()

	s.expectFileMaintenanceDriver()
	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(true).Once()
	s.mockStorage.EXPECT().GetBytes("framework/maintenance.json").Return(nil, err).Once()
	s.mockResponse.EXPECT().String(contractshttp.StatusServiceUnavailable, err.Error()).Return(abortableResponse).Once()

	middleware := CheckForMaintenanceMode()
	assert.PanicsWithError(s.T(), err.Error(), func() {
		middleware(s.mockCtx)
	})
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_StorageFileInvalidJSON() {
	err := errors.New("invalid character 'i' looking for beginning of value")
	s.mockAbortableResponse.EXPECT().Abort().Return(err).Once()

	s.expectFileMaintenanceDriver()
	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(true).Once()
	s.mockStorage.EXPECT().GetBytes("framework/maintenance.json").Return([]byte("invalid json"), nil).Once()
	s.mockResponse.EXPECT().String(contractshttp.StatusServiceUnavailable, err.Error()).Return(s.mockAbortableResponse).Once()

	middleware := CheckForMaintenanceMode()
	assert.PanicsWithError(s.T(), err.Error(), func() {
		middleware(s.mockCtx)
	})
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_SecretDoesNotMatch() {
	s.expectFileMaintenanceDriver()
	s.mockHash.EXPECT().Check("invalid-secret", "hashed-secret").Return(false).Once()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Once()
	s.mockRequest.EXPECT().Query("secret", "").Return("invalid-secret").Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(true).Once()
	s.mockStorage.EXPECT().GetBytes("framework/maintenance.json").Return([]byte(`{"secret":"hashed-secret", "reason": "Under Maintenance", "status": 503}`), nil).Once()
	s.mockAbortableResponse.EXPECT().Abort().Return(nil).Once()
	s.mockResponse.EXPECT().String(contractshttp.StatusServiceUnavailable, "Under Maintenance").Return(s.mockAbortableResponse).Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_SecretMatches() {
	s.expectFileMaintenanceDriver()
	s.mockHash.EXPECT().Check("valid-secret", "hashed-secret").Return(true).Once()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Twice()
	s.mockRequest.EXPECT().Next().Once()
	s.mockRequest.EXPECT().Query("secret", "").Return("valid-secret").Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(true).Once()
	s.mockStorage.EXPECT().GetBytes("framework/maintenance.json").Return([]byte(`{"secret":"hashed-secret", "reason": "Under Maintenance", "status": 503}`), nil).Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_Redirect() {
	s.expectFileMaintenanceDriver()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Twice()
	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Once()
	s.mockRequest.EXPECT().Query("secret", "").Return("").Once()
	s.mockRequest.EXPECT().Path().Return("/").Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(true).Once()
	s.mockStorage.EXPECT().GetBytes("framework/maintenance.json").Return([]byte(`{"redirect": "/maintenance", "status": 503}`), nil).Once()
	s.mockResponse.EXPECT().Redirect(contractshttp.StatusTemporaryRedirect, "/maintenance").Return(s.mockAbortableResponse).Once()
	s.mockAbortableResponse.EXPECT().Abort().Return(nil).Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_Render() {
	s.expectFileMaintenanceDriver()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Twice()
	s.mockRequest.EXPECT().Query("secret", "").Return("").Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(true).Once()
	s.mockStorage.EXPECT().GetBytes("framework/maintenance.json").Return([]byte(`{"render": "maintenance", "status": 503}`), nil).Once()
	s.mockRequest.EXPECT().Abort(contractshttp.StatusServiceUnavailable).Once()

	mocksView := mockshttp.NewResponseView(s.T())
	mocksHttpResponse := mockshttp.NewResponse(s.T())
	mocksHttpResponse.EXPECT().Render().Return(nil).Once()
	mocksView.EXPECT().Make("maintenance").Return(mocksHttpResponse).Once()
	s.mockResponse.EXPECT().View().Return(mocksView).Once()
	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_CacheDriver() {
	s.expectCacheMaintenanceDriver()
	s.mockCache.EXPECT().Has("framework:maintenance").Return(true).Once()
	s.mockCache.EXPECT().GetString("framework:maintenance").Return(`{"reason":"Under Maintenance", "status": 503}`).Once()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Once()
	s.mockRequest.EXPECT().Query("secret", "").Return("").Once()
	s.mockAbortableResponse.EXPECT().Abort().Return(nil).Once()
	s.mockResponse.EXPECT().String(contractshttp.StatusServiceUnavailable, "Under Maintenance").Return(s.mockAbortableResponse).Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_NamedCacheStore() {
	mockCacheDriver := mockscache.NewDriver(s.T())
	s.mockApp.EXPECT().MakeConfig().Return(s.mockConfig).Once()
	s.mockApp.EXPECT().MakeCache().Return(s.mockCache).Once()
	s.mockApp.EXPECT().MakeStorage().Return(s.mockStorage).Once()
	s.mockApp.EXPECT().MakeHash().Return(s.mockHash).Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("cache").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_STORE").Return("redis").Once()
	s.mockCache.EXPECT().Store("redis").Return(mockCacheDriver).Once()
	mockCacheDriver.EXPECT().Has("framework:maintenance").Return(true).Once()
	mockCacheDriver.EXPECT().GetString("framework:maintenance").Return(`{"reason":"Under Maintenance", "status": 503}`).Once()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Once()
	s.mockRequest.EXPECT().Query("secret", "").Return("").Once()
	s.mockAbortableResponse.EXPECT().Abort().Return(nil).Once()
	s.mockResponse.EXPECT().String(contractshttp.StatusServiceUnavailable, "Under Maintenance").Return(s.mockAbortableResponse).Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}
