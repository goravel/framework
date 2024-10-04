package session

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mockconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/str"
)

type ManagerTestSuite struct {
	suite.Suite
	mockConfig *mockconfig.Config
	manager    *Manager
	json       foundation.Json
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}

func (s *ManagerTestSuite) SetupTest() {
	s.mockConfig = mockconfig.NewConfig(s.T())
	s.mockConfig.On("GetInt", "session.lifetime").Return(120).Once()
	s.mockConfig.On("GetInt", "session.gc_interval", 30).Return(30).Once()
	s.mockConfig.On("GetString", "session.files").Return("storage/framework/sessions").Once()
	s.manager = s.getManager()
	s.json = json.NewJson()
}

func (s *ManagerTestSuite) TearDownSuite() {
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ManagerTestSuite) TestDriver() {
	// provide driver name
	driver, err := s.manager.Driver("file")
	s.Nil(err)
	s.NotNil(driver)
	s.Equal("*driver.File", fmt.Sprintf("%T", driver))

	// provide no driver name
	s.mockConfig.On("GetString", "session.driver").Return("file").Once()

	driver, err = s.manager.Driver()
	s.Nil(err)
	s.NotNil(driver)
	s.Equal("*driver.File", fmt.Sprintf("%T", driver))

	// provide custom driver
	s.mockConfig.On("GetInt", "session.gc_interval", 30).Return(30).Once()
	err = s.manager.Extend("test", NewCustomDriver)
	s.Nil(err)
	driver, err = s.manager.Driver("test")
	s.Nil(err)
	s.NotNil(driver)
	s.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))

	// not supported a driver
	s.mockConfig.On("GetString", "session.driver").Return("not_supported").Once()
	driver, err = s.manager.Driver()
	s.NotNil(err)
	s.ErrorIs(err, errors.ErrSessionDriverNotSupported)
	s.Equal(errors.ErrSessionDriverNotSupported.Args("not_supported").Error(), err.Error())
	s.Nil(driver)

	// driver is not set
	s.mockConfig.On("GetString", "session.driver").Return("").Once()
	driver, err = s.manager.Driver()
	s.NotNil(err)
	s.ErrorIs(err, errors.ErrSessionDriverIsNotSet)
	s.Nil(driver)
}

func (s *ManagerTestSuite) TestExtend() {
	s.mockConfig.On("GetInt", "session.gc_interval", 30).Return(30).Once()
	err := s.manager.Extend("test", NewCustomDriver)
	s.Nil(err)
	driver, err := s.manager.Driver("test")
	s.Nil(err)
	s.NotNil(driver)
	s.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))

	// driver already exists
	err = s.manager.Extend("test", NewCustomDriver)
	s.ErrorIs(err, errors.ErrSessionDriverAlreadyExists)
	s.EqualError(err, errors.ErrSessionDriverAlreadyExists.Args("test").Error())
}

func (s *ManagerTestSuite) TestBuildSession() {
	driver, err := s.manager.Driver("file")
	s.Nil(err)
	s.NotNil(driver)
	s.Equal("*driver.File", fmt.Sprintf("%T", driver))

	s.mockConfig.On("GetString", "session.cookie").Return("test_cookie").Once()
	session, err := s.manager.BuildSession(driver)
	s.Nil(err)
	s.NotNil(session)

	session.Put("name", "goravel")

	s.Equal("test_cookie", session.GetName())
	s.Equal("goravel", session.Get("name"))

	s.manager.ReleaseSession(session)
	s.Empty(session.GetName())
	s.Empty(session.All())

	// driver is nil
	session, err = s.manager.BuildSession(nil)
	s.ErrorIs(err, errors.ErrSessionDriverIsNotSet)
	s.Nil(session)
}

func (s *ManagerTestSuite) TestGetDefaultDriver() {
	s.mockConfig.On("GetString", "session.driver").Return("file")
	s.Equal("file", s.manager.getDefaultDriver())
}

func (s *ManagerTestSuite) getManager() *Manager {
	return NewManager(s.mockConfig, s.json)
}

func BenchmarkSession_ReadWrite(b *testing.B) {
	s := new(ManagerTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()

	// provide driver name
	driver, err := s.manager.Driver("file")
	s.Nil(err)
	s.NotNil(driver)
	s.Equal("*driver.File", fmt.Sprintf("%T", driver))

	// provide no driver name
	s.mockConfig.On("GetString", "session.driver").Return("file").Once()

	driver, err = s.manager.Driver()
	s.Nil(err)
	s.NotNil(driver)
	s.Equal("*driver.File", fmt.Sprintf("%T", driver))

	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := str.Random(32)
		s1 := str.Random(32)
		s.Nil(driver.Write(id, s1))
		data, err := driver.Read(id)
		s.Nil(err)
		s.Equal(s, data)
	}
	b.StopTimer()

	s.Nil(driver.Destroy("test"))
	s.Nil(os.RemoveAll("storage"))

}

type CustomDriver struct{}

func NewCustomDriver() sessioncontract.Driver {
	return &CustomDriver{}
}

func (c *CustomDriver) Close() error {
	return nil
}

func (c *CustomDriver) Destroy(string) error {
	return nil
}

func (c *CustomDriver) Gc(int) error {
	return nil
}

func (c *CustomDriver) Get(string) string {
	return ""
}

func (c *CustomDriver) Open(string, string) error {
	return nil
}

func (c *CustomDriver) Read(string) (string, error) {
	return "", nil
}

func (c *CustomDriver) Write(string, string) error {
	return nil
}
