package session

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	contractssession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mockconfig "github.com/goravel/framework/mocks/config"
	mocksession "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/session/driver"
)

func MockDriverFactory(mockDriverInstance *mocksession.Driver) func() (contractssession.Driver, error) {
	return func() (contractssession.Driver, error) {
		if mockDriverInstance == nil {
			return nil, fmt.Errorf("mock driver instance not provided")
		}
		return mockDriverInstance, nil
	}
}

type CustomDriver struct{}

func NewCustomDriverFactory() (contractssession.Driver, error) { return &CustomDriver{}, nil }
func (c *CustomDriver) Close() error                           { return nil }
func (c *CustomDriver) Destroy(string) error                   { return nil }
func (c *CustomDriver) Gc(int) error                           { return nil }
func (c *CustomDriver) Open(string, string) error              { return nil }
func (c *CustomDriver) Read(string) (string, error)            { return "", nil }
func (c *CustomDriver) Write(string, string) error             { return nil }

type ManagerTestSuite struct {
	suite.Suite
	mockConfig       *mockconfig.Config
	manager          *Manager
	json             foundation.Json
	mockOtherDriver  *mocksession.Driver
	mockOtherFactory func() (contractssession.Driver, error)
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}

func (s *ManagerTestSuite) SetupTest() {
	s.mockConfig = mockconfig.NewConfig(s.T())
	s.json = json.NewJson()
	s.mockOtherDriver = new(mocksession.Driver)
	s.mockOtherFactory = MockDriverFactory(s.mockOtherDriver)

	s.mockConfig.On("Get", "session.drivers", mock.AnythingOfType("map[string]interface {}")).Return(
		map[string]any{},
	).Maybe()

	s.manager = NewManager(s.mockConfig, s.json)
	s.Require().NotNil(s.manager)
}

func (s *ManagerTestSuite) TearDownSuite() {
	testPath := "storage/framework"
	if _, err := os.Stat(testPath); err == nil {
		os.RemoveAll(testPath)
	}
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ManagerTestSuite) TestDriver_ResolveInternalFileDriver() {

	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.mockConfig.On("GetInt", "session.lifetime").Return(120).Once()
	s.mockConfig.On("GetString", "session.files").Return("storage/framework/sessions_test_manager").Once()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()

	driverInstance, err := s.manager.Driver("file")
	s.Nil(err)
	s.NotNil(driverInstance)

	s.IsType(&driver.File{}, driverInstance, "Expected internal file driver type")
}

func (s *ManagerTestSuite) TestDriver_ResolveDefaultDriver_InternalFile() {
	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.mockConfig.On("GetInt", "session.lifetime").Return(120).Once()
	s.mockConfig.On("GetString", "session.files").Return("storage/framework/sessions_test_manager").Once()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()

	driverInstance, err := s.manager.Driver()
	s.Nil(err)
	s.NotNil(driverInstance)
	s.IsType(&driver.File{}, driverInstance, "Expected internal file driver type for default")
}

func (s *ManagerTestSuite) TestDriver_ResolveFileDriver_OverriddenByConfig() {

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()

	s.mockConfig.On("GetString", "session.drivers.file.driver").Return("custom").Once()
	s.mockConfig.On("Get", "session.drivers.file.via").Return(s.mockOtherFactory).Twice()
	s.mockConfig.On("Get", "session.drivers", mock.AnythingOfType("map[string]interface {}")).Return(
		map[string]any{"file": map[string]any{
			"driver": "custom",
			"via":    s.mockOtherFactory,
		}},
	).Once()

	manager := NewManager(s.mockConfig, s.json)
	s.Require().NotNil(manager)

	driverInstance, err := manager.Driver("file")
	s.Nil(err)
	s.NotNil(driverInstance)

	fmt.Println("Type of driverInstance:", fmt.Sprintf("%T", driverInstance))

	//s.Equal(s.mockOtherDriver, driverInstance, "Expected mock driver due to config override")
	s.IsType(&mocksession.Driver{}, driverInstance)
}

func (s *ManagerTestSuite) TestDriver_ResolveFileDriver_OverriddenByExtend() {

	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()

	err := s.manager.Extend("file", func() contractssession.Driver {

		return s.mockOtherDriver
	})
	s.Nil(err)

	driverInstance, err := s.manager.Driver("file")
	s.Nil(err)
	s.NotNil(driverInstance)

	s.Equal(s.mockOtherDriver, driverInstance, "Expected mock driver due to Extend override")
	s.IsType(&mocksession.Driver{}, driverInstance)
}

func (s *ManagerTestSuite) TestDriver_ResolveCustomDriver_FromConfig() {

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "session.drivers.my_driver.driver").Return("custom").Once()
	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()

	s.mockConfig.On("Get", "session.drivers.my_driver.via").Return(s.mockOtherFactory).Twice()

	s.mockConfig.On("Get", "session.drivers", mock.AnythingOfType("map[string]interface {}")).Return(
		map[string]any{"my_driver": map[string]any{
			"driver": "custom",
			"via":    s.mockOtherFactory,
		}},
	).Once()

	s.manager = NewManager(s.mockConfig, s.json)
	s.Require().NotNil(s.manager)

	driverInstance, err := s.manager.Driver("my_driver")
	s.Nil(err)
	s.NotNil(driverInstance)
	s.Equal(s.mockOtherDriver, driverInstance, "Expected mock driver for my_driver driver from config")
}

func (s *ManagerTestSuite) TestDriver_ResolveCustomDriver_FromExtend() {

	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()

	err := s.manager.Extend("extended_custom", func() contractssession.Driver {

		d, _ := NewCustomDriverFactory()
		return d
	})
	s.Nil(err)

	driverInstance, err := s.manager.Driver("extended_custom")
	s.Nil(err)
	s.NotNil(driverInstance)
	s.IsType(&CustomDriver{}, driverInstance)
}

