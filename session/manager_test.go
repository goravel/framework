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
	manager := m.getManager()
	m.mockConfig.On("GetInt", "session.lifetime").Return(120).Once()
	m.mockConfig.On("GetString", "session.files").Return("storage/framework/sessions").Once()

	// provide driver name
	driver, err := manager.Driver("file")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*driver.File", fmt.Sprintf("%T", driver))

	// provide no driver name
	m.mockConfig.On("GetString", "session.driver").Return("file").Once()
	m.mockConfig.On("GetInt", "session.lifetime").Return(120).Once()
	m.mockConfig.On("GetString", "session.files").Return("storage/framework/sessions").Once()

	driver, err = manager.Driver()
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*driver.File", fmt.Sprintf("%T", driver))

	// provide custom driver
	manager.Extend("test", func() sessioncontract.Driver {
		return NewCustomDriver()
	})
	driver, err = manager.Driver("test")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))

	// not supported a driver
	m.mockConfig.On("GetString", "session.driver").Return("not_supported").Once()
	driver, err = manager.Driver()
	m.NotNil(err)
	m.Equal("driver [not_supported] not supported", err.Error())
	m.Nil(driver)
}

func (m *ManagerTestSuite) TestExtend() {
	manager := m.getManager()
	manager.Extend("test", func() sessioncontract.Driver {
		return NewCustomDriver()
	})
	driver, err := manager.Driver("test")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))
}

func (m *ManagerTestSuite) TestBuildSession() {
	manager := m.getManager()
	m.mockConfig.On("GetString", "session.cookie").Return("test_cookie").Once()
	session := manager.BuildSession(nil)
	m.NotNil(session)
	m.Equal("test_cookie", session.GetName())
}

func (m *ManagerTestSuite) TestGetDefaultDriver() {
	manager := m.getManager()
	m.mockConfig.On("GetString", "session.driver").Return("file")
	m.Equal("file", manager.getDefaultDriver())
}

func (m *ManagerTestSuite) getManager() *Manager {
	return NewManager(m.mockConfig)
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