package middleware

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http"
	"github.com/goravel/framework/testing/mock"

	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockshash "github.com/goravel/framework/mocks/hash"
	mockshttp "github.com/goravel/framework/mocks/http"
)

type MaintenanceTestSuite struct {
	suite.Suite
	mockApp               *mocksfoundation.Application
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
	mockFactory := mock.Factory()
	s.mockApp = mocksfoundation.NewApplication(s.T())
	s.mockCtx = mockshttp.NewContext(s.T())
	s.mockFile = mocksfilesystem.NewFile(s.T())
	s.mockHash = mockFactory.Hash()
	s.mockRequest = mockshttp.NewContextRequest(s.T())
	s.mockResponse = mockshttp.NewContextResponse(s.T())
	s.mockAbortableResponse = mockshttp.NewAbortableResponse(s.T())
	s.mockStorage = mockFactory.Storage()
	http.App = s.mockApp
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_NotUnderMaintenance() {
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Next().Once()
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(false).Once()

	middleware := CheckForMaintenanceMode()
	middleware(s.mockCtx)
}

func (s *MaintenanceTestSuite) TestMaintenaneMode_StorageFilePermissionIssue() {
	err := errors.New("permission denied")
	abortableResponse := mockshttp.NewAbortableResponse(s.T())
	abortableResponse.EXPECT().Abort().Return(err).Once()

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
