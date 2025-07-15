package session

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	contractssession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mockconfig "github.com/goravel/framework/mocks/config"
	mocksession "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/session/driver"
	"github.com/goravel/framework/support/path"
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
	s.json = json.New()
	s.mockOtherDriver = new(mocksession.Driver)
	s.mockOtherFactory = MockDriverFactory(s.mockOtherDriver)

	s.mockConfig = mockconfig.NewConfig(s.T())
	s.mockConfig.EXPECT().GetString("session.default", "file").Return("file").Once()
	s.mockConfig.EXPECT().GetString("session.drivers.file.driver").Return("file").Once()
	s.mockConfig.EXPECT().GetInt("session.lifetime", 120).Return(120).Once()
	s.mockConfig.EXPECT().GetInt("session.gc_interval", 30).Return(30).Once()
	s.mockConfig.EXPECT().GetString("session.files").Return(path.Storage("framework/sessions")).Once()
	s.mockConfig.EXPECT().GetString("session.cookie").Return("goravel_session").Once()

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
	driverInstance, err := s.manager.Driver("file")
	s.Nil(err)
	s.NotNil(driverInstance)

	s.IsType(&driver.File{}, driverInstance, "Expected internal file driver type")
}

func (s *ManagerTestSuite) TestDriver_ResolveDefaultDriver_InternalFile() {
	driverInstance, err := s.manager.Driver()
	s.Nil(err)
	s.NotNil(driverInstance)
	s.IsType(&driver.File{}, driverInstance, "Expected internal file driver type for default")
}

func (s *ManagerTestSuite) TestDriver_ResolveCustomDriver() {
	s.mockConfig.On("GetString", "session.drivers.my_driver.driver").Return("custom").Once()
	s.mockConfig.On("Get", "session.drivers.my_driver.via").Return(s.mockOtherFactory).Once()

	driverInstance, err := s.manager.Driver("my_driver")
	s.Nil(err)
	s.NotNil(driverInstance)
	s.Equal(s.mockOtherDriver, driverInstance, "Expected mock driver for my_driver driver from config")
}

func (s *ManagerTestSuite) TestDriver_NotSupported() {
	s.mockConfig.On("GetString", "session.drivers.not_supported.driver").Return("not_supported").Once()

	driverInstance, err := s.manager.Driver("not_supported")
	s.ErrorIs(err, errors.SessionDriverNotSupported.Args("not_supported"))
	s.Nil(driverInstance)
}

func (s *ManagerTestSuite) TestBuildSession_WithInternalFileDriver() {
	driverInstance, err := s.manager.Driver("file")
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
	s.mockConfig.EXPECT().GetString("session.drivers.mock.driver").Return("custom").Once()
	s.mockConfig.EXPECT().Get("session.drivers.mock.via").Return(s.mockOtherFactory).Once()

	driverInstance, err := s.manager.Driver("mock")
	s.Nil(err)
	s.Require().NotNil(driverInstance)
	s.Equal(s.mockOtherDriver, driverInstance)

	session, err := s.manager.BuildSession(driverInstance)
	s.Nil(err)
	s.Require().NotNil(session)

	session.Put("data", "value_mock")
	s.Equal("goravel_session", session.GetName())
	s.Equal("value_mock", session.Get("data"))

	s.manager.ReleaseSession(session)
	s.Empty(session.All())
}

func (s *ManagerTestSuite) TestBuildSession_NilDriver() {
	session, err := s.manager.BuildSession(nil)
	s.ErrorIs(err, errors.SessionDriverIsNotSet)
	s.Nil(session)
}

func BenchmarkSession_ManagerInteraction(b *testing.B) {
	s := new(ManagerTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()

	s.mockConfig.On("GetString", "session.default").Return("file")
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
