package session

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mockconfig "github.com/goravel/framework/mocks/config"
)

type MockSessionDriver struct {
	mock.Mock
}

func (m *MockSessionDriver) Close() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockSessionDriver) Destroy(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockSessionDriver) Gc(maxLifetime int) error {
	args := m.Called(maxLifetime)
	return args.Error(0)
}
func (m *MockSessionDriver) Open(path string, name string) error {
	args := m.Called(path, name)
	return args.Error(0)
}
func (m *MockSessionDriver) Read(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}
func (m *MockSessionDriver) Write(id string, data string) error {
	args := m.Called(id, data)
	return args.Error(0)
}

func MockDriverFactory(mockDriverInstance *MockSessionDriver) func() (sessioncontract.Driver, error) {
	return func() (sessioncontract.Driver, error) {
		if mockDriverInstance == nil {
			return nil, fmt.Errorf("mock driver instance not provided to factory")
		}
		return mockDriverInstance, nil
	}
}

type CustomDriver struct{}

func NewCustomDriver() (sessioncontract.Driver, error) {
	return &CustomDriver{}, nil
}
func (c *CustomDriver) Close() error                { return nil }
func (c *CustomDriver) Destroy(string) error        { return nil }
func (c *CustomDriver) Gc(int) error                { return nil }
func (c *CustomDriver) Open(string, string) error   { return nil }
func (c *CustomDriver) Read(string) (string, error) { return "", nil }
func (c *CustomDriver) Write(string, string) error  { return nil }

type ManagerTestSuite struct {
	suite.Suite
	mockConfig        *mockconfig.Config
	manager           *Manager
	json              foundation.Json
	mockFileDriver    *MockSessionDriver
	mockFileDriverVia func() (sessioncontract.Driver, error)
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}

func (s *ManagerTestSuite) SetupTest() {
	s.mockConfig = mockconfig.NewConfig(s.T())
	s.json = json.NewJson()
	s.mockFileDriver = new(MockSessionDriver)
	s.mockFileDriverVia = MockDriverFactory(s.mockFileDriver)

	s.mockConfig.On("Get", "session.drivers", mock.AnythingOfType("map[string]interface {}")).Return(

		map[string]any{
			"file": map[string]any{
				"via": s.mockFileDriverVia,
			},
		},
	).Maybe()

	s.manager = NewManager(s.mockConfig, s.json)
	s.Require().NotNil(s.manager)

}

func (s *ManagerTestSuite) TearDownTest() {
	s.mockConfig.AssertExpectations(s.T())

}

func (s *ManagerTestSuite) TestDriver_ResolveConfiguredFileDriver() {

	s.mockConfig.On("GetString", "session.driver").Return("file").Maybe()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Maybe()

	driver, err := s.manager.Driver("file")
	s.Nil(err)
	s.NotNil(driver)

	s.Equal(s.mockFileDriver, driver)
}

func (s *ManagerTestSuite) TestDriver_ResolveDefaultDriver() {

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "session.driver").Return("file").Once()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Once()

	driver, err := s.manager.Driver()
	s.Nil(err)
	s.NotNil(driver)
	s.Equal(s.mockFileDriver, driver)
}

func (s *ManagerTestSuite) TestDriver_ResolveExtendedDriver() {

	s.mockConfig.On("GetString", "session.driver").Return("file").Maybe()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Maybe()

	err := s.manager.Extend("test", func() sessioncontract.Driver {
		d, _ := NewCustomDriver()
		return d
	})
	s.Nil(err)

	driver, err := s.manager.Driver("test")
	s.Nil(err)
	s.NotNil(driver)
	s.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))
}

func (s *ManagerTestSuite) TestDriver_NotSupported() {

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "session.driver").Return("not_supported")

	driver, err := s.manager.Driver()
	s.NotNil(err)
	s.ErrorIs(err, errors.SessionDriverNotSupported)
	s.Equal(errors.SessionDriverNotSupported.Args("not_supported").Error(), err.Error())
	s.Nil(driver)
}

func (s *ManagerTestSuite) TestDriver_NotSet() {

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "session.driver").Return("").Once()

	driver, err := s.manager.Driver()
	s.NotNil(err)
	s.ErrorIs(err, errors.SessionDriverIsNotSet)
	s.Nil(driver)
}

func (s *ManagerTestSuite) TestExtend() {

	s.mockConfig.On("GetString", "session.driver").Return("file").Maybe()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Maybe()

	err := s.manager.Extend("test", func() sessioncontract.Driver {
		d, _ := NewCustomDriver()
		return d
	})
	s.Nil(err)

	driver, err := s.manager.Driver("test")
	s.Nil(err)
	s.NotNil(driver)
	s.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))
}

func (s *ManagerTestSuite) TestExtend_AlreadyExists() {

	err1 := s.manager.Extend("test", func() sessioncontract.Driver {
		d, _ := NewCustomDriver()
		return d
	})
	s.Nil(err1)

	err2 := s.manager.Extend("test", func() sessioncontract.Driver {
		d, _ := NewCustomDriver()
		return d
	})
	s.NotNil(err2)
	s.ErrorIs(err2, errors.SessionDriverAlreadyExists)
	s.EqualError(err2, errors.SessionDriverAlreadyExists.Args("test").Error())
}

func (s *ManagerTestSuite) TestBuildSession() {

	s.mockConfig.On("GetString", "session.cookie").Return("test_cookie").Once()
	s.mockConfig.On("GetString", "session.driver").Return("file").Maybe()
	s.mockConfig.On("GetInt", "session.gc_interval").Return(30).Maybe()

	driver, err := s.manager.Driver(s.manager.getDefaultDriver())
	s.Nil(err)
	s.Require().NotNil(driver)
	s.Equal(s.mockFileDriver, driver)

	session, err := s.manager.BuildSession(driver)
	s.Nil(err)
	s.Require().NotNil(session)

	session.Put("name", "goravel")
	s.Equal("test_cookie", session.GetName())
	s.Equal("goravel", session.Get("name"))
	s.NotEmpty(session.GetID(), "Session ID should be generated or set")

	s.manager.ReleaseSession(session)
	s.Empty(session.GetName(), "Session name should be empty after release")
	s.Empty(session.All(), "Session attributes should be empty after release")
}

func (s *ManagerTestSuite) TestBuildSession_NilDriver() {

	session, err := s.manager.BuildSession(nil)
	s.ErrorIs(err, errors.SessionDriverIsNotSet)
	s.Nil(session)
}

func (s *ManagerTestSuite) TestGetDefaultDriver() {

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "session.driver").Return("custom_default").Once()

	s.Equal("custom_default", s.manager.getDefaultDriver())
}

func BenchmarkSession_ManagerInteraction(b *testing.B) {
	s := new(ManagerTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()

	s.mockConfig.On("GetString", "session.driver", "file").Return("file")
	s.mockConfig.On("GetInt", "session.gc_interval", 30).Return(30)
	s.mockConfig.On("GetString", "session.cookie").Return("bench_cookie")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		driver, err := s.manager.Driver()
		if err != nil {
			b.Fatalf("Driver() failed during benchmark: %v", err)
		}
		if driver == nil {
			b.Fatal("Driver() returned nil driver during benchmark")
		}

		session, err := s.manager.BuildSession(driver)
		if err != nil {
			b.Fatalf("BuildSession() failed during benchmark: %v", err)
		}
		if session == nil {
			b.Fatal("BuildSession() returned nil session during benchmark")
		}

		s.manager.ReleaseSession(session)

	}
	b.StopTimer()
}
