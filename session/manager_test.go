package session

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sessioncontract "github.com/goravel/framework/contracts/session"
	mockconfig "github.com/goravel/framework/mocks/config"
)

type ManagerTestSuite struct {
	suite.Suite
	mockConfig *mockconfig.Config
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}

func (m *ManagerTestSuite) SetupTest() {
	m.mockConfig = mockconfig.NewConfig(m.T())
}

func (m *ManagerTestSuite) TestDriver() {
	manager := NewManager(m.mockConfig)
	m.mockConfig.On("GetInt", "session.lifetime").Once().Return(120)
	m.mockConfig.On("GetString", "session.files").Once().Return("storage/framework/sessions")
	// provide driver name
	driver, err := manager.Driver("file")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*driver.FileDriver", fmt.Sprintf("%T", driver))

	// provide no driver name
	m.mockConfig.On("GetString", "session.driver").Once().Return("file")
	m.mockConfig.On("GetInt", "session.lifetime").Once().Return(120)
	m.mockConfig.On("GetString", "session.files").Once().Return("storage/framework/sessions")

	driver, err = manager.Driver()
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*driver.FileDriver", fmt.Sprintf("%T", driver))

	// provide custom driver
	manager.Extend("test", func() sessioncontract.Driver {
		return NewCustomDriver()
	})
	driver, err = manager.Driver("test")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))
}

func (m *ManagerTestSuite) TestExtend() {
	manager := NewManager(m.mockConfig)
	manager.Extend("test", func() sessioncontract.Driver {
		return NewCustomDriver()
	})
	driver, err := manager.Driver("test")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))
}

func (m *ManagerTestSuite) TestBuildSession() {
	manager := NewManager(m.mockConfig)
	m.mockConfig.On("GetString", "session.cookie").Once().Return("test_cookie")
	session := manager.BuildSession(nil)
	m.NotNil(session)
	m.Equal("test_cookie", session.GetName())
}

func (m *ManagerTestSuite) TestStore() {
	manager := NewManager(m.mockConfig)
	m.mockConfig.On("GetString", "session.driver").Once().Return("not_supported")
	session := manager.Store()
	m.Nil(session)

	manager.Extend("test", func() sessioncontract.Driver {
		return NewCustomDriver()
	})
	m.mockConfig.On("GetString", "session.driver").Once().Return("test")
	m.mockConfig.On("GetString", "session.cookie").Once().Return("test_cookie")
	session = manager.Store()
	m.NotNil(session)
	m.Equal("test_cookie", session.GetName())
}

func (m *ManagerTestSuite) TestGetDefaultDriver() {
	manager := NewManager(m.mockConfig)
	m.mockConfig.On("GetString", "session.driver").Return("file")
	m.Equal("file", manager.getDefaultDriver())
}

func (m *ManagerTestSuite) TestCreatDriver() {
	manager := NewManager(m.mockConfig)

	// custom driver
	manager.Extend("test", func() sessioncontract.Driver {
		return NewCustomDriver()
	})
	driver, err := manager.creatDriver("test")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver()))

	// built-in driver
	driver, err = manager.creatDriver("file")
	m.mockConfig.On("GetInt", "session.lifetime").Return(120)
	m.mockConfig.On("GetString", "session.files").Return("storage/framework/sessions")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*driver.FileDriver", fmt.Sprintf("%T", driver()))

	// not supported a driver
	driver, err = manager.creatDriver("not_supported")
	m.NotNil(err)
	m.Nil(driver)
}

type CustomDriver struct{}

func NewCustomDriver() sessioncontract.Driver {
	return &CustomDriver{}
}

func (c *CustomDriver) Close() bool {
	return true
}

func (c *CustomDriver) Destroy(string) error {
	return nil
}

func (c *CustomDriver) Gc(int) int {
	return 0
}

func (c *CustomDriver) Get(string) string {
	return ""
}

func (c *CustomDriver) Open(string, string) bool {
	return true
}

func (c *CustomDriver) Read(string) string {
	return ""
}

func (c *CustomDriver) Write(string, string) error {
	return nil
}