func (s *ManagerTestSuite) TestDriver_NotSupported() {
	mockConfig := mockconfig.NewConfig(s.T())
	mockConfig.On("GetString", "session.driver").Return("not_supported").Twice()
	mockConfig.On("Get", "session.drivers", mock.AnythingOfType("map[string]interface {}")).Return(
		map[string]any{},
	).Once()

	manager := NewManager(mockConfig, s.json)
	s.Require().NotNil(manager)

	driverInstance, err := manager.Driver(mockConfig.GetString("session.driver"))
	s.ErrorIs(err, errors.SessionDriverNotSupported.Args("not_supported"))
	s.Nil(driverInstance)
}

func (s *ManagerTestSuite) TestDriver_NotSet() {
	mockConfig := mockconfig.NewConfig(s.T())
	mockConfig.On("GetString", "session.driver").Return("").Once()
	mockConfig.On("Get", "session.drivers", mock.AnythingOfType("map[string]interface {}")).Return(
		map[string]any{},
	).Once()

	manager := NewManager(mockConfig, s.json)
	s.Require().NotNil(manager)

	driverInstance, err := manager.Driver()
	s.ErrorIs(err, errors.SessionDriverIsNotSet)
	s.Nil(driverInstance)
}

func (s *ManagerTestSuite) TestExtend() {

	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()

	err := s.manager.Extend("test", func() contractssession.Driver {
		d, _ := NewCustomDriverFactory()
		return d
	})
	s.Nil(err)

	driverInstance, err := s.manager.Driver("test")
	s.Nil(err)
	s.NotNil(driverInstance)
	s.IsType(&CustomDriver{}, driverInstance)
}

func (s *ManagerTestSuite) TestExtend_AlreadyExists() {

	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()

	err1 := s.manager.Extend("test", func() contractssession.Driver { d, _ := NewCustomDriverFactory(); return d })
	s.Nil(err1)

	err2 := s.manager.Extend("test", func() contractssession.Driver { d, _ := NewCustomDriverFactory(); return d })
	s.ErrorIs(err2, errors.SessionDriverAlreadyExists.Args("test"))

	instanceExists, errResolve := s.manager.Driver("test")
	s.Nil(errResolve)

	s.IsType(&CustomDriver{}, instanceExists)
}

func (s *ManagerTestSuite) TestBuildSession_WithInternalFileDriver() {

	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()
	s.mockConfig.On("GetInt", "session.lifetime").Return(120).Once()
	s.mockConfig.On("GetString", "session.files").Return("storage/framework/sessions_test_manager").Once()
	s.mockConfig.On("GetString", "session.cookie").Return("goravel_test_session").Maybe()

	driverInstance, err := s.manager.Driver()
	s.Nil(err)
	s.Require().NotNil(driverInstance)
	s.IsType(&driver.File{}, driverInstance)

	session, err := s.manager.BuildSession(driverInstance)
	s.Nil(err)
	s.Require().NotNil(session)

	session.Put("data", "value_internal")
	s.Equal("value_internal", session.Get("data"))

	s.manager.ReleaseSession(session)
	s.Empty(session.All())
}

func (s *ManagerTestSuite) TestBuildSession_WithMockDriver() {

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("Get", "session.drivers", mock.AnythingOfType("map[string]interface {}")).Return(
		map[string]any{"mockdrv": map[string]any{
			"driver": "custom",
			"via":    s.mockOtherFactory,
		}},
	).Once()
	s.mockConfig.On("GetString", "session.driver").Return("mockdrv").Once()
	s.mockConfig.On("GetString", "session.drivers.mockdrv.driver").Return("custom").Once()
	s.mockConfig.On("Get", "session.drivers.mockdrv.via").Return(s.mockOtherFactory).Twice()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()
	s.mockConfig.On("GetString", "session.cookie").Return("mock_cookie").Once()

	s.manager = NewManager(s.mockConfig, s.json)

	driverInstance, err := s.manager.Driver("mockdrv")
	s.Nil(err)
	s.Require().NotNil(driverInstance)
	s.Equal(s.mockOtherDriver, driverInstance)

	session, err := s.manager.BuildSession(driverInstance)
	s.Nil(err)
	s.Require().NotNil(session)

	session.Put("data", "value_mock")
	s.Equal("mock_cookie", session.GetName())
	s.Equal("value_mock", session.Get("data"))

	s.manager.ReleaseSession(session)
	s.Empty(session.All())
}

func (s *ManagerTestSuite) TestBuildSession_NilDriver() {
	session, err := s.manager.BuildSession(nil)
	s.ErrorIs(err, errors.SessionDriverIsNotSet)
	s.Nil(session)
}

func (s *ManagerTestSuite) TestGetDefaultDriver() {

	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.Equal("file", s.manager.getDefaultDriver())
}

func BenchmarkSession_ManagerInteraction(b *testing.B) {
	s := new(ManagerTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()

	s.mockConfig.On("GetString", "session.driver").Return("file")
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30)
	s.mockConfig.On("GetString", "session.cookie").Return("bench_cookie")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		driverInstance, err := s.manager.Driver()
		if err != nil {
			b.Fatalf("Driver() failed: %v", err)
		}
		if driverInstance == nil {
			b.Fatal("Driver() returned nil")
		}

		session, err := s.manager.BuildSession(driverInstance)
		if err != nil {
			b.Fatalf("BuildSession() failed: %v", err)
		}
		if session == nil {
			b.Fatal("BuildSession() returned nil")
		}

		s.manager.ReleaseSession(session)

	}
	b.StopTimer()
}
